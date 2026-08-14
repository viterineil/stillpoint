package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const CurrentVersion = 1

type Config struct {
	Version    int                  `yaml:"version" json:"version"`
	BackupSets map[string]BackupSet `yaml:"backup_sets" json:"backup_sets"`
}

type BackupSet struct {
	Sources      []Source     `yaml:"sources" json:"sources"`
	Exclusions   Exclusions   `yaml:"exclusions" json:"exclusions"`
	Secrets      Secrets      `yaml:"secrets" json:"secrets"`
	Agents       Agents       `yaml:"agents" json:"agents"`
	Docker       Docker       `yaml:"docker" json:"docker"`
	Repository   Repository   `yaml:"repository" json:"repository"`
	Retention    Retention    `yaml:"retention" json:"retention"`
	Verification Verification `yaml:"verification" json:"verification"`
	Scheduling   Scheduling   `yaml:"scheduling" json:"scheduling"`
}

type Source struct {
	Path     string         `yaml:"path" json:"path"`
	Snapshot SnapshotConfig `yaml:"snapshot" json:"snapshot"`
}

type SnapshotConfig struct {
	Provider           string `yaml:"provider" json:"provider"`
	LiveSubvolume      string `yaml:"live_subvolume" json:"live_subvolume"`
	SnapshotsDirectory string `yaml:"snapshots_directory" json:"snapshots_directory"`
}

type Exclusions struct {
	Presets                     []string `yaml:"presets" json:"presets"`
	Patterns                    []string `yaml:"patterns" json:"patterns"`
	RefuseToExcludeTrackedFiles bool     `yaml:"refuse_to_exclude_git_tracked_files" json:"refuse_to_exclude_git_tracked_files"`
}

type Secrets struct {
	Mode                           string   `yaml:"mode" json:"mode"`
	DetectComposeEnvFiles          bool     `yaml:"detect_compose_env_files" json:"detect_compose_env_files"`
	DetectFileBackedComposeSecrets bool     `yaml:"detect_file_backed_compose_secrets" json:"detect_file_backed_compose_secrets"`
	AdditionalPatterns             []string `yaml:"additional_patterns" json:"additional_patterns"`
}

type Agents struct {
	Mode                   string             `yaml:"mode" json:"mode"`
	AcknowledgementTimeout string             `yaml:"acknowledgement_timeout" json:"acknowledgement_timeout"`
	DispatchGate           DispatchGate       `yaml:"dispatch_gate" json:"dispatch_gate"`
	Cooperative            CooperativeAdapter `yaml:"cooperative" json:"cooperative"`
	Fence                  FenceAdapter       `yaml:"fence" json:"fence"`
}

type DispatchGate struct {
	Adapter string   `yaml:"adapter" json:"adapter"`
	Close   []string `yaml:"close" json:"close"`
	Status  []string `yaml:"status" json:"status"`
	Open    []string `yaml:"open" json:"open"`
}

type CooperativeAdapter struct {
	Adapter          string   `yaml:"adapter" json:"adapter"`
	Discover         []string `yaml:"discover" json:"discover"`
	Checkpoint       []string `yaml:"checkpoint" json:"checkpoint"`
	Acknowledgements []string `yaml:"acknowledgements" json:"acknowledgements"`
	Resume           []string `yaml:"resume" json:"resume"`
}

type FenceAdapter struct {
	Adapter string `yaml:"adapter" json:"adapter"`
	Unit    string `yaml:"unit" json:"unit"`
}

type Docker struct {
	Discover        bool             `yaml:"discover" json:"discover"`
	ComposeProjects []ComposeProject `yaml:"compose_projects" json:"compose_projects"`
}

type ComposeProject struct {
	Files        []string       `yaml:"files" json:"files"`
	Include      ComposeInclude `yaml:"include" json:"include"`
	NamedVolumes NamedVolumes   `yaml:"named_volumes" json:"named_volumes"`
	Databases    []Database     `yaml:"databases" json:"databases"`
}

type ComposeInclude struct {
	ComposeFiles      bool   `yaml:"compose_files" json:"compose_files"`
	Dockerfiles       bool   `yaml:"dockerfiles" json:"dockerfiles"`
	Dockerignore      bool   `yaml:"dockerignore" json:"dockerignore"`
	EnvFiles          string `yaml:"env_files" json:"env_files"`
	FileBackedSecrets string `yaml:"file_backed_secrets" json:"file_backed_secrets"`
	BindMounts        string `yaml:"bind_mounts" json:"bind_mounts"`
}

type NamedVolumes struct {
	Default  string `yaml:"default" json:"default"`
	Fallback string `yaml:"fallback" json:"fallback"`
}

type Database struct {
	Service      string `yaml:"service" json:"service"`
	Adapter      string `yaml:"adapter" json:"adapter"`
	IncludeRoles bool   `yaml:"include_roles" json:"include_roles"`
}

type Repository struct {
	Primary  PrimaryRepository `yaml:"primary" json:"primary"`
	Replicas []Replica         `yaml:"replicas" json:"replicas"`
}

type PrimaryRepository struct {
	Type            string   `yaml:"type" json:"type"`
	Path            string   `yaml:"path" json:"path"`
	Repository      string   `yaml:"repository" json:"repository"`
	PasswordCommand []string `yaml:"password_command" json:"password_command"`
}

func (p PrimaryRepository) Location() string {
	if p.Path != "" {
		return p.Path
	}
	return p.Repository
}

type Replica struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Repository  string `yaml:"repository" json:"repository"`
	Credentials string `yaml:"credentials" json:"credentials"`
	Replicate   string `yaml:"replicate" json:"replicate"`
}

type Retention struct {
	KeepHourly  int `yaml:"keep_hourly" json:"keep_hourly"`
	KeepDaily   int `yaml:"keep_daily" json:"keep_daily"`
	KeepWeekly  int `yaml:"keep_weekly" json:"keep_weekly"`
	KeepMonthly int `yaml:"keep_monthly" json:"keep_monthly"`
	KeepYearly  int `yaml:"keep_yearly" json:"keep_yearly"`
}

type Verification struct {
	Structural   string `yaml:"structural" json:"structural"`
	SampledData  string `yaml:"sampled_data" json:"sampled_data"`
	RestoreDrill string `yaml:"restore_drill" json:"restore_drill"`
}

type Scheduling struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Provider string `yaml:"provider" json:"provider"`
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Config{}, errors.New("config must contain exactly one YAML document")
		}
		return Config{}, fmt.Errorf("decode trailing config data: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.Version != CurrentVersion {
		return fmt.Errorf("unsupported config version %d (expected %d)", c.Version, CurrentVersion)
	}
	if len(c.BackupSets) == 0 {
		return errors.New("config contains no backup_sets")
	}
	for name, set := range c.BackupSets {
		if strings.TrimSpace(name) == "" {
			return errors.New("backup set name may not be empty")
		}
		if err := set.validate(); err != nil {
			return fmt.Errorf("backup set %q: %w", name, err)
		}
	}
	return nil
}

func (s BackupSet) validate() error {
	if len(s.Sources) == 0 {
		return errors.New("at least one source is required")
	}
	for _, source := range s.Sources {
		if !filepath.IsAbs(source.Path) {
			return fmt.Errorf("source path must be absolute: %q", source.Path)
		}
		if source.Snapshot.Provider == "" {
			return fmt.Errorf("source %q has no snapshot provider", source.Path)
		}
	}
	validSecretModes := map[string]bool{"include-encrypted": true, "exclude": true, "select": true, "fail": true}
	if !validSecretModes[s.Secrets.Mode] {
		return fmt.Errorf("invalid secrets mode %q", s.Secrets.Mode)
	}
	validAgentModes := map[string]bool{"hybrid": true, "command": true, "systemd-slice": true, "none": true}
	if !validAgentModes[s.Agents.Mode] {
		return fmt.Errorf("invalid agents mode %q", s.Agents.Mode)
	}
	if s.Repository.Primary.Type == "" {
		return errors.New("primary repository type is required")
	}
	if strings.TrimSpace(s.Repository.Primary.Location()) == "" {
		return errors.New("primary repository location is required")
	}
	if len(s.Repository.Primary.PasswordCommand) == 0 {
		return errors.New("primary repository password_command is required")
	}
	for _, arg := range s.Repository.Primary.PasswordCommand {
		if strings.TrimSpace(arg) == "" {
			return errors.New("password_command may not contain empty arguments")
		}
	}
	return nil
}
