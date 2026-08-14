package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/viterineil/stillpoint/internal/config"
	"github.com/viterineil/stillpoint/internal/doctor"
	"github.com/viterineil/stillpoint/internal/excludes"
	"github.com/viterineil/stillpoint/internal/safety"
	"github.com/viterineil/stillpoint/internal/version"
)

const (
	exitOK          = 0
	exitRuntime     = 1
	exitUsage       = 2
	exitUnavailable = 3
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return exitOK
	}

	switch args[0] {
	case "help", "-h", "--help":
		printHelp(stdout)
		return exitOK
	case "version":
		fmt.Fprintln(stdout, version.String())
		return exitOK
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "plan":
		return runPlan(args[1:], stdout, stderr)
	case "config":
		return runConfig(args[1:], stdout, stderr)
	case "backup":
		fmt.Fprintln(stderr, "backup execution is intentionally disabled in this bootstrap release")
		fmt.Fprintln(stderr, "the Btrfs, Restic, and durable recovery adapters must pass integration tests before this command is enabled")
		return exitUnavailable
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printHelp(stderr)
		return exitUsage
	}
}

func runDoctor(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "write machine-readable JSON")
	source := flags.String("source", "", "inspect the filesystem containing this source path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "doctor does not accept positional arguments")
		return exitUsage
	}

	report := doctor.Run(context.Background(), *source)
	if *jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "write report: %v\n", err)
			return exitRuntime
		}
	} else {
		fmt.Fprintf(stdout, "Environment: %s/%s", report.OS, report.Arch)
		if report.WSL {
			fmt.Fprint(stdout, " (WSL)")
		}
		fmt.Fprintln(stdout)
		if report.Filesystem != "" {
			fmt.Fprintf(stdout, "Source filesystem: %s\n", report.Filesystem)
		}
		for _, check := range report.Checks {
			fmt.Fprintf(stdout, "%-18s %-16s %s\n", check.Name, check.Status, check.Detail)
		}
	}

	for _, check := range report.Checks {
		if check.Required && check.Status == "missing" {
			return exitRuntime
		}
	}
	return exitOK
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type planOutput struct {
	BackupSet         string   `json:"backup_set"`
	Sources           []string `json:"sources"`
	LocalRepository   string   `json:"local_repository"`
	SnapshotProviders []string `json:"snapshot_providers"`
	SecretMode        string   `json:"secret_mode"`
	AgentMode         string   `json:"agent_mode"`
	DefaultExcludes   []string `json:"default_excludes"`
	Warnings          []string `json:"warnings,omitempty"`
	Executable        bool     `json:"executable"`
}

func runPlan(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "path to stillpoint.yaml")
	setName := flags.String("set", "", "backup set name from the config")
	repository := flags.String("repository", "", "local Restic repository (direct mode)")
	jsonOutput := flags.Bool("json", false, "write machine-readable JSON")
	var sources stringList
	flags.Var(&sources, "source", "source directory; repeat for multiple sources (direct mode)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "plan does not accept positional arguments")
		return exitUsage
	}

	plan, err := buildPlan(*configPath, *setName, sources, *repository)
	if err != nil {
		fmt.Fprintf(stderr, "plan: %v\n", err)
		return exitRuntime
	}
	if *jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(plan); err != nil {
			fmt.Fprintf(stderr, "write plan: %v\n", err)
			return exitRuntime
		}
		return exitOK
	}

	fmt.Fprintf(stdout, "Backup set:       %s\n", plan.BackupSet)
	for _, source := range plan.Sources {
		fmt.Fprintf(stdout, "Source:           %s\n", source)
	}
	fmt.Fprintf(stdout, "Local repository: %s\n", plan.LocalRepository)
	fmt.Fprintf(stdout, "Snapshot:         %s\n", strings.Join(plan.SnapshotProviders, ", "))
	fmt.Fprintf(stdout, "Agent mode:       %s\n", plan.AgentMode)
	fmt.Fprintf(stdout, "Secret mode:      %s (whole-repository Restic encryption)\n", plan.SecretMode)
	fmt.Fprintln(stdout, "Default exclusions:")
	for _, pattern := range plan.DefaultExcludes {
		fmt.Fprintf(stdout, "  - %s\n", pattern)
	}
	for _, warning := range plan.Warnings {
		fmt.Fprintf(stdout, "WARNING: %s\n", warning)
	}
	fmt.Fprintln(stdout, "Executable:       no (bootstrap safety gate)")
	return exitOK
}

func buildPlan(configPath, requestedSet string, directSources []string, directRepository string) (planOutput, error) {
	if configPath != "" {
		if len(directSources) != 0 || directRepository != "" {
			return planOutput{}, errors.New("--config cannot be combined with --source or --repository")
		}
		cfg, err := config.Load(configPath)
		if err != nil {
			return planOutput{}, err
		}
		name, set, err := selectSet(cfg, requestedSet)
		if err != nil {
			return planOutput{}, err
		}
		var paths, providers []string
		for _, source := range set.Sources {
			paths = append(paths, source.Path)
			providers = append(providers, source.Snapshot.Provider)
		}
		location := set.Repository.Primary.Location()
		if set.Repository.Primary.Type != "local" {
			return planOutput{}, fmt.Errorf("bootstrap planner requires a local primary repository, got %q", set.Repository.Primary.Type)
		}
		layout, err := safety.ValidateLocalLayout(paths, location)
		if err != nil {
			return planOutput{}, err
		}
		patterns := excludes.Defaults()
		patterns = append(patterns, set.Exclusions.Patterns...)
		return planOutput{
			BackupSet: name, Sources: layout.Sources, LocalRepository: layout.Destination,
			SnapshotProviders: unique(providers), SecretMode: set.Secrets.Mode,
			AgentMode: set.Agents.Mode, DefaultExcludes: unique(patterns),
			Warnings: planWarnings(set.Agents.Mode, set.Secrets.Mode), Executable: false,
		}, nil
	}

	if requestedSet != "" {
		return planOutput{}, errors.New("--set requires --config")
	}
	layout, err := safety.ValidateLocalLayout(directSources, directRepository)
	if err != nil {
		return planOutput{}, err
	}
	return planOutput{
		BackupSet: "direct", Sources: layout.Sources, LocalRepository: layout.Destination,
		SnapshotProviders: []string{"not-configured"}, SecretMode: "fail",
		AgentMode: "none", DefaultExcludes: excludes.Defaults(),
		Warnings: planWarnings("none", "fail"), Executable: false,
	}, nil
}

func selectSet(cfg config.Config, requested string) (string, config.BackupSet, error) {
	if requested != "" {
		set, ok := cfg.BackupSets[requested]
		if !ok {
			return "", config.BackupSet{}, fmt.Errorf("backup set %q does not exist", requested)
		}
		return requested, set, nil
	}
	if len(cfg.BackupSets) != 1 {
		names := make([]string, 0, len(cfg.BackupSets))
		for name := range cfg.BackupSets {
			names = append(names, name)
		}
		sort.Strings(names)
		return "", config.BackupSet{}, fmt.Errorf("choose a backup set with --set (available: %s)", strings.Join(names, ", "))
	}
	for name, set := range cfg.BackupSets {
		return name, set, nil
	}
	panic("unreachable")
}

func planWarnings(agentMode, secretMode string) []string {
	var warnings []string
	if agentMode == "none" {
		warnings = append(warnings, "agent coordination is disabled; this would be crash-consistent only")
	}
	if secretMode == "exclude" {
		warnings = append(warnings, "secret-bearing files are intentionally excluded and may be required for recovery")
	}
	if secretMode == "select" {
		warnings = append(warnings, "interactive secret selections must be resolved before an unattended backup")
	}
	return warnings
}

func unique(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "validate" {
		fmt.Fprintln(stderr, "usage: stillpoint config validate [--config PATH]")
		return exitUsage
	}
	flags := flag.NewFlagSet("config validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("config", "stillpoint.yaml", "configuration file")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "config validate does not accept positional arguments")
		return exitUsage
	}
	cfg, err := config.Load(*path)
	if err != nil {
		fmt.Fprintf(stderr, "invalid configuration: %v\n", err)
		return exitRuntime
	}
	fmt.Fprintf(stdout, "%s is valid (schema version %d, %d backup set(s))\n", *path, cfg.Version, len(cfg.BackupSets))
	return exitOK
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `Stillpoint coordinates point-in-time backups for continuously changing dev environments.

Usage:
  stillpoint <command> [options]

Available now:
  config validate   Strictly parse and validate a versioned YAML config
  doctor            Inspect Restic, Btrfs, systemd, Docker, and rclone
  plan              Validate source/destination safety and print a dry plan
  version           Print build information

Safety-gated:
  backup            Disabled until snapshot/recovery integration tests pass

Examples:
  stillpoint doctor --source /home/neil/dev
  stillpoint config validate --config stillpoint.yaml
  stillpoint plan --config stillpoint.yaml --set development

Stillpoint is alpha software. A plan is not a backup; test every restore path.`)
}
