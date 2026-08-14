package doctor

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Check struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Required bool   `json:"required"`
	Detail   string `json:"detail"`
}

type Report struct {
	OS         string  `json:"os"`
	Arch       string  `json:"arch"`
	WSL        bool    `json:"wsl"`
	Filesystem string  `json:"filesystem,omitempty"`
	Checks     []Check `json:"checks"`
}

func Run(ctx context.Context, source string) Report {
	report := Report{OS: runtime.GOOS, Arch: runtime.GOARCH, WSL: isWSL()}
	report.Checks = append(report.Checks,
		binary("restic", true, "install Restic before running backups"),
		binary("btrfs", true, "install btrfs-progs for the MVP snapshot provider"),
		binary("systemctl", true, "systemd is required for the systemd-slice fence"),
		binary("docker", false, "Docker is optional unless Compose discovery is enabled"),
		binary("rclone", false, "rclone is optional and used for Google Drive replication"),
	)

	if path, err := exec.LookPath("docker"); err == nil {
		probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		cmd := exec.CommandContext(probeCtx, path, "compose", "version")
		if output, err := cmd.CombinedOutput(); err == nil {
			report.Checks = append(report.Checks, Check{Name: "docker-compose", Status: "present", Detail: firstLine(string(output))})
		} else {
			report.Checks = append(report.Checks, Check{Name: "docker-compose", Status: "warning", Detail: "Docker CLI found, but 'docker compose version' failed"})
		}
	}

	if source != "" {
		report.Filesystem = filesystem(ctx, source)
	}
	return report
}

func binary(name string, required bool, missingDetail string) Check {
	path, err := exec.LookPath(name)
	if err != nil {
		status := "optional-missing"
		if required {
			status = "missing"
		}
		return Check{Name: name, Status: status, Required: required, Detail: missingDetail}
	}
	return Check{Name: name, Status: "present", Required: required, Detail: path}
}

func filesystem(ctx context.Context, source string) string {
	path, err := exec.LookPath("findmnt")
	if err != nil {
		return "unknown (findmnt is unavailable)"
	}
	probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	output, err := exec.CommandContext(probeCtx, path, "-n", "-o", "FSTYPE", "-T", source).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	value := strings.ToLower(string(data))
	return strings.Contains(value, "microsoft") || strings.Contains(value, "wsl")
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if i := strings.IndexByte(value, '\n'); i >= 0 {
		return value[:i]
	}
	return value
}
