package intervals

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// DefaultWorkoutLibrarySearchLimit bounds default public searches to terse folder summaries.
const DefaultWorkoutLibrarySearchLimit = 3

// ErrSharedWorkoutFolderUnverified means sharing did not confirm the requested folder in the target library.
var ErrSharedWorkoutFolderUnverified = errors.New("shared workout folder not verified in target library")

// WorkoutLibrarySearchParams contains public library filters, with weeks and hours per week as integers.
type WorkoutLibrarySearchParams struct {
	Query           string
	Type            string
	ActivityTypes   []string
	WeeksMin        int
	WeeksMax        int
	HoursPerWeekMin int
	HoursPerWeekMax int
	WorkoutTargets  []string
	Limit           int
}

// WorkoutLibraryActivityTypes returns the public folder-search sport enum.
func WorkoutLibraryActivityTypes() []string {
	return []string{"Ride", "Run", "Swim", "WeightTraining", "Hike", "Walk", "AlpineSki", "BackcountrySki", "Badminton", "Canoeing", "Crossfit", "EBikeRide", "EMountainBikeRide", "Elliptical", "Golf", "GravelRide", "TrackRide", "Handcycle", "HighIntensityIntervalTraining", "Hockey", "IceSkate", "InlineSkate", "Kayaking", "Kitesurf", "MountainBikeRide", "Cyclocross", "NordicSki", "OpenWaterSwim", "Padel", "Pilates", "Pickleball", "Racquetball", "Rugby", "RockClimbing", "RollerSki", "Rowing", "Sail", "Skateboard", "Snowboard", "Snowshoe", "Soccer", "Squash", "StairStepper", "StandUpPaddling", "Surfing", "TableTennis", "Tennis", "TrailRun", "Transition", "Velomobile", "VirtualRide", "VirtualRow", "VirtualRun", "VirtualSki", "WaterSport", "Wheelchair", "Windsurf", "Workout", "Yoga", "Other"}
}

// Validate checks icuvisor's bounded filters before making a public search request.
func (p WorkoutLibrarySearchParams) Validate() error {
	if n := utf8.RuneCountInString(strings.TrimSpace(p.Query)); n == 0 || n > 200 {
		return errors.New("q must contain 1 to 200 characters")
	}
	if p.Type != "" && p.Type != "FOLDER" && p.Type != "PLAN" {
		return errors.New("type must be FOLDER or PLAN")
	}
	if p.WeeksMin < 0 || p.WeeksMax < p.WeeksMin || p.WeeksMax > 520 {
		return errors.New("weeks must be ordered between 0 and 520")
	}
	if p.HoursPerWeekMin < 0 || p.HoursPerWeekMax < p.HoursPerWeekMin || p.HoursPerWeekMax > 168 {
		return errors.New("hours per week must be ordered between 0 and 168")
	}
	if p.Limit < 0 || p.Limit > 100 {
		return errors.New("limit must be between 1 and 100, or omitted")
	}
	if len(p.ActivityTypes) > 10 {
		return errors.New("at most 10 activity types are supported")
	}
	for _, sport := range p.ActivityTypes {
		if !slices.Contains(WorkoutLibraryActivityTypes(), sport) {
			return errors.New("activity_types contains an unsupported sport")
		}
	}
	if len(p.WorkoutTargets) > 4 {
		return errors.New("at most 4 workout targets are supported")
	}
	for _, target := range p.WorkoutTargets {
		if !slices.Contains([]string{"AUTO", "POWER", "HR", "PACE"}, target) {
			return errors.New("workout_targets must contain AUTO, POWER, HR or PACE")
		}
	}
	return nil
}

// SearchWorkoutLibrary searches globally published folders without changing any athlete's library.
func (c *Client) SearchWorkoutLibrary(ctx context.Context, params WorkoutLibrarySearchParams) ([]WorkoutFolder, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("searching workout library: %w", err)
	}
	limit := params.Limit
	if limit == 0 {
		limit = DefaultWorkoutLibrarySearchLimit
	}
	query := url.Values{"q": {strings.TrimSpace(params.Query)}, "weeks_min": {strconv.Itoa(params.WeeksMin)}, "weeks_max": {strconv.Itoa(params.WeeksMax)}, "hours_per_week_min": {strconv.Itoa(params.HoursPerWeekMin)}, "hours_per_week_max": {strconv.Itoa(params.HoursPerWeekMax)}, "limit": {strconv.Itoa(limit)}}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	for _, sport := range params.ActivityTypes {
		query.Add("activity_types", sport)
	}
	for _, target := range params.WorkoutTargets {
		query.Add("workout_targets", target)
	}
	var folders []WorkoutFolder
	if err := c.doJSONQuery(ctx, &folders, query, "folder-search"); err != nil {
		return nil, fmt.Errorf("searching workout library: %w", err)
	}
	return folders, nil
}

// NormalizeWorkoutFolderID validates the public API's positive int32 folder ID.
func NormalizeWorkoutFolderID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Trim(value, "0123456789") != "" {
		return "", errors.New("folder_id must contain digits only")
	}
	id, err := strconv.ParseInt(value, 10, 32)
	if err != nil || id <= 0 {
		return "", errors.New("folder_id must be a positive 32-bit integer string")
	}
	return strconv.FormatInt(id, 10), nil
}

// AddSharedWorkoutFolder explicitly shares a public folder into the selected athlete's library and verifies membership.
// Public folders retain their owner's athlete_id; ownership is not the selected athlete's authorization identity.
// POST is not retried. After an interrupted or unverified write, inspect the library before retrying.
func (c *Client) AddSharedWorkoutFolder(ctx context.Context, folderID string) (WorkoutFolder, error) {
	id, err := NormalizeWorkoutFolderID(folderID)
	if err != nil {
		return WorkoutFolder{}, err
	}
	var folders []WorkoutFolder
	if err := c.doJSONNoBody(ctx, http.MethodPost, &folders, "athlete", c.athleteID, "folders", id, "share-with-me"); err != nil {
		return WorkoutFolder{}, fmt.Errorf("adding shared workout folder: %w", err)
	}
	if !slices.ContainsFunc(folders, func(folder WorkoutFolder) bool { return folder.ID == id }) {
		return WorkoutFolder{}, ErrSharedWorkoutFolderUnverified
	}
	verified, err := c.ListWorkoutFolders(ctx)
	if err != nil {
		return WorkoutFolder{}, fmt.Errorf("verifying shared workout folder; inspect library before retrying: %w", err)
	}
	for _, folder := range verified {
		if folder.ID == id {
			return folder, nil
		}
	}
	return WorkoutFolder{}, ErrSharedWorkoutFolderUnverified
}
