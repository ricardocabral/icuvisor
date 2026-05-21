package tools

import (
	"context"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestCatalogDescriptors(t *testing.T) {
	t.Parallel()

	descriptors := Catalog()
	if len(descriptors) == 0 {
		t.Fatal("Catalog() returned no tools")
	}

	snakeCase := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	allowedGroups := map[string]struct{}{
		"activities":      {},
		"analyzers":       {},
		"coach":           {},
		"custom-items":    {},
		"events":          {},
		"fitness":         {},
		"meta":            {},
		"settings":        {},
		"wellness":        {},
		"workout-library": {},
	}
	seen := make(map[string]ToolDescriptor, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.Name == "" {
			t.Fatal("Catalog() returned descriptor with empty name")
		}
		if !snakeCase.MatchString(descriptor.Name) {
			t.Fatalf("tool name %q is not snake_case", descriptor.Name)
		}
		if _, exists := seen[descriptor.Name]; exists {
			t.Fatalf("duplicate Catalog() descriptor for %q", descriptor.Name)
		}
		if descriptor.Group == "" || descriptor.Tier == "" || descriptor.Safety == "" || descriptor.Summary == "" || descriptor.Anchor == "" {
			t.Fatalf("descriptor %q has an empty required field: %#v", descriptor.Name, descriptor)
		}
		if _, ok := allowedGroups[descriptor.Group]; !ok {
			t.Fatalf("descriptor %q group = %q, want one of the documented groups", descriptor.Name, descriptor.Group)
		}
		if descriptor.Anchor != descriptor.Name {
			t.Fatalf("descriptor %q anchor = %q, want same as name", descriptor.Name, descriptor.Anchor)
		}
		seen[descriptor.Name] = descriptor
	}

	for i := 1; i < len(descriptors); i++ {
		prev := descriptors[i-1]
		cur := descriptors[i]
		if prev.Group > cur.Group || (prev.Group == cur.Group && prev.Name > cur.Name) {
			t.Fatalf("Catalog() is not sorted by group then name at %d: %q/%q before %q/%q", i, prev.Group, prev.Name, cur.Group, cur.Name)
		}
	}
}

func TestCatalogMatchesRegistryAndPRDRegisteredTools(t *testing.T) {
	t.Parallel()

	catalogNames := descriptorNameSet(Catalog())
	registeredNames := registeredToolNameSet(t)
	if diff := missingNames(registeredNames, catalogNames); len(diff) > 0 {
		t.Fatalf("registered tools missing from Catalog(): %v", diff)
	}
	if diff := missingNames(catalogNames, registeredNames); len(diff) > 0 {
		t.Fatalf("Catalog() returned tools not registered by registry: %v", diff)
	}

	prdNames := prdToolCatalogNames(t)
	for _, name := range prdNames {
		_, inRegistry := registeredNames[name]
		_, inCatalog := catalogNames[name]
		if inRegistry && !inCatalog {
			t.Fatalf("PRD-documented registered tool %q missing from Catalog()", name)
		}
	}
}

func TestCatalogIncludesAnalyzerFamilyPlacement(t *testing.T) {
	t.Parallel()

	descriptors := descriptorNameSet(Catalog())
	cases := map[string]string{
		computeActivitySegmentStatsName: "Use when the prompt asks for an average, maximum, normalized power, or zone-time statistic over one explicit activity segment as the analyzer-family raw-stream exception",
		analyzeTrendName:                "Use when the prompt asks whether an analysis metric is trending up, trending down, or changing versus baseline",
		analyzeDistributionName:         "Use when the prompt asks for an analysis metric's distribution, histogram, quantiles, or outliers",
		analyzeCorrelationName:          "Use when the prompt asks whether two analysis metrics are correlated or lagged together",
		analyzeEffortsDeltaName:         "Use when the prompt asks whether best-effort power, heart-rate, or pace buckets changed versus baseline",
		getFitnessProjectionName:        "Use when the prompt asks to project CTL, ATL, or TSB forward",
		computeZoneTimeName:             "Use when the prompt asks for time in power, heart-rate, or pace zones over a date window",
		computeLoadBalanceName:          "Use when the prompt asks whether training distribution is polarized, pyramidal, threshold-heavy, or balanced across low/moderate/high intensity",
		computeBaselineName:             "Use when the prompt asks whether a metric is high, low, suppressed, elevated, or unusual versus a baseline window",
		computeComplianceRateName:       "Use when the prompt asks how well completed activities matched scheduled workouts, targets, sport, or event type",
	}
	for name, summaryNeedle := range cases {
		descriptor, exists := descriptors[name]
		if !exists {
			t.Fatalf("Catalog() missing %q", name)
		}
		wantGroup := "analyzers"
		if name == getFitnessProjectionName {
			wantGroup = "fitness"
		}
		wantTier := safety.ToolsetFull
		if _, candidate := analyzerCorePromotionCandidateSet()[name]; candidate {
			wantTier = safety.ToolsetCore
		}
		if descriptor.Group != wantGroup || descriptor.Tier != string(wantTier) {
			t.Fatalf("descriptor = %#v, want %s/%s", descriptor, wantGroup, wantTier)
		}
		if !strings.Contains(descriptor.Summary, summaryNeedle) {
			t.Fatalf("summary for %s = %q, want %q", name, descriptor.Summary, summaryNeedle)
		}
	}
}

func TestCatalogAnalyzerActivationHints(t *testing.T) {
	t.Parallel()

	descriptors := descriptorNameSet(Catalog())
	for _, name := range analyzerFamilyCatalogNames() {
		descriptor, exists := descriptors[name]
		if !exists {
			t.Fatalf("Catalog() missing analyzer-family tool %q", name)
		}
		if !strings.HasPrefix(descriptor.Summary, "Use when the prompt asks ") {
			t.Fatalf("summary for %s = %q, want concrete prompt-shape activation hint", name, descriptor.Summary)
		}
		avoidance := strings.Contains(descriptor.Summary, "do not fetch get_") && (strings.Contains(descriptor.Summary, "reduce them in chat") || strings.Contains(descriptor.Summary, "bin them in chat") || strings.Contains(descriptor.Summary, "model the curve in chat"))
		if !avoidance {
			t.Fatalf("summary for %s = %q, want explicit get_* row/stream avoidance language", name, descriptor.Summary)
		}
	}
}

func analyzerFamilyCatalogNames() []string {
	return []string{
		analyzeTrendName,
		analyzeDistributionName,
		analyzeCorrelationName,
		analyzeEffortsDeltaName,
		computeZoneTimeName,
		computeLoadBalanceName,
		computeBaselineName,
		computeComplianceRateName,
		computeActivitySegmentStatsName,
		getActivityHistogramName,
		getFitnessProjectionName,
	}
}

func TestCatalogSummariesUseFirstDescriptionSentence(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		want string
	}{
		{name: getActivitiesName, want: "List activities for a date range with terse unit-disambiguated rows, calories_burned as active/exercise calories, Strava-unavailable detection, and opaque pagination."},
		{name: updateWellnessName, want: "Update one athlete-local wellness row with sparse manual fields: subjective scales, measurements, injury text, and locked; legacy feel remains in the input schema for compatibility but rejects with field_not_writable: feel (not accepted by intervals.icu wellness write), device-owned sleepScore rejects with field_not_writable: sleepScore (device-managed), and _native rejects with field_not_writable: _native (bridge-managed)."},
	}

	descriptors := descriptorNameSet(Catalog())
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			descriptor, exists := descriptors[tc.name]
			if !exists {
				t.Fatalf("Catalog() missing %q", tc.name)
			}
			if descriptor.Summary != tc.want {
				t.Fatalf("summary for %q = %q, want %q", tc.name, descriptor.Summary, tc.want)
			}
		})
	}
}

func descriptorNameSet(descriptors []ToolDescriptor) map[string]ToolDescriptor {
	out := make(map[string]ToolDescriptor, len(descriptors))
	for _, descriptor := range descriptors {
		out[descriptor.Name] = descriptor
	}
	return out
}

func registeredToolNameSet(t *testing.T) map[string]ToolDescriptor {
	t.Helper()

	registrar := &collectingRegistrar{}
	registry := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{
		Version:          "test",
		TimezoneFallback: "UTC",
		Capability:       safety.NewCapability(safety.ModeFull),
		Toolset:          safety.ToolsetFull,
		CoachModeEnabled: true,
		CoachConfig: coach.Config{Athletes: []coach.Athlete{
			{ID: "i12345", Label: "Test Athlete"},
		}},
	})
	if err := registry.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	out := make(map[string]ToolDescriptor, len(registrar.tools))
	for _, tool := range registrar.tools {
		out[tool.Name] = ToolDescriptor{Name: tool.Name}
	}
	return out
}

func prdToolCatalogNames(t *testing.T) []string {
	t.Helper()

	body, err := os.ReadFile("../../docs/prd/PRD-icuvisor.md")
	if err != nil {
		t.Fatalf("reading PRD: %v", err)
	}
	text := string(body)
	start := strings.Index(text, "#### C. Tool catalog")
	end := strings.Index(text, "#### D. Response shaping")
	if start < 0 || end < start {
		t.Fatalf("could not locate PRD §7.2.C tool catalog")
	}

	toolName := regexp.MustCompile("`([a-z][a-z0-9_]+)`")
	seen := map[string]struct{}{}
	for _, line := range strings.Split(text[start:end], "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- `") {
			continue
		}
		for _, match := range toolName.FindAllStringSubmatch(trimmed, -1) {
			seen[match[1]] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}

func missingNames(want map[string]ToolDescriptor, got map[string]ToolDescriptor) []string {
	missing := make([]string, 0)
	for name := range want {
		if _, exists := got[name]; !exists {
			missing = append(missing, name)
		}
	}
	slices.Sort(missing)
	return missing
}
