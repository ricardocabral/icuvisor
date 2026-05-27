package tools

import (
	"fmt"
	"math"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

func shapeGetActivitiesResponse(activities []intervals.Activity, gearResolutions map[string]activityGearResolution, args GetActivitiesRequest, nextToken string, version string, timezoneFallback string, debugMetadata bool, unitSystem response.UnitSystem, customFieldCodes []string, asOfMeta *response.AsOfMetadata, shaping ...responseShaping) (any, error) {
	rows := make([]getActivitiesRow, 0, len(activities))
	for _, activity := range activities {
		rows = append(rows, activityRow(activity, args.IncludeFull, timezoneFallback, unitSystem, gearResolutions[activity.ID], customFieldCodes))
	}
	meta := getActivitiesMeta{PageSize: args.PageSize, NextPageToken: nextToken, MoreAvailable: nextToken != "", IncludeFull: args.IncludeFull, Timezone: timezoneFallback, FieldSemantics: activityFieldSemantics(rows)}
	if asOfMeta != nil {
		meta.AsOf = asOfMeta.AsOf
		meta.AsOfDate = asOfMeta.AsOfDate
		meta.AsOfWeekday = asOfMeta.AsOfWeekday
		meta.Timezone = asOfMeta.Timezone
	}
	payload := getActivitiesResponse{Activities: rows, Meta: meta}
	shapeCfg := responseShapingOrDefault(shaping)
	return response.Shape(payload, shapeCfg.options(args.IncludeFull, []string{"activities"}, version, debugMetadata, getActivitiesName, unitSystem))
}

func activityRow(activity intervals.Activity, includeFull bool, timezoneFallback string, unitSystem response.UnitSystem, gearResolution activityGearResolution, customFieldCodes []string) getActivitiesRow {
	row := getActivitiesRow{ActivityID: activity.ID, Name: strings.TrimSpace(stringValue(activity.Name)), Sport: stringValue(activity.Type), SubType: stringValue(activity.SubType), StartDateLocal: stringValue(activity.StartDateLocal), StartDateUTC: stringValue(activity.StartDate), Timezone: firstNonEmpty(stringValue(activity.Timezone), timezoneFallback), Tags: rawStringArray(activity.Raw, "tags")}
	applyActivityGearResolution(&row, gearResolution)
	if isStravaBlocked(activity) {
		row.StravaImported = true
		row.Unavailable = &unavailableReason{Reason: "strava_blocked", Workaround: stravaBlockedWorkaround(activity.Raw)}
		if includeFull {
			row.Full = activity.Raw
		}
		return row
	}
	row.MovingTimeSeconds = intValue(activity.MovingTime)
	row.ElapsedTimeSeconds = intValue(activity.ElapsedTime)
	row.TrainingLoad = intValue(activity.TrainingLoad)
	row.AverageHeartRateBPM = intValue(activity.AverageHeartRate)
	row.MaxHeartRateBPM = intValue(activity.MaxHeartRate)
	row.AverageCadenceRPM = activity.AverageCadence
	row.CaloriesBurned = activity.Calories
	row.CarbsIngestedG = activity.CarbsIngested
	row.CarbsUsedG = activity.CarbsUsed
	row.DeviceName = stringValue(activity.DeviceName)
	row.HasStreams = len(activity.StreamTypes) > 0
	if activity.TotalElevationGain != nil {
		value := *activity.TotalElevationGain
		row.ElevationGainM = &value
	}
	if activity.TotalElevationLoss != nil {
		value := *activity.TotalElevationLoss
		row.ElevationLossM = &value
	}
	applyActivityDistanceAndPace(&row, activity, unitSystem)
	applyActivitySpeed(&row, activity.AverageSpeed, true, unitSystem)
	applyActivitySpeed(&row, activity.MaxSpeed, false, unitSystem)
	row.CustomFields = activityCustomFields(activity.Raw, customFieldCodes)
	if includeFull {
		row.Full = activity.Raw
	}
	return row
}

// activityCustomFields surfaces athlete-defined activity custom field values,
// keyed by their upstream field code. Only non-null scalar values are returned.
func activityCustomFields(raw map[string]any, customFieldCodes []string) map[string]any {
	if len(raw) == 0 || len(customFieldCodes) == 0 {
		return nil
	}
	fields := make(map[string]any)
	for _, code := range customFieldCodes {
		value, ok := raw[code]
		if !ok || !isCustomIntervalFieldValue(value) {
			continue
		}
		fields[code] = value
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func applyActivityDistanceAndPace(row *getActivitiesRow, activity intervals.Activity, unitSystem response.UnitSystem) {
	distanceMeters := firstFloat(activity.ICUDistance, activity.Distance)
	if distanceMeters == nil || *distanceMeters <= 0 {
		return
	}
	converted := response.ToPreferred(*distanceMeters, units.UnitM, unitSystem)
	value := round(converted.Value, 3)
	if converted.Unit == units.UnitMI {
		row.DistanceMI = &value
	} else {
		row.DistanceKM = &value
	}
	if row.MovingTimeSeconds > 0 && isRunLikeActivity(activity) {
		pace := float64(row.MovingTimeSeconds) / converted.Value
		pace = round(pace, 1)
		if converted.Unit == units.UnitMI {
			row.PaceSecondsPerMile = &pace
		} else {
			row.PaceSecondsPerKM = &pace
		}
	}
}

func applyActivityGearResolution(row *getActivitiesRow, resolution activityGearResolution) {
	if strings.TrimSpace(resolution.GearID) == "" {
		return
	}
	row.GearID = resolution.GearID
	row.GearName = resolution.Name
	row.GearResolution = resolution.Status
}

func applyActivitySpeed(row *getActivitiesRow, speed *float64, average bool, unitSystem response.UnitSystem) {
	if speed == nil || *speed <= 0 {
		return
	}
	converted := response.ToPreferred(*speed, units.UnitMS, unitSystem)
	value := round(converted.Value, 3)
	if average {
		if converted.Unit == units.UnitMPH {
			row.AverageSpeedMPH = &value
		} else {
			row.AverageSpeedKMH = &value
		}
		return
	}
	if converted.Unit == units.UnitMPH {
		row.MaxSpeedMPH = &value
	} else {
		row.MaxSpeedKMH = &value
	}
}

func isRunLikeActivity(activity intervals.Activity) bool {
	sport := strings.ToLower(strings.TrimSpace(stringValue(activity.Type) + " " + stringValue(activity.SubType)))
	return strings.Contains(sport, "run") || strings.Contains(sport, "jog")
}

func firstFloat(values ...*float64) *float64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func activityFieldSemantics(rows []getActivitiesRow) map[string]string {
	semantics := map[string]string{}
	for _, row := range rows {
		if row.CaloriesBurned != nil {
			semantics["calories_burned"] = "Active/exercise calories burned from upstream activity calories. Distinct from wellness kcal_consumed (dietary intake)."
		}
		if row.CarbsIngestedG != nil {
			semantics["carbs_ingested_g"] = "Carbohydrates consumed/ingested during the activity in grams from upstream carbs_ingested."
		}
		if row.CarbsUsedG != nil {
			semantics["carbs_used_g"] = "Upstream estimate of carbohydrates used/burned during the activity in grams from upstream carbs_used."
		}
	}
	if len(semantics) == 0 {
		return nil
	}
	return semantics
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func anyString(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func round(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}
