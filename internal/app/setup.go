package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
)

// SetupRunner executes the interactive setup subcommand.
type SetupRunner func(context.Context, SetupOptions) error

// SetupPrompter reads setup confirmations and masked secrets.
type SetupPrompter interface {
	Confirm(ctx context.Context, prompt string, defaultYes bool) (bool, error)
	ReadSecret(ctx context.Context, prompt string) (string, error)
}

// SetupOptions carries parsed setup flags and injectable dependencies.
type SetupOptions struct {
	ConfigPath string
	Offline    bool
	Force      bool
	Stdout     io.Writer
	Stderr     io.Writer

	CredentialStore credstore.Store
	Prompter        SetupPrompter
	ConfigExists    func(string) (bool, error)
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
		ConfigPath:      path,
		Offline:         parsed.offline,
		Force:           parsed.force,
		Stdout:          stdout,
		Stderr:          stderr,
		CredentialStore: store,
		Prompter:        prompter,
		ConfigExists:    opts.SetupConfigExists,
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
	if strings.TrimSpace(secret) == "" {
		return newSetupUsageError("missing intervals.icu API key")
	}
	_, _ = fmt.Fprintln(stdout, "Setup checks passed; connection verification and writing continue in this setup flow.")
	return nil
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
			return setupArgs{}, newSetupUsageError("unknown setup flag %q", arg)
		}
	}
	return parsed, nil
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
