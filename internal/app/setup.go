package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

// SetupRunner executes the interactive setup subcommand.
type SetupRunner func(context.Context, SetupOptions) error

// SetupPrompter reads setup confirmations, free-form answers, and masked secrets.
type SetupPrompter interface {
	Confirm(ctx context.Context, prompt string, defaultYes bool) (bool, error)
	ReadLine(ctx context.Context, prompt string) (string, error)
	ReadSecret(ctx context.Context, prompt string) (string, error)
}

// SetupProfile contains the autodetected athlete fields setup needs.
type SetupProfile struct {
	AthleteID    string
	DisplayName  string
	FTP          int
	TimezoneName string
}

// SetupProfileFetcher verifies an API key and returns the authenticated athlete profile.
type SetupProfileFetcher func(context.Context, string) (SetupProfile, error)

// SetupOptions carries parsed setup flags and injectable dependencies.
type SetupOptions struct {
	ConfigPath string
	Offline    bool
	Force      bool
	Stdout     io.Writer
	Stderr     io.Writer

	CredentialStore  credstore.Store
	Prompter         SetupPrompter
	ConfigExists     func(string) (bool, error)
	ProfileFetcher   SetupProfileFetcher
	TimezoneDetector func() string
}

type setupArgs struct {
	configPath string
	offline    bool
	force      bool
	help       bool
}

func runSetupCommand(ctx context.Context, opts Options, args []string) error {
	parsed, err := parseSetupArgs(args)
	if err != nil {
		return err
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	if parsed.help {
		return writeSetupHelp(stdout)
	}
	path, err := resolveSetupConfigPath(parsed.configPath)
	if err != nil {
		return err
	}
	runner := opts.SetupRunner
	if runner == nil {
		runner = RunSetup
	}
	store := opts.SetupCredentialStore
	if store == nil {
		store = credstore.OSKeychain()
	}
	prompter := opts.SetupPrompter
	if prompter == nil {
		prompter = newTerminalPrompter(opts.Stdin, stdout)
	}
	return runner(ctx, SetupOptions{
		ConfigPath:       path,
		Offline:          parsed.offline,
		Force:            parsed.force,
		Stdout:           stdout,
		Stderr:           stderr,
		CredentialStore:  store,
		Prompter:         prompter,
		ConfigExists:     opts.SetupConfigExists,
		ProfileFetcher:   opts.SetupProfileFetcher,
		TimezoneDetector: opts.SetupTimezoneDetector,
	})
}

// RunSetup performs the terminal setup flow.
func RunSetup(ctx context.Context, opts SetupOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	store := opts.CredentialStore
	if store == nil {
		store = credstore.OSKeychain()
	}
	prompter := opts.Prompter
	if prompter == nil {
		prompter = newTerminalPrompter(nil, stdout)
	}
	configExists := opts.ConfigExists
	if configExists == nil {
		configExists = fileExists
	}
	profileFetcher := opts.ProfileFetcher
	if profileFetcher == nil {
		profileFetcher = defaultSetupProfileFetcher
	}
	timezoneDetector := opts.TimezoneDetector
	if timezoneDetector == nil {
		timezoneDetector = detectLocalTimezone
	}

	_, _ = fmt.Fprintln(stdout, "Welcome to icuvisor.")
	_, _ = fmt.Fprintln(stdout, "This setup stores your intervals.icu API key in the OS keychain and writes non-secret settings to your icuvisor config file.")

	if _, err := store.Get(ctx, credstore.IntervalsAPIKeyAccount); err == nil {
		overwrite, promptErr := prompter.Confirm(ctx, "An API key is already stored. Overwrite? [y/N]", false)
		if promptErr != nil {
			return fmt.Errorf("confirm API key overwrite: %w", promptErr)
		}
		if !overwrite {
			_, _ = fmt.Fprintln(stdout, "Setup canceled; nothing changed.")
			return nil
		}
	} else if !errors.Is(err, credstore.ErrNotFound) {
		return fmt.Errorf("read intervals.icu API key from OS keychain service %q account %q: %w", credstore.ServiceName, credstore.IntervalsAPIKeyAccount, err)
	}

	exists, err := configExists(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("check config file %q: %w", opts.ConfigPath, err)
	}
	if exists && !opts.Force {
		overwrite, promptErr := prompter.Confirm(ctx, fmt.Sprintf("A config file already exists at %s. Overwrite? [y/N]", opts.ConfigPath), false)
		if promptErr != nil {
			return fmt.Errorf("confirm config overwrite: %w", promptErr)
		}
		if !overwrite {
			_, _ = fmt.Fprintln(stdout, "Setup canceled; nothing changed.")
			return nil
		}
	}

	secret, err := prompter.ReadSecret(ctx, "Paste your intervals.icu API key (from https://intervals.icu/settings):")
	if err != nil {
		return fmt.Errorf("read intervals.icu API key: %w", err)
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return newSetupUsageError("missing intervals.icu API key")
	}

	profile, err := setupProfile(ctx, setupProfileOptions{offline: opts.Offline, secret: secret, fetcher: profileFetcher, prompter: prompter, stdout: stdout})
	if err != nil {
		return err
	}
	timezoneName, err := setupTimezone(ctx, prompter, stdout, timezoneDetector, opts.Offline)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "Setup checks passed for athlete %s with timezone %s; writing continues in this setup flow.\n", profile.AthleteID, timezoneName)
	return nil
}

type setupProfileOptions struct {
	offline  bool
	secret   string
	fetcher  SetupProfileFetcher
	prompter SetupPrompter
	stdout   io.Writer
}

func setupProfile(ctx context.Context, opts setupProfileOptions) (SetupProfile, error) {
	if opts.offline {
		_, _ = fmt.Fprintln(opts.stdout, "Offline setup skips intervals.icu verification. Your API key will be stored, but icuvisor cannot confirm it works until you run a tool.")
		athleteID, err := opts.prompter.ReadLine(ctx, "Athlete ID (accepts 12345 or i12345):")
		if err != nil {
			return SetupProfile{}, fmt.Errorf("read athlete ID: %w", err)
		}
		normalized, err := config.NormalizeAthleteID(athleteID)
		if err != nil {
			return SetupProfile{}, err
		}
		return SetupProfile{AthleteID: normalized}, nil
	}

	profile, err := opts.fetcher(ctx, opts.secret)
	if err != nil {
		if errors.Is(err, intervals.ErrUnauthorized) {
			return SetupProfile{}, errors.New("API key not accepted by intervals.icu. Double-check the key on https://intervals.icu/settings.")
		}
		return SetupProfile{}, fmt.Errorf("could not reach intervals.icu. Nothing was written. Re-run setup when online, or use --offline to store settings without verification: %w", err)
	}
	normalized, err := config.NormalizeAthleteID(profile.AthleteID)
	if err != nil {
		return SetupProfile{}, fmt.Errorf("normalizing autodetected athlete ID: %w", err)
	}
	profile.AthleteID = normalized
	name := strings.TrimSpace(profile.DisplayName)
	if name == "" {
		name = normalized
	}
	if profile.FTP > 0 {
		_, _ = fmt.Fprintf(opts.stdout, "Checking intervals.icu… connected as %q (athlete %s, FTP %d W).\n", name, normalized, profile.FTP)
	} else {
		_, _ = fmt.Fprintf(opts.stdout, "Checking intervals.icu… connected as %q (athlete %s).\n", name, normalized)
	}
	return profile, nil
}

func setupTimezone(ctx context.Context, prompter SetupPrompter, stdout io.Writer, detector func() string, offline bool) (string, error) {
	if offline {
		answer, err := prompter.ReadLine(ctx, "Timezone (IANA name, for example Europe/Madrid):")
		if err != nil {
			return "", fmt.Errorf("read timezone: %w", err)
		}
		return validateTimezone(answer)
	}

	detected := strings.TrimSpace(detector())
	if detected == "" {
		detected = config.DefaultTimezone
	}
	if _, err := time.LoadLocation(detected); err != nil {
		detected = config.DefaultTimezone
	}
	useDetected, err := prompter.Confirm(ctx, fmt.Sprintf("Detected timezone: %s. Use this? [Y/n]", detected), true)
	if err != nil {
		return "", fmt.Errorf("confirm timezone: %w", err)
	}
	if useDetected {
		return detected, nil
	}
	answer, err := prompter.ReadLine(ctx, "Timezone (IANA name, for example Europe/Madrid):")
	if err != nil {
		return "", fmt.Errorf("read timezone: %w", err)
	}
	timezoneName, err := validateTimezone(answer)
	if err != nil {
		return "", err
	}
	_, _ = fmt.Fprintf(stdout, "Using timezone: %s.\n", timezoneName)
	return timezoneName, nil
}

func validateTimezone(value string) (string, error) {
	timezoneName := strings.TrimSpace(value)
	if _, err := time.LoadLocation(timezoneName); err != nil {
		return "", fmt.Errorf("invalid timezone %q; use an IANA timezone like Europe/Madrid", timezoneName)
	}
	return timezoneName, nil
}

func defaultSetupProfileFetcher(ctx context.Context, apiKey string) (SetupProfile, error) {
	client, err := intervals.NewClient(intervals.Options{Config: config.Config{APIKey: apiKey, AthleteID: "0", APIBaseURL: config.DefaultAPIBaseURL, HTTPTimeout: config.DefaultHTTPTimeout}})
	if err != nil {
		return SetupProfile{}, err
	}
	profile, err := client.GetAuthenticatedAthleteProfile(ctx)
	if err != nil {
		return SetupProfile{}, err
	}
	return setupProfileFromIntervals(profile), nil
}

func setupProfileFromIntervals(profile intervals.AthleteWithSportSettings) SetupProfile {
	return SetupProfile{AthleteID: profile.ID, DisplayName: displayName(profile), FTP: profileFTP(profile), TimezoneName: profile.Timezone}
}

func displayName(profile intervals.AthleteWithSportSettings) string {
	if strings.TrimSpace(profile.Name) != "" {
		return strings.TrimSpace(profile.Name)
	}
	return strings.TrimSpace(strings.Join([]string{strings.TrimSpace(profile.FirstName), strings.TrimSpace(profile.LastName)}, " "))
}

func profileFTP(profile intervals.AthleteWithSportSettings) int {
	for _, sport := range profile.SportSettings {
		if sport.FTP > 0 {
			return sport.FTP
		}
	}
	return 0
}

func detectLocalTimezone() string {
	return detectLocalTimezoneWith(time.Local.String(), os.Getenv("TZ"), os.Readlink)
}

func detectLocalTimezoneWith(localName string, tzEnv string, readlink func(string) (string, error)) string {
	if zone, ok := validTimezoneName(tzEnv); ok {
		return zone
	}
	if zone, ok := validTimezoneName(localName); ok {
		return zone
	}
	if readlink != nil {
		if target, err := readlink("/etc/localtime"); err == nil {
			if zone, ok := zoneFromLocaltimeTarget(target); ok {
				return zone
			}
		}
	}
	return config.DefaultTimezone
}

func validTimezoneName(value string) (string, bool) {
	zone := strings.TrimSpace(value)
	zone = strings.TrimPrefix(zone, ":")
	if zone == "" || zone == "Local" || strings.HasPrefix(zone, "/") || strings.Contains(zone, "..") {
		return "", false
	}
	if _, err := time.LoadLocation(zone); err != nil {
		return "", false
	}
	return zone, true
}

func zoneFromLocaltimeTarget(target string) (string, bool) {
	trimmed := strings.TrimSpace(target)
	for _, marker := range []string{"/zoneinfo/", "/usr/share/zoneinfo/"} {
		if index := strings.LastIndex(trimmed, marker); index >= 0 {
			candidate := trimmed[index+len(marker):]
			return validTimezoneName(candidate)
		}
	}
	return "", false
}

func parseSetupArgs(args []string) (setupArgs, error) {
	var parsed setupArgs
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h" || arg == "help":
			parsed.help = true
		case arg == "--config":
			value, next, err := requireFlagValue(args, i, "--config", "/path/to/icuvisor.json")
			if err != nil {
				return setupArgs{}, setupUsageError(err)
			}
			parsed.configPath = value
			i = next
		case strings.HasPrefix(arg, "--config="):
			value, err := requireInlineFlagValue(arg, "--config", "/path/to/icuvisor.json")
			if err != nil {
				return setupArgs{}, setupUsageError(err)
			}
			parsed.configPath = value
		case arg == "--offline":
			parsed.offline = true
		case arg == "--force":
			parsed.force = true
		default:
			return setupArgs{}, newSetupUsageError("unknown setup flag %q", unknownSetupFlagName(arg))
		}
	}
	return parsed, nil
}

func unknownSetupFlagName(arg string) string {
	if strings.HasPrefix(arg, "--") {
		name, _, hasValue := strings.Cut(arg, "=")
		if hasValue {
			return name
		}
	}
	return arg
}

func setupUsageError(err error) error {
	var usageErr UsageError
	if errors.As(err, &usageErr) {
		msg := strings.TrimSuffix(usageErr.message, "\nRun 'icuvisor --help' for usage.")
		return newSetupUsageError("%s", msg)
	}
	return err
}

func newSetupUsageError(format string, args ...any) UsageError {
	return UsageError{message: fmt.Sprintf(format, args...) + "\nRun 'icuvisor setup --help' for usage."}
}

func resolveSetupConfigPath(path string) (string, error) {
	if trimmed := strings.TrimSpace(path); trimmed != "" {
		return trimmed, nil
	}
	if envPath := strings.TrimSpace(os.Getenv(config.EnvConfigPath)); envPath != "" {
		return envPath, nil
	}
	resolved, err := config.DefaultPath()
	if err != nil {
		return "", fmt.Errorf("resolve default config path: %w", err)
	}
	return resolved, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

type terminalPrompter struct {
	in     io.Reader
	out    io.Writer
	reader *bufio.Reader
}

func newTerminalPrompter(in io.Reader, out io.Writer) *terminalPrompter {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = io.Discard
	}
	return &terminalPrompter{in: in, out: out, reader: bufio.NewReader(in)}
}

func (p *terminalPrompter) Confirm(ctx context.Context, prompt string, defaultYes bool) (bool, error) {
	for {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		_, _ = fmt.Fprint(p.out, prompt+" ")
		answer, err := p.reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return false, err
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "" {
			return defaultYes, nil
		}
		switch answer {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			_, _ = fmt.Fprintln(p.out, "Please answer y or n.")
		}
	}
}

func (p *terminalPrompter) ReadLine(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	_, _ = fmt.Fprintln(p.out, prompt)
	_, _ = fmt.Fprint(p.out, "> ")
	answer, err := p.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(answer), nil
}

func (p *terminalPrompter) ReadSecret(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	file, ok := p.in.(*os.File)
	if !ok {
		return "", errors.New("masked API-key input requires an interactive terminal")
	}
	_, _ = fmt.Fprintln(p.out, prompt)
	_, _ = fmt.Fprint(p.out, "> ")
	secret, err := term.ReadPassword(int(file.Fd()))
	_, _ = fmt.Fprintln(p.out)
	if err != nil {
		return "", err
	}
	return string(secret), nil
}
