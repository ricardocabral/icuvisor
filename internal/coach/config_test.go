package coach

import (
	"errors"
	"strings"
	"testing"
)

func TestParseMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Mode
		wantErr bool
	}{
		{name: "empty defaults off", input: "", want: ModeOff},
		{name: "off", input: "off", want: ModeOff},
		{name: "on mixed case", input: " ON ", want: ModeOn},
		{name: "auto", input: "auto", want: ModeAuto},
		{name: "invalid", input: "yes", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMode(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("ParseMode() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMode() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("ParseMode() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidateConfigNormalizesRosterAndPatterns(t *testing.T) {
	t.Parallel()

	cfg, err := ValidateConfig(Config{Athletes: []Athlete{
		{ID: "123", Label: " Rider ", AllowedTools: []string{" get_* ", "get_*"}, DeniedTools: []string{"delete_event"}},
	}, DefaultAthleteID: "123"}, ModeOn, testNormalizeAthleteID)
	if err != nil {
		t.Fatalf("ValidateConfig() error = %v", err)
	}
	if cfg.DefaultAthleteID != "i123" || len(cfg.Athletes) != 1 || cfg.Athletes[0].ID != "i123" || cfg.Athletes[0].Label != "Rider" {
		t.Fatalf("ValidateConfig() = %#v, want normalized roster", cfg)
	}
	if got := cfg.Athletes[0].AllowedTools; len(got) != 1 || got[0] != "get_*" {
		t.Fatalf("AllowedTools = %#v, want deduped get_*", got)
	}
}

func TestValidateConfigStateMachine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mode       Mode
		cfg        Config
		wantMode   Mode
		wantErrSub string
	}{
		{name: "auto empty disables", mode: ModeAuto, cfg: Config{}, wantMode: ModeOff},
		{name: "auto roster enables and fills single default", mode: ModeAuto, cfg: Config{Athletes: []Athlete{{ID: "1", AllowedTools: []string{"*"}}}}, wantMode: ModeOn},
		{name: "on empty roster errors", mode: ModeOn, cfg: Config{}, wantErrSub: "coach mode is on"},
		{name: "multiple enabled athletes require default", mode: ModeOn, cfg: Config{Athletes: []Athlete{{ID: "1"}, {ID: "2"}}}, wantErrSub: "default_athlete_id is required"},
		{name: "off still validates invalid ACL", mode: ModeOff, cfg: Config{Athletes: []Athlete{{ID: "1", AllowedTools: []string{"missing_tool"}}}}, wantErrSub: "unknown athlete-scoped tool"},
		{name: "empty allowed means deny all but is valid", mode: ModeOff, cfg: Config{Athletes: []Athlete{{ID: "1"}}}, wantMode: ModeOff},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateConfig(tc.cfg, tc.mode, testNormalizeAthleteID)
			if tc.wantErrSub != "" {
				if err == nil {
					t.Fatal("ValidateConfig() error = nil, want error")
				}
				if !strings.Contains(err.Error(), tc.wantErrSub) {
					t.Fatalf("error = %q, want %q", err, tc.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateConfig() error = %v", err)
			}
			if effective := EffectiveMode(tc.mode, got); effective != tc.wantMode {
				t.Fatalf("EffectiveMode() = %q, want %q", effective, tc.wantMode)
			}
		})
	}
}

func testNormalizeAthleteID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "bad") {
		return "", errors.New("invalid athlete ID")
	}
	value = strings.TrimPrefix(strings.TrimPrefix(value, "i"), "I")
	return "i" + value, nil
}
