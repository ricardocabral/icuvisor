package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WorkoutFolder contains a workout-library folder or plan and preserves raw upstream fields.
type WorkoutFolder struct {
	Raw rawJSONObject `json:"-"`

	ID              string    `json:"-"`
	AthleteID       *string   `json:"athlete_id"`
	Type            *string   `json:"type"`
	Name            *string   `json:"name"`
	Description     *string   `json:"description"`
	Children        []Workout `json:"children"`
	Visibility      *string   `json:"visibility"`
	NumWorkouts     *int      `json:"num_workouts"`
	ActivityTypes   []string  `json:"activity_types"`
	DurationWeeks   *int      `json:"duration_weeks"`
	HoursPerWeekMin *int      `json:"hours_per_week_min"`
	HoursPerWeekMax *int      `json:"hours_per_week_max"`
	WorkoutTargets  []string  `json:"workout_targets"`
	Blurb           *string   `json:"blurb"`
}

// UnmarshalJSON decodes WorkoutFolder while retaining the original object for full responses.
func (f *WorkoutFolder) UnmarshalJSON(data []byte) error {
	type folderAlias WorkoutFolder
	var raw rawJSONObject
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded folderAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*f = WorkoutFolder(decoded)
	f.Raw = raw
	f.ID = rawIDString(raw["id"])
	return nil
}

// WriteWorkoutParams contains writable workout-library template fields.
type WriteWorkoutParams struct {
	WorkoutID      string
	Name           string
	NameSet        bool
	FolderID       string
	FolderIDSet    bool
	Description    *string
	DescriptionSet bool
	Tags           []string
	TagsSet        bool
	Sport          string
	SportSet       bool
}

type workoutLibraryWriteRequest struct {
	Description *string   `json:"description,omitempty"`
	FolderID    *string   `json:"folder_id,omitempty"`
	Name        *string   `json:"name,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
	Sport       *string   `json:"type,omitempty"`
}

// Workout contains a workout-library template and preserves raw upstream fields including workout_doc.
type Workout struct {
	Raw rawJSONObject `json:"-"`

	ID              string   `json:"-"`
	AthleteID       *string  `json:"athlete_id"`
	Name            *string  `json:"name"`
	Description     *string  `json:"description"`
	Type            *string  `json:"type"`
	Indoor          *bool    `json:"indoor"`
	TrainingLoad    *int     `json:"icu_training_load"`
	MovingTime      *int     `json:"moving_time"`
	Updated         *string  `json:"updated"`
	WorkoutDoc      any      `json:"workout_doc"`
	FolderID        any      `json:"folder_id"`
	Target          *string  `json:"target"`
	Targets         []string `json:"targets"`
	Tags            []string `json:"tags"`
	Distance        *float64 `json:"distance"`
	Intensity       *float64 `json:"icu_intensity"`
	CarbsPerHour    *int     `json:"carbs_per_hour"`
	Joules          *int     `json:"joules"`
	JoulesAboveFTP  *int     `json:"joules_above_ftp"`
	HideFromAthlete *bool    `json:"hide_from_athlete"`
	Day             *int     `json:"day"`
	Days            *int     `json:"days"`
	PlanApplied     *string  `json:"plan_applied"`
	Time            *string  `json:"time"`
	Subtype         *string  `json:"sub_type"`
	ForWeek         *bool    `json:"for_week"`
}

// UnmarshalJSON decodes Workout while retaining the original object for full responses.
func (w *Workout) UnmarshalJSON(data []byte) error {
	type workoutAlias Workout
	var raw rawJSONObject
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded workoutAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*w = Workout(decoded)
	w.Raw = raw
	w.ID = rawIDString(raw["id"])
	return nil
}

// ListWorkoutFolders lists workout-library folders, plans, and nested children for the configured athlete.
func (c *Client) ListWorkoutFolders(ctx context.Context) ([]WorkoutFolder, error) {
	var folders []WorkoutFolder
	if err := c.doJSON(ctx, &folders, "athlete", c.athleteID, "folders"); err != nil {
		return nil, fmt.Errorf("listing workout folders: %w", err)
	}
	return folders, nil
}

// ListLibraryWorkouts lists all workout templates in the configured athlete's workout library.
func (c *Client) ListLibraryWorkouts(ctx context.Context) ([]Workout, error) {
	var workouts []Workout
	if err := c.doJSON(ctx, &workouts, "athlete", c.athleteID, "workouts"); err != nil {
		return nil, fmt.Errorf("listing library workouts: %w", err)
	}
	return workouts, nil
}

// CreateLibraryWorkout creates a workout-library template for the configured athlete.
func (c *Client) CreateLibraryWorkout(ctx context.Context, params WriteWorkoutParams) (Workout, error) {
	body, err := writeWorkoutBody(params, false)
	if err != nil {
		return Workout{}, err
	}
	var workout Workout
	if err := c.doJSONBody(ctx, http.MethodPost, body, &workout, "athlete", c.athleteID, "workouts"); err != nil {
		return Workout{}, fmt.Errorf("creating library workout: %w", err)
	}
	return workout, nil
}

// UpdateLibraryWorkout sparsely updates a workout-library template for the configured athlete.
func (c *Client) UpdateLibraryWorkout(ctx context.Context, params WriteWorkoutParams) (Workout, error) {
	workoutID := strings.TrimSpace(params.WorkoutID)
	if workoutID == "" {
		return Workout{}, fmt.Errorf("updating library workout: workout ID is required")
	}
	body, err := writeWorkoutBody(params, true)
	if err != nil {
		return Workout{}, err
	}
	var workout Workout
	if err := c.doJSONBody(ctx, http.MethodPut, body, &workout, "athlete", c.athleteID, "workouts", workoutID); err != nil {
		return Workout{}, fmt.Errorf("updating library workout %s: %w", workoutID, err)
	}
	return workout, nil
}

// DeleteLibraryWorkout deletes a workout-library template for the configured athlete.
func (c *Client) DeleteLibraryWorkout(ctx context.Context, workoutID string) error {
	workoutID = strings.TrimSpace(workoutID)
	if workoutID == "" {
		return fmt.Errorf("deleting library workout: workout ID is required")
	}
	if err := c.doNoJSON(ctx, "athlete", c.athleteID, "workouts", workoutID); err != nil {
		return fmt.Errorf("deleting library workout %s: %w", workoutID, err)
	}
	return nil
}

func (c *Client) doNoJSON(ctx context.Context, pathParts ...string) error {
	for attempt := 1; ; attempt++ {
		req, err := c.newRequest(ctx, http.MethodDelete, pathParts...)
		if err != nil {
			return err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			if retry, wait := c.decideRetry(ctx, http.MethodDelete, nil, err, attempt); retry {
				if sleepErr := c.sleepBeforeRetry(ctx, wait); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			return fmt.Errorf("calling intervals.icu: %w", err)
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
			apiErr := errorForStatus(resp.StatusCode, retryAfter)
			retry, wait := c.decideRetry(ctx, http.MethodDelete, resp, nil, attempt)
			_, _ = io.Copy(io.Discard, resp.Body)
			closeErr := resp.Body.Close()
			if retry {
				if sleepErr := c.sleepBeforeRetry(ctx, wait); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			if closeErr != nil {
				return fmt.Errorf("closing intervals.icu response: %w", closeErr)
			}
			return fmt.Errorf("calling intervals.icu: %w", apiErr)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("closing intervals.icu response: %w", err)
		}
		return nil
	}
}

func writeWorkoutBody(params WriteWorkoutParams, allowSparse bool) (workoutLibraryWriteRequest, error) {
	var body workoutLibraryWriteRequest
	fieldsSet := 0
	name := strings.TrimSpace(params.Name)
	sport := strings.TrimSpace(params.Sport)
	if !allowSparse {
		folderID := strings.TrimSpace(params.FolderID)
		if name == "" {
			return workoutLibraryWriteRequest{}, fmt.Errorf("writing workout: name is required")
		}
		if sport == "" {
			return workoutLibraryWriteRequest{}, fmt.Errorf("writing workout: sport is required")
		}
		if folderID == "" {
			return workoutLibraryWriteRequest{}, fmt.Errorf("writing workout: folder ID is required")
		}
		body.Name = &name
		body.Sport = &sport
		body.FolderID = &folderID
		if params.Description != nil {
			body.Description = params.Description
		}
		if len(params.Tags) > 0 {
			tags := append([]string(nil), params.Tags...)
			body.Tags = &tags
		}
		return body, nil
	}
	if params.NameSet {
		if name == "" {
			return workoutLibraryWriteRequest{}, fmt.Errorf("writing workout: name cannot be empty")
		}
		body.Name = &name
		fieldsSet++
	}
	if params.FolderIDSet {
		folderID := strings.TrimSpace(params.FolderID)
		body.FolderID = &folderID
		fieldsSet++
	}
	if params.DescriptionSet {
		if params.Description == nil {
			return workoutLibraryWriteRequest{}, fmt.Errorf("writing workout: description cannot be null")
		}
		body.Description = params.Description
		fieldsSet++
	}
	if params.TagsSet {
		tags := append([]string{}, params.Tags...)
		body.Tags = &tags
		fieldsSet++
	}
	if params.SportSet {
		if sport == "" {
			return workoutLibraryWriteRequest{}, fmt.Errorf("writing workout: sport cannot be empty")
		}
		body.Sport = &sport
		fieldsSet++
	}
	if fieldsSet == 0 {
		return workoutLibraryWriteRequest{}, fmt.Errorf("writing workout: at least one field is required")
	}
	return body, nil
}
