package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validConfig = `
version: 1
backup_sets:
  development:
    sources:
      - path: /home/neil/dev
        snapshot:
          provider: btrfs
          live_subvolume: /mnt/stillpoint-dev/live
          snapshots_directory: /mnt/stillpoint-dev/snapshots
    exclusions:
      presets: [node-dependencies]
      patterns: []
      refuse_to_exclude_git_tracked_files: true
    secrets:
      mode: select
      detect_compose_env_files: true
      detect_file_backed_compose_secrets: true
      additional_patterns: []
    agents:
      mode: hybrid
      acknowledgement_timeout: 2m
      dispatch_gate:
        adapter: command
        close: [agent-manager, gate, close]
        status: [agent-manager, gate, status]
        open: [agent-manager, gate, open]
      cooperative:
        adapter: command
        discover: [agent-manager, list]
        checkpoint: [agent-manager, checkpoint]
        acknowledgements: [agent-manager, acks]
        resume: [agent-manager, resume]
      fence:
        adapter: systemd-slice
        unit: agents.slice
    docker:
      discover: false
      compose_projects: []
    repository:
      primary:
        type: local
        path: /mnt/d/Backups/Stillpoint/development
        password_command: [secret-tool, lookup, stillpoint, development]
      replicas: []
    retention: {}
    verification: {}
    scheduling:
      enabled: false
      provider: auto
`

func TestLoadValidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stillpoint.yaml")
	if err := os.WriteFile(path, []byte(validConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := cfg.BackupSets["development"]; !ok {
		t.Fatal("development backup set not loaded")
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stillpoint.yaml")
	bad := strings.Replace(validConfig, "version: 1", "version: 1\nunknown: true", 1)
	if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() unexpectedly accepted an unknown field")
	}
}

func TestValidateRejectsPlaintextStyleSecretMode(t *testing.T) {
	cfg := Config{Version: 1, BackupSets: map[string]BackupSet{
		"bad": {
			Sources: []Source{{Path: "/tmp/source", Snapshot: SnapshotConfig{Provider: "btrfs"}}},
			Secrets: Secrets{Mode: "plaintext"},
			Agents:  Agents{Mode: "none"},
			Repository: Repository{Primary: PrimaryRepository{
				Type: "local", Path: "/tmp/repo", PasswordCommand: []string{"secret-helper"},
			}},
		},
	}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() unexpectedly accepted plaintext secret mode")
	}
}
