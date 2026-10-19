package store_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/store"
)

// Helper process entry point for subprocess tests (AC1, AC3).
func init() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		runHelperProcess()
		os.Exit(0)
	}
}

func runHelperProcess() {
	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "missing helper subcommand\n")
		os.Exit(2)
	}
	cmd := args[1]
	switch cmd {
	case "lock-and-wait":
		// Usage: helper -- lock-and-wait <manifestPath>
		if len(args) < 3 {
			os.Exit(2)
		}
		manifestPath := args[2]
		ctx := context.Background()
		held, err := store.AcquireManifestLock(ctx, manifestPath, store.WithTimeout(30*time.Second))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to acquire lock: %v\n", err)
			os.Exit(1)
		}
		defer held.Release()

		// Signal to parent that lock is held
		fmt.Println("LOCKED")
		os.Stdout.Sync()

		// Wait indefinitely until killed
		select {}

	case "contend-counter":
		// Usage: helper -- contend-counter <lockPath> <counterFile> <iterations>
		if len(args) < 5 {
			os.Exit(2)
		}
		lockPath := args[2]
		counterFile := args[3]
		iters, err := strconv.Atoi(args[4])
		if err != nil {
			os.Exit(2)
		}

		ctx := context.Background()
		for i := 0; i < iters; i++ {
			fl, err := store.LockFileDirect(ctx, lockPath, store.WithTimeout(30*time.Second))
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to lock in iteration %d: %v\n", i, err)
				os.Exit(1)
			}

			// Read counter
			val := 0
			data, err := os.ReadFile(counterFile)
			if err == nil && len(strings.TrimSpace(string(data))) > 0 {
				val, _ = strconv.Atoi(strings.TrimSpace(string(data)))
			}
			val++
			time.Sleep(2 * time.Millisecond) // critical section window
			if err := os.WriteFile(counterFile, []byte(strconv.Itoa(val)), 0666); err != nil {
				_ = fl.Unlock()
				os.Exit(1)
			}

			if err := fl.Unlock(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unlock: %v\n", err)
				os.Exit(1)
			}
		}
		os.Exit(0)
	}
}

// AC1: Dois processos concorrentes se excluem nos três SOs.
func TestAC1_ConcurrentProcessesMutualExclusion(t *testing.T) {
	tempDir := t.TempDir()
	lockPath := filepath.Join(tempDir, "shared.lock")
	counterFile := filepath.Join(tempDir, "counter.txt")
	if err := os.WriteFile(counterFile, []byte("0"), 0666); err != nil {
		t.Fatalf("failed to init counter file: %v", err)
	}

	const iterations = 15
	cmd1 := exec.Command(os.Args[0], "-test.run=TestAC1_ConcurrentProcessesMutualExclusion", "--", "contend-counter", lockPath, counterFile, strconv.Itoa(iterations))
	cmd1.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	cmd2 := exec.Command(os.Args[0], "-test.run=TestAC1_ConcurrentProcessesMutualExclusion", "--", "contend-counter", lockPath, counterFile, strconv.Itoa(iterations))
	cmd2.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	var out1, out2 strings.Builder
	cmd1.Stdout = &out1
	cmd1.Stderr = &out1
	cmd2.Stdout = &out2
	cmd2.Stderr = &out2

	if err := cmd1.Start(); err != nil {
		t.Fatalf("failed to start cmd1: %v", err)
	}
	if err := cmd2.Start(); err != nil {
		_ = cmd1.Process.Kill()
		t.Fatalf("failed to start cmd2: %v", err)
	}

	err1 := cmd1.Wait()
	err2 := cmd2.Wait()

	if err1 != nil {
		t.Fatalf("cmd1 failed: %v, output: %s", err1, out1.String())
	}
	if err2 != nil {
		t.Fatalf("cmd2 failed: %v, output: %s", err2, out2.String())
	}

	data, err := os.ReadFile(counterFile)
	if err != nil {
		t.Fatalf("failed to read counter file: %v", err)
	}
	finalCount, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatalf("invalid counter value: %q", string(data))
	}

	expected := iterations * 2
	if finalCount != expected {
		t.Fatalf("expected counter %d, got %d (mutual exclusion violated)", expected, finalCount)
	}
}

// AC1 (goroutines with independent file descriptors): Verifies mutual exclusion within same process.
func TestAC1_ConcurrentGoroutinesMutualExclusion(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte("target: test\n"), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	var activeInCritical int32
	var violations int32
	const numGoroutines = 8
	const opsPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func() {
			defer wg.Done()
			for op := 0; op < opsPerGoroutine; op++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				held, err := store.AcquireManifestLock(ctx, manifestPath, store.WithPollInterval(2*time.Millisecond, 15*time.Millisecond))
				cancel()
				if err != nil {
					t.Errorf("acquire failed: %v", err)
					return
				}

				curr := atomic.AddInt32(&activeInCritical, 1)
				if curr > 1 {
					atomic.AddInt32(&violations, 1)
				}
				time.Sleep(3 * time.Millisecond)
				atomic.AddInt32(&activeInCritical, -1)

				if err := held.Release(); err != nil {
					t.Errorf("release failed: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()
	if v := atomic.LoadInt32(&violations); v > 0 {
		t.Fatalf("detected %d mutual exclusion violations among goroutines", v)
	}
}

// AC2: Caminhos relativos, absolutos e symlink do mesmo manifesto compartilham lock M.
func TestAC2_ManifestLockIdentitySharing(t *testing.T) {
	tempDir := t.TempDir()
	manifestReal := filepath.Join(tempDir, "gira.yaml")
	if err := os.WriteFile(manifestReal, []byte("system: ac2-test\n"), 0644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Symlink to manifest in the same directory
	symlinkPath := filepath.Join(tempDir, "symlink-gira.yaml")
	if err := os.Symlink(manifestReal, symlinkPath); err != nil {
		t.Skipf("skipping symlink test on platform without symlink support: %v", err)
	}

	// Another symlink in a subdirectory pointing to manifest
	subDir := filepath.Join(tempDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subDir: %v", err)
	}
	subSymlink := filepath.Join(subDir, "link-to-manifest.yaml")
	if err := os.Symlink(manifestReal, subSymlink); err != nil {
		t.Fatalf("failed to create sub symlink: %v", err)
	}

	// Relative path
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	relPath, err := filepath.Rel(wd, manifestReal)
	if err != nil {
		relPath = manifestReal
	}

	// Verify all resolve to the same canonical lock file
	lockPathAbs, err := store.ResolveManifestLockPath(manifestReal)
	if err != nil {
		t.Fatalf("ResolveManifestLockPath abs failed: %v", err)
	}
	lockPathRel, err := store.ResolveManifestLockPath(relPath)
	if err != nil {
		t.Fatalf("ResolveManifestLockPath rel failed: %v", err)
	}
	lockPathSym, err := store.ResolveManifestLockPath(symlinkPath)
	if err != nil {
		t.Fatalf("ResolveManifestLockPath sym failed: %v", err)
	}
	lockPathSubSym, err := store.ResolveManifestLockPath(subSymlink)
	if err != nil {
		t.Fatalf("ResolveManifestLockPath sub sym failed: %v", err)
	}

	if lockPathAbs != lockPathRel {
		t.Fatalf("lock path mismatch abs vs rel: %s != %s", lockPathAbs, lockPathRel)
	}
	if lockPathAbs != lockPathSym {
		t.Fatalf("lock path mismatch abs vs symlink: %s != %s", lockPathAbs, lockPathSym)
	}
	if lockPathAbs != lockPathSubSym {
		t.Fatalf("lock path mismatch abs vs sub symlink: %s != %s", lockPathAbs, lockPathSubSym)
	}

	// Test 1: Acquire via relative path -> blocks absolute and symlink
	ctx := context.Background()
	heldRel, err := store.AcquireManifestLock(ctx, relPath)
	if err != nil {
		t.Fatalf("failed to acquire relative lock: %v", err)
	}

	// Attempt via absolute path with short timeout -> must return ErrIndexLocked
	shortCtx, cancelShort := context.WithTimeout(ctx, 60*time.Millisecond)
	_, errAbs := store.AcquireManifestLock(shortCtx, manifestReal, store.WithPollInterval(2*time.Millisecond, 10*time.Millisecond))
	cancelShort()
	if errAbs == nil || !errors.Is(errAbs, store.ErrIndexLocked) {
		t.Fatalf("expected ErrIndexLocked for absolute path, got: %v", errAbs)
	}
	if !strings.Contains(errAbs.Error(), "index_locked") {
		t.Fatalf("expected error string to contain 'index_locked', got: %v", errAbs)
	}

	// Attempt via symlink path with short timeout -> must return ErrIndexLocked
	shortCtx2, cancelShort2 := context.WithTimeout(ctx, 60*time.Millisecond)
	_, errSym := store.AcquireManifestLock(shortCtx2, symlinkPath, store.WithPollInterval(2*time.Millisecond, 10*time.Millisecond))
	cancelShort2()
	if errSym == nil || !errors.Is(errSym, store.ErrIndexLocked) {
		t.Fatalf("expected ErrIndexLocked for symlink path, got: %v", errSym)
	}

	// Release relative lock
	if err := heldRel.Release(); err != nil {
		t.Fatalf("failed to release rel lock: %v", err)
	}

	// Test 2: Acquire via symlink -> blocks relative and absolute
	heldSym, err := store.AcquireManifestLock(ctx, symlinkPath)
	if err != nil {
		t.Fatalf("failed to acquire symlink lock: %v", err)
	}

	shortCtx3, cancelShort3 := context.WithTimeout(ctx, 60*time.Millisecond)
	_, errRel2 := store.AcquireManifestLock(shortCtx3, relPath, store.WithPollInterval(2*time.Millisecond, 10*time.Millisecond))
	cancelShort3()
	if errRel2 == nil || !errors.Is(errRel2, store.ErrIndexLocked) {
		t.Fatalf("expected ErrIndexLocked for relative path when symlink held, got: %v", errRel2)
	}

	if err := heldSym.Release(); err != nil {
		t.Fatalf("failed to release symlink lock: %v", err)
	}

	// Test 3: Now absolute acquires without issue
	heldAbs, err := store.AcquireManifestLock(ctx, manifestReal)
	if err != nil {
		t.Fatalf("failed to acquire absolute lock: %v", err)
	}
	if err := heldAbs.Release(); err != nil {
		t.Fatalf("failed to release abs lock: %v", err)
	}
}

// AC3: Morte de processo libera lock e retry funciona.
func TestAC3_ProcessDeathReleasesLock(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte("target: death-test\n"), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Start helper process to hold the lock
	cmd := exec.Command(os.Args[0], "-test.run=TestAC3_ProcessDeathReleasesLock", "--", "lock-and-wait", manifestPath)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper process: %v", err)
	}

	// Wait for helper to signal "LOCKED"
	buf := make([]byte, 16)
	n, readErr := stdout.Read(buf)
	if readErr != nil && readErr != io.EOF {
		_ = cmd.Process.Kill()
		t.Fatalf("failed reading helper output: %v", readErr)
	}
	if !strings.Contains(string(buf[:n]), "LOCKED") {
		_ = cmd.Process.Kill()
		t.Fatalf("helper did not signal LOCKED, output: %s", string(buf[:n]))
	}

	// Verify main process cannot acquire lock while child lives
	shortCtx, cancelShort := context.WithTimeout(context.Background(), 80*time.Millisecond)
	_, errBlocked := store.AcquireManifestLock(shortCtx, manifestPath, store.WithPollInterval(2*time.Millisecond, 10*time.Millisecond))
	cancelShort()
	if errBlocked == nil || !errors.Is(errBlocked, store.ErrIndexLocked) {
		_ = cmd.Process.Kill()
		t.Fatalf("expected ErrIndexLocked while child holds lock, got: %v", errBlocked)
	}

	// Kill child process
	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("failed to kill child process: %v", err)
	}
	_ = cmd.Wait()

	// Retry acquisition in parent: must succeed now that child is dead
	retryCtx, cancelRetry := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelRetry()

	held, errRetry := store.AcquireManifestLock(retryCtx, manifestPath, store.WithPollInterval(2*time.Millisecond, 10*time.Millisecond))
	if errRetry != nil {
		t.Fatalf("retry failed to acquire lock after child death: %v", errRetry)
	}
	if err := held.Release(); err != nil {
		t.Fatalf("failed to release lock after retry: %v", err)
	}
}

// AC4: Timeout retorna index_locked.
func TestAC4_TimeoutReturnsIndexLocked(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte("target: timeout-test\n"), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	ctx := context.Background()
	held1, err := store.AcquireManifestLock(ctx, manifestPath)
	if err != nil {
		t.Fatalf("failed to acquire first lock: %v", err)
	}
	defer held1.Release()

	// Contender with 60ms timeout
	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Millisecond)
	defer cancel()

	_, errContender := store.AcquireManifestLock(timeoutCtx, manifestPath, store.WithPollInterval(2*time.Millisecond, 10*time.Millisecond))
	elapsed := time.Since(start)

	if errContender == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if !errors.Is(errContender, store.ErrIndexLocked) {
		t.Fatalf("expected errors.Is(err, ErrIndexLocked), got: %v", errContender)
	}
	if !strings.Contains(errContender.Error(), "index_locked") {
		t.Fatalf("expected error string to contain 'index_locked', got: %v", errContender)
	}
	if elapsed < 50*time.Millisecond {
		t.Fatalf("timeout returned prematurely after %v", elapsed)
	}
}

// AC5: Falha injetada de lock retorna lock_unavailable sem artefatos gravados.
func TestAC5_InjectedLockErrorReturnsLockUnavailable(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "gira.yaml")
	snapshotPath := filepath.Join(tempDir, "snapshot.json")
	storeDir := filepath.Join(tempDir, "store")

	injectedErr := errors.New("simulated remote filesystem locking error (ENOLCK)")
	targets := store.LockTargets{
		ManifestPath: manifestPath,
		StoreDir:     storeDir,
		SnapshotPath: snapshotPath,
	}

	opts := store.LockOptions{
		Timeout:           1 * time.Second,
		InjectedLockError: injectedErr,
	}

	executedWork := false
	err := store.WithLocks(context.Background(), targets, opts, func() error {
		executedWork = true
		// If lock was mistakenly acquired, work would execute and write files
		return os.WriteFile(snapshotPath, []byte("dirty snapshot"), 0644)
	})

	if err == nil {
		t.Fatalf("expected lock_unavailable error, got nil")
	}
	if !errors.Is(err, store.ErrLockUnavailable) {
		t.Fatalf("expected errors.Is(err, ErrLockUnavailable), got: %v", err)
	}
	if !strings.Contains(err.Error(), "lock_unavailable") {
		t.Fatalf("expected error string to contain 'lock_unavailable', got: %v", err)
	}
	if executedWork {
		t.Fatalf("work callback was executed despite lock failure (fail-closed violated)")
	}
	if _, statErr := os.Stat(snapshotPath); !os.IsNotExist(statErr) {
		t.Fatalf("artifact %s was written despite lock failure", snapshotPath)
	}
}

// AC6: Ordem M→S→P é verificada sem deadlock.
func TestAC6_StrictLockOrderMSP(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte("test\n"), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
	storeDir := filepath.Join(tempDir, "store")
	snapshotPath := filepath.Join(tempDir, "snapshot.json")

	// 1. Verify standard acquisition M -> S -> P and release P -> S -> M
	var acquiredOrder []store.LockKind
	var releasedOrder []store.LockKind
	var mu sync.Mutex

	targets := store.LockTargets{
		ManifestPath: manifestPath,
		StoreDir:     storeDir,
		SnapshotPath: snapshotPath,
	}

	opts := []store.LockOption{
		store.WithTimeout(5 * time.Second),
		store.WithLockHooks(
			func(kind store.LockKind, path string) {
				mu.Lock()
				acquiredOrder = append(acquiredOrder, kind)
				mu.Unlock()
			},
			func(kind store.LockKind, path string) {
				mu.Lock()
				releasedOrder = append(releasedOrder, kind)
				mu.Unlock()
			},
		),
	}

	held, err := store.AcquireLocks(context.Background(), targets, opts...)
	if err != nil {
		t.Fatalf("failed to acquire M-S-P locks: %v", err)
	}

	// Verify acquired order is strictly M -> S -> P
	expectedAcquired := []store.LockKind{store.LockKindManifest, store.LockKindStore, store.LockKindSnapshot}
	if len(acquiredOrder) != len(expectedAcquired) {
		t.Fatalf("expected %d acquired locks, got %d", len(expectedAcquired), len(acquiredOrder))
	}
	for i, k := range expectedAcquired {
		if acquiredOrder[i] != k {
			t.Errorf("acquired lock %d: expected %s, got %s", i, k, acquiredOrder[i])
		}
	}

	// Release locks
	if err := held.Release(); err != nil {
		t.Fatalf("failed to release locks: %v", err)
	}

	// Verify release order is strictly P -> S -> M
	expectedReleased := []store.LockKind{store.LockKindSnapshot, store.LockKindStore, store.LockKindManifest}
	if len(releasedOrder) != len(expectedReleased) {
		t.Fatalf("expected %d released locks, got %d", len(expectedReleased), len(releasedOrder))
	}
	for i, k := range expectedReleased {
		if releasedOrder[i] != k {
			t.Errorf("released lock %d: expected %s, got %s", i, k, releasedOrder[i])
		}
	}

	// 2. Verify LockChain detects and rejects out-of-order attempts
	lc := store.NewLockChain(store.WithTimeout(2 * time.Second))
	defer lc.Release()

	// Acquire S first
	if err := lc.AcquireStore(context.Background(), storeDir); err != nil {
		t.Fatalf("failed to acquire S in chain: %v", err)
	}

	// Attempting M after S must fail with ErrLockOrder
	errOutOfOrder := lc.AcquireManifest(context.Background(), manifestPath)
	if errOutOfOrder == nil || !errors.Is(errOutOfOrder, store.ErrLockOrder) {
		t.Fatalf("expected ErrLockOrder when acquiring M after S, got: %v", errOutOfOrder)
	}
	if !strings.Contains(errOutOfOrder.Error(), "lock_order_violation") {
		t.Fatalf("expected 'lock_order_violation' in error string, got: %v", errOutOfOrder)
	}
	if err := lc.Release(); err != nil {
		t.Fatalf("failed to release lc: %v", err)
	}

	// 3. Verify concurrent writers acquiring M -> S -> P do not deadlock
	const numWriters = 6
	var wg sync.WaitGroup
	wg.Add(numWriters)

	for w := 0; w < numWriters; w++ {
		go func(id int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			h, err := store.AcquireLocks(ctx, targets, store.WithPollInterval(2*time.Millisecond, 15*time.Millisecond))
			if err != nil {
				t.Errorf("writer %d failed: %v", id, err)
				return
			}
			time.Sleep(3 * time.Millisecond)
			if err := h.Release(); err != nil {
				t.Errorf("writer %d release failed: %v", id, err)
			}
		}(w)
	}

	wg.Wait()
}

// AC7: CI 3 SOs passa (testes determinísticos passam com go test -race ./...).
// Verifies that locks and partial acquisitions behave deterministically under race detector.
func TestAC7_DeterministicRaceFreeExecution(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "gira.yaml")
	if err := os.WriteFile(manifestPath, []byte("deterministic\n"), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
	storeDir := filepath.Join(tempDir, "store")

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for loop := 0; loop < 5; loop++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				held, err := store.AcquireLocks(ctx, store.LockTargets{
					ManifestPath: manifestPath,
					StoreDir:     storeDir,
				}, store.WithPollInterval(2*time.Millisecond, 10*time.Millisecond))
				if err != nil {
					cancel()
					t.Errorf("worker %d loop %d failed: %v", workerID, loop, err)
					return
				}
				time.Sleep(2 * time.Millisecond)
				_ = held.Release()
				cancel()
			}
		}(i)
	}
	wg.Wait()
}
