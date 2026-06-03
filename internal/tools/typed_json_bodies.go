package tools

import "encoding/json"

type jsonObject = map[string]any

type workoutDocSummaryRow struct {
	Present        bool                      `json:"present"`
	StepCount      *int                      `json:"step_count,omitempty"`
	Name           string                    `json:"name,omitempty"`
	TopLevelKeys   *[]string                 `json:"top_level_keys,omitempty"`
	TargetPreviews []workoutTargetPreviewRow `json:"target_previews,omitempty"`
}

type trainingPlanSummaryRow struct {
	ID           string   `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	FolderID     string   `json:"folder_id,omitempty"`
	Type         string   `json:"type,omitempty"`
	Category     string   `json:"category,omitempty"`
	ChildCount   *int     `json:"child_count,omitempty"`
	WorkoutCount *int     `json:"workout_count,omitempty"`
	TopLevelKeys []string `json:"top_level_keys,omitempty"`
}

func rawJSONMap(in map[string]any) json.RawMessage {
	if in == nil {
		return nil
	}
	encoded, err := json.Marshal(cloneJSONMap(in))
	if err != nil {
		return nil
	}
	return json.RawMessage(encoded)
}
