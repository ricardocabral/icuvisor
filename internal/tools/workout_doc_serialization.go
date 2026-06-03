package tools

import (
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

func updateWorkoutSerializeOptions(profile intervals.AthleteWithSportSettings, args updateWorkoutRequest) workoutdoc.SerializeOptions {
	if !args.sportProvided {
		return workoutdoc.SerializeOptions{}
	}
	return workoutDocSerializeOptionsForSport(profile, args.Sport)
}

func workoutDocSerializeOptionsForSport(profile intervals.AthleteWithSportSettings, sport string) workoutdoc.SerializeOptions {
	setting, ok := sportSettingsForSport(profile, sport)
	if !ok {
		return workoutdoc.SerializeOptions{}
	}
	order := normalizedWorkoutOrder(setting.WorkoutOrder)
	if order == "" {
		return workoutdoc.SerializeOptions{}
	}
	return workoutdoc.SerializeOptions{WorkoutOrder: order}
}

func sportSettingsForSport(profile intervals.AthleteWithSportSettings, sport string) (intervals.SportSettings, bool) {
	want := normalizedSportName(sport)
	if want == "" {
		return intervals.SportSettings{}, false
	}
	for _, setting := range profile.SportSettings {
		if normalizedSportName(setting.Type) == want {
			return setting, true
		}
		for _, candidate := range setting.Types {
			if normalizedSportName(candidate) == want {
				return setting, true
			}
		}
	}
	return intervals.SportSettings{}, false
}

func normalizedWorkoutOrder(order string) string {
	switch strings.ToUpper(strings.TrimSpace(order)) {
	case "POWER_HR_PACE", "HR_POWER_PACE", "PACE_HR_POWER":
		return strings.ToUpper(strings.TrimSpace(order))
	default:
		return ""
	}
}

func normalizedSportName(sport string) string {
	return strings.ToLower(strings.TrimSpace(sport))
}
