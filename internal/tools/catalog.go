package tools

import (
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

// ToolDescriptor describes one registered MCP tool for generated documentation.
type ToolDescriptor struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Tier    string `json:"tier"`
	Safety  string `json:"safety"`
	Summary string `json:"summary"`
	Anchor  string `json:"anchor"`
}

// Catalog returns the registered tool catalog metadata used by generated documentation.
func Catalog() []ToolDescriptor {
	tools := catalogTools()
	descriptors := make([]ToolDescriptor, 0, len(tools))
	for _, tool := range tools {
		descriptors = append(descriptors, ToolDescriptor{
			Name:    tool.Name,
			Group:   toolCatalogGroup(tool.Name),
			Tier:    tool.EffectiveToolset().String(),
			Safety:  string(tool.Requirement.effective()),
			Summary: toolSummary(tool.Description),
			Anchor:  tool.Name,
		})
	}
	sortToolDescriptors(descriptors)
	return descriptors
}

type registryToolOptions struct {
	version          string
	timezoneFallback string
	debugMetadata    bool
	capability       safety.Capability
	shaping          responseShaping
	gearCache        *gearListCache
	coachModeEnabled bool
	coachConfig      coach.Config
}

func catalogTools() []Tool {
	var client *intervals.Client
	shaping := responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetFull}
	tools := registryBaseTools(client, registryToolOptions{
		version:          "catalog",
		timezoneFallback: "UTC",
		capability:       safety.NewCapability(safety.ModeFull),
		shaping:          shaping,
		gearCache:        newGearListCache(),
		coachModeEnabled: true,
	})
	tools = append(tools, newListAdvancedCapabilitiesTool(tools, safety.ToolsetFull, shaping))
	return tools
}

func registryBaseTools(client *intervals.Client, opts registryToolOptions) []Tool {
	capability := capabilityOrSafe(opts.capability)
	tools := make([]Tool, 0, 39)
	tools = append(tools,
		newGetAthleteProfileTool(client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetFitnessTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetTrainingSummaryTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetWellnessDataTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newUpdateWellnessTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newUpdateSportSettingsTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, capability, opts.shaping),
		newGetBestEffortsTool(client, opts.version, opts.debugMetadata, opts.shaping),
		newGetPowerCurvesTool(client, opts.version, opts.debugMetadata, opts.shaping),
		newGetActivitiesToolWithGear(client, client, client, opts.gearCache, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetEventsTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetEventByIDTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newAddOrUpdateEventTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newApplyTrainingPlanTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, capability, opts.shaping),
		newDeleteEventTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newDeleteEventsByDateRangeTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newLinkActivityToEventTool(client, client, client, opts.version, opts.debugMetadata, opts.shaping),
		newGetTrainingPlanTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetWorkoutLibraryTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetWorkoutsInFolderTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newCreateWorkoutTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newUpdateWorkoutTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newDeleteWorkoutTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newDeleteSportSettingsTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetCustomItemsTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetCustomItemByIDTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newCreateCustomItemTool(client, client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newUpdateCustomItemTool(client, client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newDeleteCustomItemTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetActivityDetailsToolWithGear(client, client, client, opts.gearCache, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newDeleteActivityTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetActivityIntervalsTool(client, client, opts.version, opts.debugMetadata, opts.shaping),
		newGetActivityStreamsTool(client, client, opts.version, opts.debugMetadata, opts.shaping),
		newGetActivitySplitsTool(client, client, client, client, opts.version, opts.debugMetadata, opts.shaping),
		newGetActivityMessagesTool(client, client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newAddActivityMessageTool(client, client, opts.version, opts.debugMetadata, opts.shaping),
		newGetExtendedMetricsTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
		newGetGearListTool(client, opts.gearCache, opts.version, opts.debugMetadata, opts.shaping),
		newDeleteGearTool(client, client, opts.version, opts.timezoneFallback, opts.debugMetadata, opts.shaping),
	)
	if opts.coachModeEnabled {
		tools = append(tools, newListAthletesTool(opts.coachConfig), newSelectAthleteTool(opts.coachConfig))
	}
	return tools
}

func sortToolDescriptors(descriptors []ToolDescriptor) {
	sort.SliceStable(descriptors, func(i, j int) bool {
		if descriptors[i].Group != descriptors[j].Group {
			return descriptors[i].Group < descriptors[j].Group
		}
		return descriptors[i].Name < descriptors[j].Name
	})
}

func toolCatalogGroup(name string) string {
	switch name {
	case getAthleteProfileName, updateSportSettingsName, deleteSportSettingsName, getGearListName, deleteGearName:
		return "settings"
	case getFitnessName, getTrainingSummaryName, getBestEffortsName, getPowerCurvesName:
		return "fitness"
	case getWellnessDataName, updateWellnessName:
		return "wellness"
	case getActivitiesName, getActivityDetailsName, getActivityIntervalsName, getActivityStreamsName, getActivitySplitsName, getActivityMessagesName, addActivityMessageName, getExtendedMetricsName, deleteActivityName:
		return "activities"
	case getEventsName, getEventByIDName, addOrUpdateEventName, deleteEventName, deleteEventsByDateRangeName, linkActivityToEventName:
		return "events"
	case getTrainingPlanName, applyTrainingPlanName, getWorkoutLibraryName, getWorkoutsInFolderName, createWorkoutName, updateWorkoutName, deleteWorkoutName:
		return "workout-library"
	case getCustomItemsName, getCustomItemByIDName, createCustomItemName, updateCustomItemName, deleteCustomItemName:
		return "custom-items"
	case listAthletesName, selectAthleteName:
		return "coach"
	case listAdvancedCapabilitiesName:
		return "meta"
	default:
		return ""
	}
}

func toolSummary(description string) string {
	description = strings.Join(strings.Fields(description), " ")
	if description == "" {
		return ""
	}
	for i, r := range description {
		if r != '.' {
			continue
		}
		if i == len(description)-1 || description[i+1] == ' ' {
			return strings.TrimSpace(description[:i+1])
		}
	}
	return description
}
