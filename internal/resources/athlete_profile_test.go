package resources

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/athleteprofile"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeAthleteProfileClient struct {
	profiles []intervals.AthleteWithSportSettings
	err      error
	calls    int
}

func (c *fakeAthleteProfileClient) GetAthleteProfile(ctx context.Context) (intervals.AthleteWithSportSettings, error) {
	if err := ctx.Err(); err != nil {
		return intervals.AthleteWithSportSettings{}, err
	}
	c.calls++
	if c.err != nil {
		return intervals.AthleteWithSportSettings{}, c.err
	}
	if len(c.profiles) == 0 {
		return intervals.AthleteWithSportSettings{}, nil
	}
	idx := c.calls - 1
	if idx >= len(c.profiles) {
		idx = len(c.profiles) - 1
	}
	return c.profiles[idx], nil
}

func TestAthleteProfileResourceReturnsSharedShapedProfile(t *testing.T) {
	t.Parallel()

	profile := resourceTestProfile("i12345", "Example Athlete")
	client := &fakeAthleteProfileClient{profiles: []intervals.AthleteWithSportSettings{profile}}
	resource := AthleteProfileResource(client, ResourceOptions{Version: "v0.1-test", TimezoneFallback: "America/Sao_Paulo"})

	result, err := resource.Handler(context.Background(), Request{URI: AthleteProfileURI})
	if err != nil {
		t.Fatalf("resource handler error = %v", err)
	}
	wantShaped, err := athleteprofile.Shape(profile, "v0.1-test", "America/Sao_Paulo", false, false)
	if err != nil {
		t.Fatalf("athleteprofile.Shape() error = %v", err)
	}
	wantText, err := json.Marshal(wantShaped)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if result.URI != AthleteProfileURI || result.MIMEType != AthleteProfileMIMEType || result.Text != string(wantText) {
		t.Fatalf("resource result = %#v, want URI/MIME/shared shaped JSON %s", result, wantText)
	}
}

func TestAthleteProfileResourceCachesUntilTTLExpires(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	client := &fakeAthleteProfileClient{profiles: []intervals.AthleteWithSportSettings{
		resourceTestProfile("i12345", "Cached Athlete"),
		resourceTestProfile("i12345", "Refreshed Athlete"),
	}}
	resource := AthleteProfileResource(client, ResourceOptions{Version: "test", TimezoneFallback: "UTC", AthleteProfileTTL: time.Minute, Now: func() time.Time { return now }})

	first, err := resource.Handler(context.Background(), Request{URI: AthleteProfileURI})
	if err != nil {
		t.Fatalf("first read error = %v", err)
	}
	now = now.Add(30 * time.Second)
	second, err := resource.Handler(context.Background(), Request{URI: AthleteProfileURI})
	if err != nil {
		t.Fatalf("second read error = %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("client calls after cached read = %d, want 1", client.calls)
	}
	if second.Text != first.Text || !strings.Contains(second.Text, "Cached Athlete") {
		t.Fatalf("cached read text = %s, want first cached payload", second.Text)
	}
	now = now.Add(31 * time.Second)
	third, err := resource.Handler(context.Background(), Request{URI: AthleteProfileURI})
	if err != nil {
		t.Fatalf("third read error = %v", err)
	}
	if client.calls != 2 {
		t.Fatalf("client calls after expired read = %d, want 2", client.calls)
	}
	if !strings.Contains(third.Text, "Refreshed Athlete") {
		t.Fatalf("refreshed read text = %s, want refreshed athlete", third.Text)
	}
}

func TestNewRegistryWithOptionsRegistersAthleteProfileResource(t *testing.T) {
	t.Parallel()

	client := &fakeAthleteProfileClient{profiles: []intervals.AthleteWithSportSettings{resourceTestProfile("i12345", "Example Athlete")}}
	registrar := &captureRegistrar{}
	if err := NewRegistryWithOptions(client, ResourceOptions{Version: "v0.1-test", TimezoneFallback: "UTC"}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	var resource Resource
	for _, candidate := range registrar.resources {
		if candidate.URI == AthleteProfileURI {
			resource = candidate
			break
		}
	}
	if resource.URI == "" {
		t.Fatalf("registered resources = %#v, missing %s", registrar.resources, AthleteProfileURI)
	}
	if resource.Name != "athlete_profile" || resource.Title != "Athlete profile" || resource.MIMEType != AthleteProfileMIMEType {
		t.Fatalf("resource metadata = %#v, want athlete profile metadata", resource)
	}
}

func TestAthleteProfileResourceHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &fakeAthleteProfileClient{profiles: []intervals.AthleteWithSportSettings{resourceTestProfile("i12345", "Example Athlete")}}
	_, err := AthleteProfileResource(client, ResourceOptions{}).Handler(ctx, Request{URI: AthleteProfileURI})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("handler error = %v, want context.Canceled", err)
	}
	if client.calls != 0 {
		t.Fatalf("client calls = %d, want 0", client.calls)
	}
}

func TestAthleteProfileResourceCanceledReadDoesNotWaitForInFlightRefresh(t *testing.T) {
	client := &blockingAthleteProfileClient{
		started: make(chan struct{}),
		release: make(chan struct{}),
		profile: resourceTestProfile("i12345", "Blocked Athlete"),
	}
	resource := AthleteProfileResource(client, ResourceOptions{Version: "test", TimezoneFallback: "UTC"})
	firstDone := make(chan error, 1)
	go func() {
		_, err := resource.Handler(context.Background(), Request{URI: AthleteProfileURI})
		firstDone <- err
	}()

	select {
	case <-client.started:
	case <-time.After(time.Second):
		t.Fatal("first refresh did not start")
	}

	ctx, cancel := context.WithCancel(context.Background())
	secondDone := make(chan error, 1)
	go func() {
		_, err := resource.Handler(ctx, Request{URI: AthleteProfileURI})
		secondDone <- err
	}()
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-secondDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("second read error = %v, want context.Canceled", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("second canceled read waited for in-flight refresh")
	}

	close(client.release)
	select {
	case err := <-firstDone:
		if err != nil {
			t.Fatalf("first read error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("first refresh did not finish")
	}
	if client.calls != 1 {
		t.Fatalf("client calls = %d, want one shared in-flight refresh", client.calls)
	}
}

type blockingAthleteProfileClient struct {
	started chan struct{}
	release chan struct{}
	profile intervals.AthleteWithSportSettings
	calls   int
}

func (c *blockingAthleteProfileClient) GetAthleteProfile(ctx context.Context) (intervals.AthleteWithSportSettings, error) {
	c.calls++
	close(c.started)
	select {
	case <-ctx.Done():
		return intervals.AthleteWithSportSettings{}, ctx.Err()
	case <-c.release:
		return c.profile, nil
	}
}

func resourceTestProfile(id string, name string) intervals.AthleteWithSportSettings {
	return intervals.AthleteWithSportSettings{
		ID:             id,
		Name:           name,
		PreferredUnits: "metric",
		Timezone:       "Europe/Lisbon",
		SportSettings: []intervals.SportSettings{{
			Types:          []string{"Ride"},
			FTP:            250,
			LTHR:           170,
			PowerZones:     []int{100, 150, 200},
			PowerZoneNames: []string{"Z1", "Z2", "Z3"},
		}},
	}
}
