// Package clients contains small shared client interfaces used across internal packages.
package clients

import (
	"context"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

// ProfileClient fetches athlete profile data for tools and resources.
type ProfileClient interface {
	GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error)
}

var _ ProfileClient = (*intervals.Client)(nil)
