package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DefaultLockTimeout is the uniform 60-second deadline governing lock acquisition (D8, ADR 0014 §4).
const DefaultLockTimeout = 60 * time.Second

var (
	// ErrIndexLocked indicates that concurrency timeout expired while waiting for index locks.
	ErrIndexLocked = errors.New("index_locked: lock acquisition timed out")

	// ErrLockUnavailable indicates that platform locking primitives failed or are unsupported
	// on the underlying filesystem (e.g. remote FS without POSIX/Win32 locking, permission errors).
	ErrLockUnavailable = errors.New("lock_unavailable: file locking unavailable or failed")

	// ErrLockOrder indicates an attempt to acquire locks out of the strict global order M -> S -> P.
	ErrLockOrder = errors.New("lock_order_violation: locks must be acquired in strict M -> S -> P order")
)

// LockKind identifies the target resource in the M -> S -> P acquisition hierarchy.
type LockKind string

const (
	LockKindManifest LockKind = "M"
	LockKindStore    LockKind = "S"
	LockKindSnapshot LockKind = "P"
)

// LockLevel assigns a numerical precedence to enforce M (1) -> S (2) -> P (3).
type LockLevel int

const (
	LockLevelNone LockLevel = iota
	LockLevelManifest
	LockLevelStore
	LockLevelSnapshot
)

// LockError provides structured error reporting while conforming to ErrIndexLocked / ErrLockUnavailable.
type LockError struct {
	Code   string   // "index_locked", "lock_unavailable", "lock_order_violation"
	Target string   // file path or resource name
	Kind   LockKind // M, S, or P
	Err    error    // underlying error
}

func (e *LockError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s lock on %s: %v", e.Code, e.Kind, e.Target, e.Err)
	}
	return fmt.Sprintf("%s: %s lock on %s", e.Code, e.Kind, e.Target)
}

func (e *LockError) Unwrap() error {
	return e.Err
}

func (e *LockError) Is(target error) bool {
	switch target {
	case ErrIndexLocked:
		return e.Code == "index_locked"
	case ErrLockUnavailable:
		return e.Code == "lock_unavailable"
	case ErrLockOrder:
		return e.Code == "lock_order_violation"
	}
	return false
}

// LockTargets specifies which resources among M, S, P need to be acquired.
type LockTargets struct {
	// ManifestPath is the path to gira.yaml (M). Optional if no manifest is locked.
	ManifestPath string
	// StoreDir is the store root or lineage directory (S). Optional if no store is locked.
	StoreDir string
	// SnapshotPath is the destination path for detached snapshot or copy (P). Optional.
	SnapshotPath string
}

// LockOptions configures lock acquisition behavior, timeouts, backoff, and testing hooks.
type LockOptions struct {
	// Timeout controls how long to wait before returning ErrIndexLocked.
	// If 0 and ctx has no deadline, DefaultLockTimeout (60s) is applied.
	Timeout time.Duration

	// PollInterval controls initial polling backoff interval (default 5ms).
	PollInterval time.Duration

	// MaxPollInterval controls maximum polling backoff interval (default 50ms).
	MaxPollInterval time.Duration

	// InjectedLockError simulates a platform locking failure returning ErrLockUnavailable (AC5).
	InjectedLockError error

	// OnAcquired is an optional callback invoked immediately after a lock is acquired.
	OnAcquired func(kind LockKind, path string)

	// OnReleased is an optional callback invoked immediately after a lock is released.
	OnReleased func(kind LockKind, path string)
}

type LockOption func(*LockOptions)

// WithTimeout sets a custom lock timeout.
func WithTimeout(d time.Duration) LockOption {
	return func(o *LockOptions) { o.Timeout = d }
}

// WithInjectedError injects a simulated locking failure (returns ErrLockUnavailable).
func WithInjectedError(err error) LockOption {
	return func(o *LockOptions) { o.InjectedLockError = err }
}

// WithPollInterval sets initial and max polling backoff intervals.
func WithPollInterval(poll, maxPoll time.Duration) LockOption {
	return func(o *LockOptions) {
		o.PollInterval = poll
		o.MaxPollInterval = maxPoll
	}
}

// WithLockHooks registers callbacks for lock acquisition and release events.
func WithLockHooks(onAcquired, onReleased func(kind LockKind, path string)) LockOption {
	return func(o *LockOptions) {
		o.OnAcquired = onAcquired
		o.OnReleased = onReleased
	}
}

// FileLock represents an active OS-level file lock on a specific file path.
type FileLock struct {
	mu   sync.Mutex
	file *os.File
	path string
	kind LockKind
}

func (fl *FileLock) Path() string    { return fl.path }
func (fl *FileLock) Kind() LockKind  { return fl.kind }
func (fl *FileLock) File() *os.File  { return fl.file }

// Unlock releases the OS-level file lock and closes the file descriptor.
func (fl *FileLock) Unlock() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	if fl.file == nil {
		return nil
	}
	unlockErr := platformUnlockFile(fl.file)
	closeErr := fl.file.Close()
	fl.file = nil
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}

// HeldLocks tracks a collection of acquired locks and enforces reverse-order release (P -> S -> M).
type HeldLocks struct {
	mu           sync.Mutex
	locks        []*FileLock
	releaseOrder []LockKind
	opts         LockOptions
}

// Release releases all held locks in strictly reverse order (P -> S -> M).
// It is idempotent and thread-safe.
func (hl *HeldLocks) Release() error {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	if len(hl.locks) == 0 {
		return nil
	}
	var errs []error
	for i := len(hl.locks) - 1; i >= 0; i-- {
		fl := hl.locks[i]
		if err := fl.Unlock(); err != nil {
			errs = append(errs, err)
		}
		hl.releaseOrder = append(hl.releaseOrder, fl.Kind())
		if hl.opts.OnReleased != nil {
			hl.opts.OnReleased(fl.Kind(), fl.Path())
		}
	}
	hl.locks = nil
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Order returns the slice of LockKinds in the order they were acquired.
func (hl *HeldLocks) Order() []LockKind {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	res := make([]LockKind, len(hl.locks))
	for i, l := range hl.locks {
		res[i] = l.Kind()
	}
	return res
}

// ReleaseOrder returns the slice of LockKinds in the order they were released.
func (hl *HeldLocks) ReleaseOrder() []LockKind {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	res := make([]LockKind, len(hl.releaseOrder))
	copy(res, hl.releaseOrder)
	return res
}

// Locks returns a copy of the currently held FileLocks.
func (hl *HeldLocks) Locks() []*FileLock {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	res := make([]*FileLock, len(hl.locks))
	copy(res, hl.locks)
	return res
}

// ResolveManifestLockPath returns the canonical lock file path for a manifest.
// If the manifest exists on disk, physical file identity is queried so that relative paths,
// absolute paths, and symlinks point to the same lock file in .gira/locks/.
// If the manifest does not exist yet (e.g. before gira init), the canonical parent directory
// is used with a pending lock file.
func ResolveManifestLockPath(manifestPath string) (string, error) {
	if strings.TrimSpace(manifestPath) == "" {
		return "", fmt.Errorf("manifest path cannot be empty")
	}

	if fi, err := os.Stat(manifestPath); err == nil && !fi.IsDir() {
		realPath, evalErr := filepath.EvalSymlinks(manifestPath)
		if evalErr != nil {
			realPath = manifestPath
		}
		absPath, absErr := filepath.Abs(realPath)
		if absErr == nil {
			realPath = absPath
		}

		phyID, idErr := getPhysicalFileID(realPath)
		manifestDir := filepath.Dir(realPath)
		lockDir := filepath.Join(manifestDir, ".gira", "locks")
		if idErr == nil && phyID != "" {
			return filepath.Join(lockDir, fmt.Sprintf("m-%s.lock", phyID)), nil
		}
		return filepath.Join(lockDir, "manifest.lock"), nil
	}

	dir := filepath.Dir(manifestPath)
	realDir, evalErr := filepath.EvalSymlinks(dir)
	if evalErr != nil {
		realDir = dir
	}
	absDir, absErr := filepath.Abs(realDir)
	if absErr == nil {
		realDir = absDir
	}
	lockDir := filepath.Join(realDir, ".gira", "locks")
	return filepath.Join(lockDir, "manifest.lock"), nil
}

// ResolveStoreLockPath returns the canonical lock file path for a store directory.
func ResolveStoreLockPath(storeDir string) (string, error) {
	if strings.TrimSpace(storeDir) == "" {
		return "", fmt.Errorf("store directory cannot be empty")
	}
	realDir, evalErr := filepath.EvalSymlinks(storeDir)
	if evalErr != nil {
		realDir = storeDir
	}
	absDir, absErr := filepath.Abs(realDir)
	if absErr == nil {
		realDir = absDir
	}
	return filepath.Join(absDir, ".lock"), nil
}

// ResolveSnapshotLockPath returns the canonical lock file path for a snapshot file.
func ResolveSnapshotLockPath(snapshotPath string) (string, error) {
	if strings.TrimSpace(snapshotPath) == "" {
		return "", fmt.Errorf("snapshot path cannot be empty")
	}
	realPath := snapshotPath
	if fi, err := os.Stat(snapshotPath); err == nil && !fi.IsDir() {
		if rp, evalErr := filepath.EvalSymlinks(snapshotPath); evalErr == nil {
			realPath = rp
		}
	} else {
		parent := filepath.Dir(snapshotPath)
		if rp, evalErr := filepath.EvalSymlinks(parent); evalErr == nil {
			realPath = filepath.Join(rp, filepath.Base(snapshotPath))
		}
	}
	absPath, absErr := filepath.Abs(realPath)
	if absErr == nil {
		realPath = absPath
	}
	return realPath + ".lock", nil
}

// tryAcquireSingleFile attempts to create/open lockPath and acquire exclusive OS lock within ctx.
func tryAcquireSingleFile(ctx context.Context, lockPath string, kind LockKind, opts LockOptions) (*FileLock, error) {
	if opts.InjectedLockError != nil {
		return nil, &LockError{
			Code:   "lock_unavailable",
			Target: lockPath,
			Kind:   kind,
			Err:    fmt.Errorf("%w: %v", ErrLockUnavailable, opts.InjectedLockError),
		}
	}

	dir := filepath.Dir(lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, &LockError{
			Code:   "lock_unavailable",
			Target: lockPath,
			Kind:   kind,
			Err:    fmt.Errorf("%w: failed to create lock directory %s: %v", ErrLockUnavailable, dir, err),
		}
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, &LockError{
			Code:   "lock_unavailable",
			Target: lockPath,
			Kind:   kind,
			Err:    fmt.Errorf("%w: failed to open lock file %s: %v", ErrLockUnavailable, lockPath, err),
		}
	}

	pollInterval := opts.PollInterval
	if pollInterval <= 0 {
		pollInterval = 5 * time.Millisecond
	}
	maxPoll := opts.MaxPollInterval
	if maxPoll <= 0 {
		maxPoll = 50 * time.Millisecond
	}

	timer := time.NewTimer(0)
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	defer timer.Stop()

	for {
		lockErr := platformLockFile(f)
		if lockErr == nil {
			_ = f.Truncate(0)
			_, _ = f.Seek(0, 0)
			_, _ = fmt.Fprintf(f, "pid=%d\nkind=%s\n", os.Getpid(), kind)
			_ = f.Sync()

			return &FileLock{
				file: f,
				path: lockPath,
				kind: kind,
			}, nil
		}

		if isLockedError(lockErr) {
			select {
			case <-ctx.Done():
				_ = f.Close()
				return nil, &LockError{
					Code:   "index_locked",
					Target: lockPath,
					Kind:   kind,
					Err:    fmt.Errorf("%w: lock acquisition timed out on %s: %v", ErrIndexLocked, lockPath, ctx.Err()),
				}
			default:
			}

			timer.Reset(pollInterval)
			select {
			case <-ctx.Done():
				_ = f.Close()
				return nil, &LockError{
					Code:   "index_locked",
					Target: lockPath,
					Kind:   kind,
					Err:    fmt.Errorf("%w: lock acquisition timed out on %s: %v", ErrIndexLocked, lockPath, ctx.Err()),
				}
			case <-timer.C:
				pollInterval *= 2
				if pollInterval > maxPoll {
					pollInterval = maxPoll
				}
			}
			continue
		}

		_ = f.Close()
		return nil, &LockError{
			Code:   "lock_unavailable",
			Target: lockPath,
			Kind:   kind,
			Err:    fmt.Errorf("%w: locking failed on %s: %v", ErrLockUnavailable, lockPath, lockErr),
		}
	}
}

// AcquireLocks acquires locks in strictly ordered sequence M -> S -> P.
// If any lock fails, all previously acquired locks are released in reverse order before returning (fail-closed).
func AcquireLocks(ctx context.Context, targets LockTargets, optFns ...LockOption) (*HeldLocks, error) {
	opts := LockOptions{
		Timeout:         DefaultLockTimeout,
		PollInterval:    5 * time.Millisecond,
		MaxPollInterval: 50 * time.Millisecond,
	}
	for _, fn := range optFns {
		fn(&opts)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline && opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	if targets.ManifestPath == "" && targets.StoreDir == "" && targets.SnapshotPath == "" {
		return nil, fmt.Errorf("no lock targets specified")
	}

	held := &HeldLocks{opts: opts}

	failClosed := func(err error) (*HeldLocks, error) {
		_ = held.Release()
		return nil, err
	}

	// 1. M (Manifest lock)
	if targets.ManifestPath != "" {
		lockPath, err := ResolveManifestLockPath(targets.ManifestPath)
		if err != nil {
			return failClosed(&LockError{Code: "lock_unavailable", Target: targets.ManifestPath, Kind: LockKindManifest, Err: err})
		}
		fl, err := tryAcquireSingleFile(ctx, lockPath, LockKindManifest, opts)
		if err != nil {
			return failClosed(err)
		}
		held.locks = append(held.locks, fl)
		if opts.OnAcquired != nil {
			opts.OnAcquired(LockKindManifest, lockPath)
		}
	}

	// 2. S (Store lock)
	if targets.StoreDir != "" {
		lockPath, err := ResolveStoreLockPath(targets.StoreDir)
		if err != nil {
			return failClosed(&LockError{Code: "lock_unavailable", Target: targets.StoreDir, Kind: LockKindStore, Err: err})
		}
		fl, err := tryAcquireSingleFile(ctx, lockPath, LockKindStore, opts)
		if err != nil {
			return failClosed(err)
		}
		held.locks = append(held.locks, fl)
		if opts.OnAcquired != nil {
			opts.OnAcquired(LockKindStore, lockPath)
		}
	}

	// 3. P (Snapshot lock)
	if targets.SnapshotPath != "" {
		lockPath, err := ResolveSnapshotLockPath(targets.SnapshotPath)
		if err != nil {
			return failClosed(&LockError{Code: "lock_unavailable", Target: targets.SnapshotPath, Kind: LockKindSnapshot, Err: err})
		}
		fl, err := tryAcquireSingleFile(ctx, lockPath, LockKindSnapshot, opts)
		if err != nil {
			return failClosed(err)
		}
		held.locks = append(held.locks, fl)
		if opts.OnAcquired != nil {
			opts.OnAcquired(LockKindSnapshot, lockPath)
		}
	}

	return held, nil
}

// AcquireManifestLock acquires solely the M (Manifest) lock.
func AcquireManifestLock(ctx context.Context, manifestPath string, opts ...LockOption) (*HeldLocks, error) {
	return AcquireLocks(ctx, LockTargets{ManifestPath: manifestPath}, opts...)
}

// AcquireStoreLock acquires solely the S (Store) lock.
func AcquireStoreLock(ctx context.Context, storeDir string, opts ...LockOption) (*HeldLocks, error) {
	return AcquireLocks(ctx, LockTargets{StoreDir: storeDir}, opts...)
}

// AcquireSnapshotLock acquires solely the P (Snapshot) lock.
func AcquireSnapshotLock(ctx context.Context, snapshotPath string, opts ...LockOption) (*HeldLocks, error) {
	return AcquireLocks(ctx, LockTargets{SnapshotPath: snapshotPath}, opts...)
}

// LockFileDirect acquires an exclusive lock directly on the specified file path.
func LockFileDirect(ctx context.Context, path string, opts ...LockOption) (*FileLock, error) {
	options := LockOptions{
		Timeout:         DefaultLockTimeout,
		PollInterval:    5 * time.Millisecond,
		MaxPollInterval: 50 * time.Millisecond,
	}
	for _, fn := range opts {
		fn(&options)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline && options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	return tryAcquireSingleFile(ctx, path, "DIRECT", options)
}

// LockChain enables incremental acquisition of locks while strictly preventing out-of-order calls.
type LockChain struct {
	mu           sync.Mutex
	currentLevel LockLevel
	held         *HeldLocks
}

// NewLockChain creates a new LockChain enforcing M -> S -> P acquisition order.
func NewLockChain(opts ...LockOption) *LockChain {
	options := LockOptions{
		Timeout:         DefaultLockTimeout,
		PollInterval:    5 * time.Millisecond,
		MaxPollInterval: 50 * time.Millisecond,
	}
	for _, fn := range opts {
		fn(&options)
	}
	return &LockChain{
		currentLevel: LockLevelNone,
		held:         &HeldLocks{opts: options},
	}
}

// AcquireManifest acquires M lock. Fails if S or P was already acquired.
func (lc *LockChain) AcquireManifest(ctx context.Context, manifestPath string) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if lc.currentLevel >= LockLevelManifest {
		return &LockError{
			Code:   "lock_order_violation",
			Target: manifestPath,
			Kind:   LockKindManifest,
			Err:    fmt.Errorf("%w: cannot acquire M at level %d", ErrLockOrder, lc.currentLevel),
		}
	}
	lockPath, err := ResolveManifestLockPath(manifestPath)
	if err != nil {
		return &LockError{Code: "lock_unavailable", Target: manifestPath, Kind: LockKindManifest, Err: err}
	}
	fl, err := tryAcquireSingleFile(ctx, lockPath, LockKindManifest, lc.held.opts)
	if err != nil {
		return err
	}
	lc.held.locks = append(lc.held.locks, fl)
	lc.currentLevel = LockLevelManifest
	if lc.held.opts.OnAcquired != nil {
		lc.held.opts.OnAcquired(LockKindManifest, lockPath)
	}
	return nil
}

// AcquireStore acquires S lock. Fails if P was already acquired.
func (lc *LockChain) AcquireStore(ctx context.Context, storeDir string) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if lc.currentLevel >= LockLevelStore {
		return &LockError{
			Code:   "lock_order_violation",
			Target: storeDir,
			Kind:   LockKindStore,
			Err:    fmt.Errorf("%w: cannot acquire S at level %d", ErrLockOrder, lc.currentLevel),
		}
	}
	lockPath, err := ResolveStoreLockPath(storeDir)
	if err != nil {
		return &LockError{Code: "lock_unavailable", Target: storeDir, Kind: LockKindStore, Err: err}
	}
	fl, err := tryAcquireSingleFile(ctx, lockPath, LockKindStore, lc.held.opts)
	if err != nil {
		return err
	}
	lc.held.locks = append(lc.held.locks, fl)
	lc.currentLevel = LockLevelStore
	if lc.held.opts.OnAcquired != nil {
		lc.held.opts.OnAcquired(LockKindStore, lockPath)
	}
	return nil
}

// AcquireSnapshot acquires P lock. Fails if P was already acquired.
func (lc *LockChain) AcquireSnapshot(ctx context.Context, snapshotPath string) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if lc.currentLevel >= LockLevelSnapshot {
		return &LockError{
			Code:   "lock_order_violation",
			Target: snapshotPath,
			Kind:   LockKindSnapshot,
			Err:    fmt.Errorf("%w: cannot acquire P at level %d", ErrLockOrder, lc.currentLevel),
		}
	}
	lockPath, err := ResolveSnapshotLockPath(snapshotPath)
	if err != nil {
		return &LockError{Code: "lock_unavailable", Target: snapshotPath, Kind: LockKindSnapshot, Err: err}
	}
	fl, err := tryAcquireSingleFile(ctx, lockPath, LockKindSnapshot, lc.held.opts)
	if err != nil {
		return err
	}
	lc.held.locks = append(lc.held.locks, fl)
	lc.currentLevel = LockLevelSnapshot
	if lc.held.opts.OnAcquired != nil {
		lc.held.opts.OnAcquired(LockKindSnapshot, lockPath)
	}
	return nil
}

// Release releases all held locks in reverse order (P -> S -> M).
func (lc *LockChain) Release() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.currentLevel = LockLevelNone
	return lc.held.Release()
}

// Order returns the order in which locks were acquired.
func (lc *LockChain) Order() []LockKind {
	return lc.held.Order()
}

// ReleaseOrder returns the order in which locks were released.
func (lc *LockChain) ReleaseOrder() []LockKind {
	return lc.held.ReleaseOrder()
}

// WithLocks executes fn inside the M -> S -> P lock scope.
// If lock acquisition fails, no modification occurs (fail-closed).
func WithLocks(ctx context.Context, targets LockTargets, opts LockOptions, fn func() error) error {
	locks, err := AcquireLocks(ctx, targets, func(o *LockOptions) { *o = opts })
	if err != nil {
		return err
	}
	defer locks.Release()
	return fn()
}
