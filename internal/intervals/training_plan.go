package intervals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// TrainingPlan contains the upstream active training-plan assignment and preserves raw fields.
type TrainingPlan struct {
	Raw map[string]any `json:"-"`

	Active         bool    `json:"-"`
	TrainingPlanID string  `json:"-"`
	PlanID         string  `json:"-"`
	Name           *string `json:"name"`
	Alias          *string `json:"alias"`
	StartDate      *string `json:"training_plan_start_date"`
	Timezone       *string `json:"timezone"`
	LastApplied    *string `json:"last_applied"`
	PlanApplied    *string `json:"plan_applied"`
	TrainingPlan   any     `json:"training_plan"`
	Plan           any     `json:"plan"`
}

// UnmarshalJSON decodes TrainingPlan while retaining the original object for full responses.
func (p *TrainingPlan) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*p = TrainingPlan{}
		return nil
	}
	type trainingPlanAlias TrainingPlan
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded trainingPlanAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = TrainingPlan(decoded)
	p.Raw = raw
	p.TrainingPlanID = rawIDString(firstTrainingPlanRaw(raw, "training_plan_id", "id"))
	p.PlanID = rawIDString(firstTrainingPlanRaw(raw, "plan_id"))
	p.Active = p.TrainingPlanID != "" || p.PlanID != ""
	return nil
}

// GetTrainingPlan retrieves the active training-plan assignment for the configured athlete.
func (c *Client) GetTrainingPlan(ctx context.Context) (TrainingPlan, error) {
	var plan TrainingPlan
	if err := c.doJSON(ctx, &plan, "athlete", c.athleteID, "training-plan"); err != nil {
		if errors.Is(err, ErrNotFound) {
			return TrainingPlan{}, nil
		}
		return TrainingPlan{}, fmt.Errorf("getting training plan: %w", err)
	}
	return plan, nil
}

func firstTrainingPlanRaw(raw map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := raw[key]; ok && value != nil {
			return value
		}
	}
	return nil
}
