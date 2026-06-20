package intervals

import (
	"context"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
)

// UpstreamAthlete is the sanitized public subset returned by the upstream athletes probe.
type UpstreamAthlete struct {
	AthleteID string `json:"athlete_id"`
	Name      string `json:"name,omitempty"`
}

type upstreamAthleteDTO struct {
	AthleteID   string `json:"athlete_id"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	AthleteName string `json:"athlete_name"`
	DisplayName string `json:"display_name"`
}

// ListUpstreamAthletes retrieves and sanitizes GET /api/v1/athletes probe rows.
func (c *Client) ListUpstreamAthletes(ctx context.Context) ([]UpstreamAthlete, error) {
	var rows []upstreamAthleteDTO
	if err := c.doJSON(ctx, &rows, "athletes"); err != nil {
		return nil, fmt.Errorf("listing upstream athletes: %w", err)
	}

	out := make([]UpstreamAthlete, 0, len(rows))
	for i, row := range rows {
		athleteID, err := sanitizeUpstreamAthleteID(row)
		if err != nil {
			return nil, fmt.Errorf("sanitizing upstream athlete row %d: %w", i, err)
		}
		out = append(out, UpstreamAthlete{AthleteID: athleteID, Name: sanitizeUpstreamAthleteName(row)})
	}
	return out, nil
}

func sanitizeUpstreamAthleteID(row upstreamAthleteDTO) (string, error) {
	candidate := strings.TrimSpace(row.AthleteID)
	if candidate == "" {
		candidate = strings.TrimSpace(row.ID)
	}
	return config.NormalizeAthleteID(candidate)
}

func sanitizeUpstreamAthleteName(row upstreamAthleteDTO) string {
	for _, candidate := range []string{row.Name, row.AthleteName, row.DisplayName} {
		if name := strings.TrimSpace(candidate); name != "" {
			return name
		}
	}
	return ""
}
