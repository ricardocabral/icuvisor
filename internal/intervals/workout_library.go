package intervals

import (
	"context"
	"encoding/json"
	"fmt"
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
