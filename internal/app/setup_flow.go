package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/cli/prompt"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
)

type setupDependencies struct {
	stdout           io.Writer
	store            credstore.Store
	prompter         SetupPrompter
	configExists     func(string) (bool, error)
	configWriter     SetupConfigWriter
	profileFetcher   SetupProfileFetcher
	timezoneDetector func() string
}

// RunSetup performs the terminal setup flow.
func RunSetup(ctx context.Context, opts SetupOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	deps := setupDefaults(opts)

	setupPrintIntro(deps.stdout)
	if ok, err := setupConfirmAPIKeyOverwrite(ctx, deps.store, deps.prompter, deps.stdout); err != nil || !ok {
		return err
	}
	configOverwriteAllowed, ok, err := setupConfigOverwriteAllowed(ctx, opts.ConfigPath, opts.Force, deps.configExists, deps.prompter, deps.stdout)
	if err != nil || !ok {
		return err
	}
	secret, err := setupReadAPIKey(ctx, deps.prompter)
	if err != nil {
		return err
	}
	profile, err := setupProfile(ctx, setupProfileOptions{offline: opts.Offline, secret: secret, fetcher: deps.profileFetcher, prompter: deps.prompter, stdout: deps.stdout})
	if err != nil {
		return err
	}
	timezoneName, err := setupTimezone(ctx, deps.prompter, deps.stdout, deps.timezoneDetector, opts.Offline)
	if err != nil {
		return err
	}
	if err := setupPersistConfigAndKey(ctx, opts.ConfigPath, configOverwriteAllowed, profile, timezoneName, secret, deps); err != nil {
		return err
	}
	return setupPrintCompletion(ctx, opts.ConfigPath, opts.Offline, profile, timezoneName, secret, deps)
}

func setupDefaults(opts SetupOptions) setupDependencies {
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
		prompter = prompt.NewTerminal(nil, stdout)
	}
	configExists := opts.ConfigExists
	if configExists == nil {
		configExists = fileExists
	}
	configWriter := opts.ConfigWriter
	if configWriter == nil {
		configWriter = config.Write
	}
	profileFetcher := opts.ProfileFetcher
	if profileFetcher == nil {
		profileFetcher = defaultSetupProfileFetcher
	}
	timezoneDetector := opts.TimezoneDetector
	if timezoneDetector == nil {
		timezoneDetector = detectLocalTimezone
	}
	return setupDependencies{stdout: stdout, store: store, prompter: prompter, configExists: configExists, configWriter: configWriter, profileFetcher: profileFetcher, timezoneDetector: timezoneDetector}
}

func setupPrintIntro(stdout io.Writer) {
	_, _ = fmt.Fprintln(stdout, "Welcome to icuvisor.")
	_, _ = fmt.Fprintln(stdout, "This setup stores your intervals.icu API key in the OS keychain and writes non-secret settings to your icuvisor config file.")
}

func setupConfirmAPIKeyOverwrite(ctx context.Context, store credstore.Store, prompter SetupPrompter, stdout io.Writer) (bool, error) {
	if _, err := store.Get(ctx, credstore.IntervalsAPIKeyAccount); err == nil {
		overwrite, promptErr := prompter.Confirm(ctx, "An API key is already stored. Overwrite? [y/N]", false)
		if promptErr != nil {
			return false, fmt.Errorf("confirm API key overwrite: %w", promptErr)
		}
		if !overwrite {
			_, _ = fmt.Fprintln(stdout, "Setup canceled; nothing changed.")
			return false, nil
		}
	} else if !errors.Is(err, credstore.ErrNotFound) {
		return false, fmt.Errorf("read intervals.icu API key from OS keychain service %q account %q: %w", credstore.ServiceName, credstore.IntervalsAPIKeyAccount, err)
	}
	return true, nil
}

func setupConfigOverwriteAllowed(ctx context.Context, path string, force bool, configExists func(string) (bool, error), prompter SetupPrompter, stdout io.Writer) (bool, bool, error) {
	configOverwriteAllowed := force
	exists, err := configExists(path)
	if err != nil {
		return false, false, fmt.Errorf("check config file %q: %w", path, err)
	}
	if exists && !force {
		overwrite, promptErr := prompter.Confirm(ctx, fmt.Sprintf("A config file already exists at %s. Overwrite? [y/N]", path), false)
		if promptErr != nil {
			return false, false, fmt.Errorf("confirm config overwrite: %w", promptErr)
		}
		if !overwrite {
			_, _ = fmt.Fprintln(stdout, "Setup canceled; nothing changed.")
			return false, false, nil
		}
		configOverwriteAllowed = true
	}
	return configOverwriteAllowed, true, nil
}

func setupReadAPIKey(ctx context.Context, prompter SetupPrompter) (string, error) {
	secret, err := prompter.ReadSecret(ctx, "Paste your intervals.icu API key (from https://intervals.icu/settings):")
	if err != nil {
		return "", fmt.Errorf("read intervals.icu API key: %w", err)
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", newSetupUsageError("missing intervals.icu API key")
	}
	return secret, nil
}

func setupPersistConfigAndKey(ctx context.Context, path string, allowOverwrite bool, profile SetupProfile, timezoneName string, secret string, deps setupDependencies) error {
	if err := deps.store.Set(ctx, credstore.IntervalsAPIKeyAccount, secret); err != nil {
		return fmt.Errorf("store intervals.icu API key in OS keychain service %q account %q: %w", credstore.ServiceName, credstore.IntervalsAPIKeyAccount, err)
	}
	storedSecret, err := deps.store.Get(ctx, credstore.IntervalsAPIKeyAccount)
	if err != nil {
		return fmt.Errorf("verify intervals.icu API key in OS keychain service %q account %q: %w", credstore.ServiceName, credstore.IntervalsAPIKeyAccount, err)
	}
	if storedSecret != secret {
		return errors.New("stored API key verification failed")
	}
	if err := deps.configWriter(ctx, path, config.Config{CredentialRef: config.DefaultCredentialReference(), AthleteID: profile.AthleteID, Timezone: timezoneName, APIBaseURL: config.DefaultAPIBaseURL}, config.WriteOptions{AllowOverwrite: allowOverwrite}); err != nil {
		return fmt.Errorf("write non-secret config: %w", err)
	}
	return nil
}

func setupPrintCompletion(ctx context.Context, path string, offline bool, profile SetupProfile, timezoneName string, secret string, deps setupDependencies) error {
	_, _ = fmt.Fprintf(deps.stdout, "Saved. Your key is in the OS keychain; athlete id %s + timezone %s are in %s.\n", profile.AthleteID, timezoneName, path)
	if offline {
		_, _ = fmt.Fprintln(deps.stdout, "Offline setup skipped the final intervals.icu test connection. Run an icuvisor tool when online to verify the key.")
	} else {
		verifiedProfile, err := deps.profileFetcher(ctx, secret, profile.AthleteID)
		if err != nil {
			return fmt.Errorf("final test connection failed: %w", err)
		}
		verifiedProfile.AthleteID = profile.AthleteID
		_, _ = fmt.Fprintf(deps.stdout, "Test connection OK: %s%s.\n", profileNameForOutput(verifiedProfile), profileFTPForOutput(verifiedProfile))
	}
	_, _ = fmt.Fprintln(deps.stdout, "Next: point Claude Desktop at icuvisor — see https://icuvisor.app/connect/claude-desktop/")
	return nil
}
