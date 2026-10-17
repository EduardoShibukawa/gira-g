package windowsgates

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// TestRunnerContext records who the process is and which git it will meet.
func TestRunnerContext(t *testing.T) {
	f := map[string]any{
		"go_version": runtime.Version(),
		"runner_os":  os.Getenv("RUNNER_OS"),
		"image_os":   os.Getenv("ImageOS"),
		"image_ver":  os.Getenv("ImageVersion"),
		"temp_dir":   os.TempDir(),
	}
	if p, err := exec.LookPath("git"); err == nil {
		f["git_path"] = p
	} else {
		f["git_path"] = errStr(err)
	}
	f["git_version"] = trim(git("", nil, "--version").Stdout)

	switch runtime.GOOS {
	case "windows":
		f["os_version"] = trim(run("", nil, nil, "cmd", "/c", "ver").Stdout)
		ps := func(script string) cmdResult {
			return run("", nil, nil, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
		}
		// True only for an elevated token (UAC-filtered admin returns False).
		f["token_elevated_admin_role"] = trim(ps("([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)").Stdout)
		f["user"] = trim(run("", nil, nil, "whoami").Stdout)
		groups := run("", nil, nil, "whoami", "/groups")
		f["whoami_groups_exit"], f["whoami_groups_stderr"] = groups.Exit, trim(groups.Stderr)
		for _, line := range strings.Split(groups.Stdout, "\n") {
			if strings.Contains(line, "Mandatory") || strings.Contains(line, "S-1-16-") || strings.Contains(line, "Administrators") {
				f["whoami_groups_relevant_"+strings.TrimSpace(strings.Fields(line + " x")[0])] = strings.Join(strings.Fields(line), " ")
			}
		}
		priv := run("", nil, nil, "whoami", "/priv")
		f["whoami_priv_exit"], f["whoami_priv_stderr"] = priv.Exit, trim(priv.Stderr)
		for _, line := range strings.Split(priv.Stdout, "\n") {
			if strings.Contains(line, "SymbolicLink") {
				f["se_create_symbolic_link"] = strings.Join(strings.Fields(line), " ")
			}
		}
		f["whoami_priv_raw"] = priv.Stdout
		f["whoami_groups_raw"] = groups.Stdout
		reg := run("", nil, nil, "reg", "query", `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock`, "/v", "AllowDevelopmentWithoutDevLicense")
		f["developer_mode_reg_exit"] = reg.Exit
		f["developer_mode_reg"] = strings.Join(strings.Fields(reg.Stdout+" "+reg.Stderr), " ")
	default:
		f["os_version"] = trim(run("", nil, nil, "uname", "-a").Stdout)
		if runtime.GOOS == "darwin" {
			f["sw_vers"] = strings.Join(strings.Fields(run("", nil, nil, "sw_vers").Stdout), " ")
		}
		f["uid"] = os.Getuid()
	}

	// core.symlinks as git decides it with and without system config, and a
	// freshly initialized repo (git probes the filesystem at init time).
	_, env := isolatedHome(t, false)
	r := git("", env, "config", "--show-origin", "--get-all", "core.symlinks")
	f["core_symlinks_with_system_config"] = map[string]any{"exit": r.Exit, "stdout": trim(r.Stdout)}
	sys := git("", env, "config", "--system", "--show-origin", "--list")
	f["system_config_exit"] = sys.Exit
	f["system_config"] = trim(sys.Stdout)
	repo := t.TempDir()
	git(repo, env, "init", "-q", repo)
	r = git(repo, env, "config", "--local", "--get", "core.symlinks")
	f["core_symlinks_after_init"] = map[string]any{"exit": r.Exit, "stdout": trim(r.Stdout)}
	rec(t, "runner_context", f)
}
