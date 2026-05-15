package toolchecks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const DefaultSchemaSnapshotDir = "internal/tools/schema_snapshot"

type Snapshot struct {
	ToolName string
	Path     string
	Raw      []byte
	Schema   map[string]any
}

type SchemaReport struct {
	Failures []SchemaFailure
	Added    []string
}

func (r SchemaReport) OK() bool { return len(r.Failures) == 0 }

type SchemaFailure struct {
	ToolName string
	Property string
	Kind     string
	Message  string
	Baseline string
	Current  string
}

type schemaCatalogRoundTripper struct{}

func (schemaCatalogRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("schema catalog generation must not perform HTTP")
}

var schemaCatalogToolNames = map[string]struct{}{
	"add_activity_message":                {},
	"add_or_update_event":                 {},
	"apply_training_plan":                 {},
	"create_workout":                      {},
	"delete_workout":                      {},
	"get_activities":                      {},
	"get_activity_details":                {},
	"get_activity_intervals":              {},
	"get_activity_messages":               {},
	"get_activity_splits":                 {},
	"get_activity_streams":                {},
	"get_athlete_profile":                 {},
	"get_best_efforts":                    {},
	"get_custom_item_by_id":               {},
	"get_custom_items":                    {},
	"get_event_by_id":                     {},
	"get_events":                          {},
	"get_extended_metrics":                {},
	"get_fitness":                         {},
	"get_power_curves":                    {},
	"get_training_plan":                   {},
	"get_training_summary":                {},
	"get_wellness_data":                   {},
	"get_workout_library":                 {},
	"get_workouts_in_folder":              {},
	"icuvisor_list_advanced_capabilities": {},
	"link_activity_to_event":              {},
	"update_sport_settings":               {},
	"update_wellness":                     {},
	"update_workout":                      {},
}

func generateSchemaCatalogTools() ([]tools.Tool, error) {
	client, err := intervals.NewClient(intervals.Options{
		Config: config.Config{
			APIKey:     strings.Repeat("x", 8),
			AthleteID:  "12345",
			APIBaseURL: "http://127.0.0.1",
		},
		Version:    "snapshot",
		HTTPClient: &http.Client{Transport: schemaCatalogRoundTripper{}},
	})
	if err != nil {
		return nil, fmt.Errorf("creating schema catalog client: %w", err)
	}
	registrar := &schemaRegistrar{}
	registry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{Version: "snapshot", TimezoneFallback: "UTC"})
	if err := registry.Register(context.Background(), registrar); err != nil {
		return nil, fmt.Errorf("registering tools: %w", err)
	}
	filtered := make([]tools.Tool, 0, len(schemaCatalogToolNames))
	for _, tool := range registrar.tools {
		if _, ok := schemaCatalogToolNames[tool.Name]; ok {
			filtered = append(filtered, tool)
		}
	}
	return filtered, nil
}

func GenerateSchemaSnapshots() (map[string]Snapshot, error) {
	toolCatalog, err := generateSchemaCatalogTools()
	if err != nil {
		return nil, err
	}
	out := make(map[string]Snapshot, len(toolCatalog))
	for _, tool := range toolCatalog {
		raw, err := CanonicalJSON(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("marshalling schema for %s: %w", tool.Name, err)
		}
		var schema map[string]any
		if err := json.Unmarshal(raw, &schema); err != nil {
			return nil, fmt.Errorf("decoding generated schema for %s: %w", tool.Name, err)
		}
		out[tool.Name] = Snapshot{ToolName: tool.Name, Raw: raw, Schema: schema}
	}
	return out, nil
}

func WriteGeneratedSchemaSnapshots(dir string) error {
	generated, err := GenerateSchemaSnapshots()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating snapshot directory: %w", err)
	}
	if err := removeExistingSnapshots(dir); err != nil {
		return err
	}
	for _, name := range sortedSnapshotNames(generated) {
		path := filepath.Join(dir, name+".json")
		if err := os.WriteFile(path, generated[name].Raw, 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}

func LoadSchemaSnapshots(dir string) (map[string]Snapshot, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("finding snapshots: %w", err)
	}
	out := make(map[string]Snapshot, len(matches))
	for _, path := range matches {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		var schema map[string]any
		if err := json.Unmarshal(raw, &schema); err != nil {
			return nil, fmt.Errorf("decoding %s: %w", path, err)
		}
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		out[name] = Snapshot{ToolName: name, Path: path, Raw: raw, Schema: schema}
	}
	return out, nil
}

func CheckSnapshotFreshness(currentDir string, generated map[string]Snapshot) (SchemaReport, error) {
	current, err := LoadSchemaSnapshots(currentDir)
	if err != nil {
		return SchemaReport{}, err
	}
	report := SchemaReport{}
	for _, name := range sortedSnapshotNames(generated) {
		currentSnapshot, ok := current[name]
		if !ok {
			report.Failures = append(report.Failures, SchemaFailure{ToolName: name, Kind: "missing-current-snapshot", Message: "generated live registry schema has no committed snapshot", Current: filepath.Join(currentDir, name+".json")})
			continue
		}
		if !bytes.Equal(currentSnapshot.Raw, generated[name].Raw) {
			report.Failures = append(report.Failures, SchemaFailure{ToolName: name, Kind: "snapshot-drift", Message: "committed snapshot differs from canonical live registry schema; run go run ./scripts/snapshot_tool_schemas.go", Current: currentSnapshot.Path})
		}
	}
	for _, name := range sortedSnapshotNames(current) {
		if _, ok := generated[name]; !ok {
			report.Failures = append(report.Failures, SchemaFailure{ToolName: name, Kind: "stale-current-snapshot", Message: "committed snapshot has no live registry tool", Current: current[name].Path})
		}
	}
	return report, nil
}

func CheckSchemaStability(baselineDir string, currentDir string, generated map[string]Snapshot) (SchemaReport, error) {
	baseline, err := loadBaselineSchemaSnapshots(baselineDir)
	if err != nil {
		return SchemaReport{}, err
	}
	report := SchemaReport{}
	for _, name := range sortedSnapshotNames(baseline) {
		base := baseline[name]
		current, ok := generated[name]
		if !ok {
			report.Failures = append(report.Failures, SchemaFailure{ToolName: name, Kind: "tool-removed", Message: "baseline tool is missing from current registry; removals and renames require a new compatibility plan", Baseline: base.Path, Current: filepath.Join(currentDir, name+".json")})
			continue
		}
		current.Path = filepath.Join(currentDir, name+".json")
		report.Failures = append(report.Failures, compareStableSchema(base, current)...)
	}
	for _, name := range sortedSnapshotNames(generated) {
		if _, ok := baseline[name]; !ok {
			report.Added = append(report.Added, name)
		}
	}
	return report, nil
}

func loadBaselineSchemaSnapshots(dir string) (map[string]Snapshot, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("baseline snapshot directory %s is not accessible: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("baseline snapshot path %s is not a directory", dir)
	}
	baseline, err := LoadSchemaSnapshots(dir)
	if err != nil {
		return nil, err
	}
	if len(baseline) == 0 {
		return nil, fmt.Errorf("baseline snapshot directory %s contains no *.json snapshots", dir)
	}
	return baseline, nil
}

func CanonicalJSON(value any) ([]byte, error) {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func compareStableSchema(base Snapshot, current Snapshot) []SchemaFailure {
	var failures []SchemaFailure
	if base.Schema["type"] != current.Schema["type"] {
		failures = append(failures, SchemaFailure{ToolName: base.ToolName, Kind: "root-type-changed", Message: "root schema type changed", Baseline: base.Path, Current: current.Path})
	}
	if boolValue(base.Schema["additionalProperties"]) && !boolValue(current.Schema["additionalProperties"]) {
		failures = append(failures, SchemaFailure{ToolName: base.ToolName, Kind: "root-more-restrictive", Message: "additionalProperties changed from true to false", Baseline: base.Path, Current: current.Path})
	}
	baseProps := properties(base.Schema)
	currentProps := properties(current.Schema)
	for _, prop := range sortedPropertyNames(baseProps) {
		currentProp, ok := currentProps[prop]
		if !ok {
			failures = append(failures, SchemaFailure{ToolName: base.ToolName, Property: prop, Kind: "property-removed", Message: "baseline argument property is missing; removals and renames require a new tool name", Baseline: base.Path, Current: current.Path})
			continue
		}
		if !reflect.DeepEqual(baseProps[prop], currentProp) {
			failures = append(failures, SchemaFailure{ToolName: base.ToolName, Property: prop, Kind: "property-changed", Message: "baseline argument property schema changed; stable argument schemas are additive-only", Baseline: base.Path, Current: current.Path})
		}
	}
	baseRequired := stringSet(base.Schema["required"])
	for required := range stringSet(current.Schema["required"]) {
		if _, existed := baseRequired[required]; !existed {
			failures = append(failures, SchemaFailure{ToolName: base.ToolName, Property: required, Kind: "required-added", Message: "newly required arguments are not additive; new arguments must be optional", Baseline: base.Path, Current: current.Path})
		}
	}
	return failures
}

func removeExistingSnapshots(dir string) error {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("finding existing snapshots: %w", err)
	}
	for _, match := range matches {
		if err := os.Remove(match); err != nil {
			return fmt.Errorf("removing stale snapshot %s: %w", match, err)
		}
	}
	return nil
}

func properties(schema map[string]any) map[string]any {
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return props
}

func sortedSnapshotNames(snapshots map[string]Snapshot) []string {
	names := make([]string, 0, len(snapshots))
	for name := range snapshots {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedPropertyNames(props map[string]any) []string {
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func boolValue(value any) bool {
	v, ok := value.(bool)
	return ok && v
}

func stringSet(value any) map[string]struct{} {
	out := map[string]struct{}{}
	items, ok := value.([]any)
	if !ok {
		return out
	}
	for _, item := range items {
		if text, ok := item.(string); ok {
			out[text] = struct{}{}
		}
	}
	return out
}

type schemaRegistrar struct {
	tools []tools.Tool
}

func (r *schemaRegistrar) AddTool(tool tools.Tool) error {
	r.tools = append(r.tools, tool)
	return nil
}

type schemaCatalogClient struct{}

var (
	_ tools.ProfileClient               = schemaCatalogClient{}
	_ tools.FitnessClient               = schemaCatalogClient{}
	_ tools.WellnessClient              = schemaCatalogClient{}
	_ tools.WellnessWriterClient        = schemaCatalogClient{}
	_ tools.SportSettingsWriterClient   = schemaCatalogClient{}
	_ tools.BestEffortsClient           = schemaCatalogClient{}
	_ tools.PowerCurvesClient           = schemaCatalogClient{}
	_ tools.ActivitiesClient            = schemaCatalogClient{}
	_ tools.EventsClient                = schemaCatalogClient{}
	_ tools.EventByIDClient             = schemaCatalogClient{}
	_ tools.EventWriterClient           = schemaCatalogClient{}
	_ tools.ActivityEventLinkClient     = schemaCatalogClient{}
	_ tools.TrainingPlanClient          = schemaCatalogClient{}
	_ tools.WorkoutLibraryClient        = schemaCatalogClient{}
	_ tools.WorkoutCreatorClient        = schemaCatalogClient{}
	_ tools.WorkoutUpdaterClient        = schemaCatalogClient{}
	_ tools.WorkoutDeleterClient        = schemaCatalogClient{}
	_ tools.CustomItemsClient           = schemaCatalogClient{}
	_ tools.ActivityDetailsClient       = schemaCatalogClient{}
	_ tools.ActivityIntervalsClient     = schemaCatalogClient{}
	_ tools.ActivityStreamsClient       = schemaCatalogClient{}
	_ tools.ActivityMessagesClient      = schemaCatalogClient{}
	_ tools.ActivityMessageWriterClient = schemaCatalogClient{}
	_ tools.ExtendedMetricsClient       = schemaCatalogClient{}
)

func (schemaCatalogClient) GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error) {
	return intervals.AthleteWithSportSettings{}, nil
}

func (schemaCatalogClient) ListAthleteSummary(context.Context, intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	return nil, nil
}

func (schemaCatalogClient) ListWellness(context.Context, intervals.WellnessParams) ([]intervals.Wellness, error) {
	return nil, nil
}

func (schemaCatalogClient) UpdateWellness(context.Context, intervals.WriteWellnessParams) (intervals.Wellness, error) {
	return intervals.Wellness{}, nil
}

func (schemaCatalogClient) UpdateSportSettings(context.Context, intervals.WriteSportSettingsParams) (intervals.SportSettings, error) {
	return intervals.SportSettings{}, nil
}

func (schemaCatalogClient) ListAthletePowerCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error) {
	return intervals.DataCurveSet{}, nil
}

func (schemaCatalogClient) ListAthleteHRCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error) {
	return intervals.DataCurveSet{}, nil
}

func (schemaCatalogClient) ListAthletePaceCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error) {
	return intervals.DataCurveSet{}, nil
}

func (schemaCatalogClient) ListActivities(context.Context, intervals.ListActivitiesParams) ([]intervals.Activity, error) {
	return nil, nil
}

func (schemaCatalogClient) ListEvents(context.Context, intervals.ListEventsParams) ([]intervals.Event, error) {
	return nil, nil
}

func (schemaCatalogClient) GetEvent(context.Context, string) (intervals.Event, error) {
	return intervals.Event{}, nil
}

func (schemaCatalogClient) AddOrUpdateEvent(context.Context, intervals.WriteEventParams) (intervals.Event, error) {
	return intervals.Event{}, nil
}

func (schemaCatalogClient) LinkActivityToEvent(context.Context, intervals.LinkActivityToEventParams) (intervals.Activity, error) {
	return intervals.Activity{}, nil
}

func (schemaCatalogClient) GetTrainingPlan(context.Context) (intervals.TrainingPlan, error) {
	return intervals.TrainingPlan{}, nil
}

func (schemaCatalogClient) ListWorkoutFolders(context.Context) ([]intervals.WorkoutFolder, error) {
	return nil, nil
}

func (schemaCatalogClient) ListLibraryWorkouts(context.Context) ([]intervals.Workout, error) {
	return nil, nil
}

func (schemaCatalogClient) CreateLibraryWorkout(context.Context, intervals.WriteWorkoutParams) (intervals.Workout, error) {
	return intervals.Workout{}, nil
}

func (schemaCatalogClient) UpdateLibraryWorkout(context.Context, intervals.WriteWorkoutParams) (intervals.Workout, error) {
	return intervals.Workout{}, nil
}

func (schemaCatalogClient) DeleteLibraryWorkout(context.Context, string) error {
	return nil
}

func (schemaCatalogClient) ListCustomItems(context.Context) ([]intervals.CustomItem, error) {
	return nil, nil
}

func (schemaCatalogClient) GetCustomItem(context.Context, string) (intervals.CustomItem, error) {
	return intervals.CustomItem{}, nil
}

func (schemaCatalogClient) GetActivity(context.Context, string) (intervals.Activity, error) {
	return intervals.Activity{}, nil
}

func (schemaCatalogClient) GetActivityIntervals(context.Context, string) (intervals.IntervalsDTO, error) {
	return intervals.IntervalsDTO{}, nil
}

func (schemaCatalogClient) GetActivityStreams(context.Context, intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error) {
	return nil, nil
}

func (schemaCatalogClient) GetActivityMessages(context.Context, intervals.ActivityMessagesParams) ([]intervals.ActivityMessage, error) {
	return nil, nil
}

func (schemaCatalogClient) AddActivityMessage(context.Context, intervals.AddActivityMessageParams) (intervals.NewActivityMessage, error) {
	return intervals.NewActivityMessage{}, nil
}

func (schemaCatalogClient) GetActivityPowerVsHR(context.Context, string) (intervals.PowerVsHR, error) {
	return intervals.PowerVsHR{}, nil
}
