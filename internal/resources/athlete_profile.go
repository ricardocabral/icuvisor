package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ricardocabral/icuvisor/internal/athleteprofile"
	"github.com/ricardocabral/icuvisor/internal/clients"
)

const (
	AthleteProfileURI      = "icuvisor://athlete-profile"
	AthleteProfileMIMEType = "application/json"
	AthleteProfileTTL      = 15 * time.Minute
)

type athleteProfileOptions struct {
	client           clients.ProfileClient
	version          string
	timezoneFallback string
	debugMetadata    bool
	ttl              time.Duration
	now              func() time.Time
}

type athleteProfileReader struct {
	client           clients.ProfileClient
	version          string
	timezoneFallback string
	debugMetadata    bool
	ttl              time.Duration
	now              func() time.Time

	mu        sync.Mutex
	cached    Result
	expiresAt time.Time
	hasCached bool
	refresh   *athleteProfileRefresh
}

type athleteProfileRefresh struct {
	done   chan struct{}
	result Result
	err    error
}

// AthleteProfileResource returns the dynamic cached athlete-profile resource definition.
func AthleteProfileResource(client clients.ProfileClient, opts ResourceOptions) Resource {
	reader := newAthleteProfileReader(athleteProfileOptions{
		client:           client,
		version:          opts.Version,
		timezoneFallback: opts.TimezoneFallback,
		debugMetadata:    opts.DebugMetadata,
		ttl:              opts.AthleteProfileTTL,
		now:              opts.Now,
	})
	return Resource{
		URI:         AthleteProfileURI,
		Name:        "athlete_profile",
		Title:       "Athlete profile",
		Description: "Dynamic cached athlete profile, units, thresholds, zones, and response metadata shaped like get_athlete_profile.",
		MIMEType:    AthleteProfileMIMEType,
		Handler:     reader.Read,
	}
}

func newAthleteProfileReader(opts athleteProfileOptions) *athleteProfileReader {
	ttl := opts.ttl
	if ttl <= 0 {
		ttl = AthleteProfileTTL
	}
	now := opts.now
	if now == nil {
		now = time.Now
	}
	return &athleteProfileReader{
		client:           opts.client,
		version:          athleteprofile.NormalizeVersion(opts.version),
		timezoneFallback: athleteprofile.NormalizeTimezoneFallback(opts.timezoneFallback),
		debugMetadata:    opts.debugMetadata,
		ttl:              ttl,
		now:              now,
	}
}

func (r *athleteProfileReader) Read(ctx context.Context, _ Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	now := r.now()
	r.mu.Lock()
	if r.hasCached && now.Before(r.expiresAt) {
		result := r.cached
		r.mu.Unlock()
		return result, nil
	}
	if r.refresh != nil {
		refresh := r.refresh
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case <-refresh.done:
			return refresh.result, refresh.err
		}
	}
	refresh := &athleteProfileRefresh{done: make(chan struct{})}
	r.refresh = refresh
	r.mu.Unlock()

	result, err := r.refreshProfile(ctx)

	r.mu.Lock()
	if err == nil {
		r.cached = result
		r.expiresAt = now.Add(r.ttl)
		r.hasCached = true
	}
	refresh.result = result
	refresh.err = err
	r.refresh = nil
	close(refresh.done)
	r.mu.Unlock()
	return result, err
}

func (r *athleteProfileReader) refreshProfile(ctx context.Context) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if r.client == nil {
		return Result{}, errors.New("could not fetch athlete profile; check intervals.icu credentials and athlete ID")
	}
	profile, err := r.client.GetAthleteProfile(ctx)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Result{}, ctxErr
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return Result{}, err
		}
		return Result{}, fmt.Errorf("could not fetch athlete profile; check intervals.icu credentials and athlete ID: %w", err)
	}
	shaped, err := athleteprofile.Shape(profile, r.version, r.timezoneFallback, false, r.debugMetadata)
	if err != nil {
		return Result{}, fmt.Errorf("shaping athlete profile resource: %w", err)
	}
	text, err := json.Marshal(shaped)
	if err != nil {
		return Result{}, fmt.Errorf("encoding athlete profile resource: %w", err)
	}
	return Result{URI: AthleteProfileURI, MIMEType: AthleteProfileMIMEType, Text: string(text)}, nil
}
