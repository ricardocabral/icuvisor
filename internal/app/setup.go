package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/cli/prompt"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

// SetupRunner executes the interactive setup subcommand.
type SetupRunner func(context.Context, SetupOptions) error

// SetupPrompter reads setup confirmations, free-form answers, and masked secrets.
type SetupPrompter = prompt.Prompter

// SetupProfile contains the autodetected athlete fields setup needs.
type SetupProfile struct {
	AthleteID    string
	DisplayName  string
	FTP          int
	TimezoneName string
}

// SetupProfileFetcher verifies an API key against the supplied athlete ID and returns the athlete profile.
type SetupProfileFetcher func(ctx context.Context, apiKey, athleteID string) (SetupProfile, error)

// SetupConfigWriter writes non-secret setup config fields.
type SetupConfigWriter func(context.Context, string, config.Config, config.WriteOptions) error

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
	ConfigWriter     SetupConfigWriter
	ProfileFetcher   SetupProfileFetcher
	TimezoneDetector func() string
}

type setupProfileOptions struct {
	offline  bool
	secret   string
	fetcher  SetupProfileFetcher
	prompter SetupPrompter
	stdout   io.Writer
}

func setupProfile(ctx context.Context, opts setupProfileOptions) (SetupProfile, error) {
	athleteID, err := readAthleteID(ctx, opts.prompter)
	if err != nil {
		return SetupProfile{}, err
	}

	if opts.offline {
		_, _ = fmt.Fprintln(opts.stdout, "Offline setup skips intervals.icu verification. Your API key will be stored, but icuvisor cannot confirm it works until you run a tool.")
		return SetupProfile{AthleteID: athleteID}, nil
	}

	profile, err := opts.fetcher(ctx, opts.secret, athleteID)
	if err != nil {
		if errors.Is(err, intervals.ErrUnauthorized) || errors.Is(err, intervals.ErrNotFound) {
			return SetupProfile{}, errors.New("intervals.icu rejected the API key or athlete ID. Verify both at https://intervals.icu/settings (your URL there ends with /athlete/i<digits>/)")
		}
		return SetupProfile{}, fmt.Errorf("could not reach intervals.icu. Nothing was written. Re-run setup when online, or use --offline to store settings without verification: %w", err)
	}
	profile.AthleteID = athleteID
	if profile.FTP > 0 {
		_, _ = fmt.Fprintf(opts.stdout, "Checking intervals.icu… connected as %q (athlete %s, FTP %d W).\n", profileNameForOutput(profile), profile.AthleteID, profile.FTP)
	} else {
		_, _ = fmt.Fprintf(opts.stdout, "Checking intervals.icu… connected as %q (athlete %s).\n", profileNameForOutput(profile), profile.AthleteID)
	}
	return profile, nil
}

func readAthleteID(ctx context.Context, prompter SetupPrompter) (string, error) {
	answer, err := prompter.ReadLine(ctx, "Athlete ID (find yours in the intervals.icu URL, e.g. i12345 or 12345):")
	if err != nil {
		return "", fmt.Errorf("read athlete ID: %w", err)
	}
	normalized, err := config.NormalizeAthleteID(answer)
	if err != nil {
		return "", err
	}
	return normalized, nil
}

func profileNameForOutput(profile SetupProfile) string {
	name := strings.TrimSpace(profile.DisplayName)
	if name == "" {
		return profile.AthleteID
	}
	return name
}

func profileFTPForOutput(profile SetupProfile) string {
	if profile.FTP <= 0 {
		return ""
	}
	return fmt.Sprintf(", FTP %d W", profile.FTP)
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

func defaultSetupProfileFetcher(ctx context.Context, apiKey, athleteID string) (SetupProfile, error) {
	client, err := intervals.NewClient(intervals.Options{Config: config.Config{APIKey: apiKey, AthleteID: athleteID, APIBaseURL: config.DefaultAPIBaseURL, HTTPTimeout: config.DefaultHTTPTimeout}})
	if err != nil {
		return SetupProfile{}, err
	}
	profile, err := client.GetAthleteProfile(ctx)
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
