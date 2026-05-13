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
	Raw map[string]any `json:"-"`

	ID          string    `json:"-"`
	AthleteID   *string   `json:"athlete_id"`
	Type        *string   `json:"type"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Children    []Workout `json:"children"`
	Visibility  *string   `json:"visibility"`
	NumWorkouts *int      `json:"num_workouts"`
}

// UnmarshalJSON decodes WorkoutFolder while retaining the original object for full responses.
func (f *WorkoutFolder) UnmarshalJSON(data []byte) error {
	type folderAlias WorkoutFolder
	var raw map[string]any
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

// Workout contains a workout-library template and preserves raw upstream fields including workout_doc.
type Workout struct {
	Raw map[string]any `json:"-"`

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
	var raw map[string]any
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
	if err := c.doNoJSON(ctx, http.MethodDelete, "athlete", c.athleteID, "workouts", workoutID); err != nil {
		return fmt.Errorf("deleting library workout %s: %w", workoutID, err)
	}
	return nil
}

func (c *Client) doNoJSON(ctx context.Context, method string, pathParts ...string) error {
	for attempt := 1; ; attempt++ {
		req, err := c.newRequest(ctx, method, pathParts...)
		if err != nil {
			return err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			if c.shouldRetryTransport(ctx, attempt) {
				if sleepErr := c.sleepBeforeRetry(ctx, attempt, 0); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			return fmt.Errorf("calling intervals.icu: %w", err)
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
			apiErr := errorForStatus(resp.StatusCode, retryAfter)
			_, _ = io.Copy(io.Discard, resp.Body)
			closeErr := resp.Body.Close()
			if c.shouldRetryStatus(resp.StatusCode, attempt) {
				if sleepErr := c.sleepBeforeRetry(ctx, attempt, retryAfter); sleepErr != nil {
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

func writeWorkoutBody(params WriteWorkoutParams, allowSparse bool) (map[string]any, error) {
	body := map[string]any{}
	name := strings.TrimSpace(params.Name)
	sport := strings.TrimSpace(params.Sport)
	if !allowSparse {
		if name == "" {
			return nil, fmt.Errorf("writing workout: name is required")
		}
		if sport == "" {
			return nil, fmt.Errorf("writing workout: sport is required")
		}
		body["name"] = name
		body["type"] = sport
		if folderID := strings.TrimSpace(params.FolderID); folderID != "" {
			body["folder_id"] = folderID
		}
		if params.Description != nil {
			body["description"] = *params.Description
		}
		if len(params.Tags) > 0 {
			body["tags"] = append([]string(nil), params.Tags...)
		}
		return body, nil
	}
	if params.NameSet {
		if name == "" {
			return nil, fmt.Errorf("writing workout: name cannot be empty")
		}
		body["name"] = name
	}
	if params.FolderIDSet {
		body["folder_id"] = strings.TrimSpace(params.FolderID)
	}
	if params.DescriptionSet {
		if params.Description == nil {
			return nil, fmt.Errorf("writing workout: description cannot be null")
		}
		body["description"] = *params.Description
	}
	if params.TagsSet {
		body["tags"] = append([]string(nil), params.Tags...)
	}
	if params.SportSet {
		if sport == "" {
			return nil, fmt.Errorf("writing workout: sport cannot be empty")
		}
		body["type"] = sport
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("writing workout: at least one field is required")
	}
	return body, nil
}
