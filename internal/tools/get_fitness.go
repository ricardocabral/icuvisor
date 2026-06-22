package tools

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getFitnessName        = "get_fitness"
	getFitnessDescription = "Get CTL, ATL, and TSB fitness trends for a local date range. Dates are athlete-local YYYY-MM-DD values. Optionally computes per-sport load trend estimates from visible category training load."
	fetchFitnessMessage   = "could not fetch fitness data; check intervals.icu credentials, athlete ID, and date range"

	perSportLoadTrendMethod     = "computed_from_summary_category_load"
	perSportLoadTrendWarmupDays = 84
)

type getFitnessRequest struct {
	StartDate                 string `json:"start_date"`
	EndDate                   string `json:"end_date"`
	IncludeFull               bool   `json:"include_full,omitempty"`
	IncludePerSportLoadTrends bool   `json:"include_per_sport_load_trends,omitempty"`
}

type fitnessResponse struct {
	Rows               []fitnessRow              `json:"fitness"`
	PerSportLoadTrends []perSportLoadTrendBucket `json:"per_sport_load_trends,omitempty"`
	Meta               fitnessMeta               `json:"_meta"`
}

type fitnessRow struct {
	Date string         `json:"date"`
	CTL  *float64       `json:"ctl,omitempty"`
	ATL  *float64       `json:"atl,omitempty"`
	TSB  *float64       `json:"tsb,omitempty"`
	Full map[string]any `json:"full,omitempty"`
}

type perSportLoadTrendBucket struct {
	Sport string                 `json:"sport"`
	Rows  []perSportLoadTrendRow `json:"rows"`
}

type perSportLoadTrendRow struct {
	Date             string   `json:"date"`
	TrainingLoad     float64  `json:"training_load"`
	CTL              float64  `json:"ctl"`
	ATL              float64  `json:"atl"`
	TSB              float64  `json:"tsb"`
	SourceCategories []string `json:"source_categories,omitempty"`
}

type fitnessMeta struct {
	ServerVersion      string                 `json:"server_version"`
	StartDate          string                 `json:"start_date"`
	EndDate            string                 `json:"end_date"`
	Timezone           string                 `json:"timezone"`
	Count              int                    `json:"count"`
	IncludeFull        bool                   `json:"include_full"`
	PerSportLoadTrends *perSportLoadTrendMeta `json:"per_sport_load_trends,omitempty"`
}

type perSportLoadTrendMeta struct {
	Method                     string              `json:"method"`
	SourceEndpoint             string              `json:"source_endpoint"`
	SourceField                string              `json:"source_field"`
	WarmupDaysRequested        int                 `json:"warmup_days_requested"`
	WarmupSummaryDaysAvailable int                 `json:"warmup_summary_days_available"`
	CTLTimeConstantDays        int                 `json:"ctl_time_constant_days"`
	ATLTimeConstantDays        int                 `json:"atl_time_constant_days"`
	BucketMapping              map[string][]string `json:"bucket_mapping"`
	SourceCategoriesByBucket   map[string][]string `json:"source_categories_by_bucket"`
	MissingRequestedDates      []string            `json:"missing_requested_dates,omitempty"`
	Caveats                    []string            `json:"caveats"`
}

func newGetFitnessTool(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getFitnessName, Description: getFitnessDescription, InputSchema: getFitnessInputSchema(), OutputSchema: genericOutputSchema("Fitness rows with CTL, ATL, and TSB, optionally with computed per-sport load trend estimates."), Handler: getFitnessHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func getFitnessHandler(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetFitnessRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidFitnessArgumentsMessage, err)
		}
		unitSystem, timezone, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchFitnessMessage, err)
		}
		fetchStart := args.StartDate
		if args.IncludePerSportLoadTrends {
			fetchStart = perSportLoadTrendLookbackStart(args.StartDate)
		}
		rows, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: fetchStart, End: args.EndDate})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchFitnessMessage, err)
		}
		requestedRows := filterFitnessRowsByDate(rows, args.StartDate, args.EndDate)
		payload := fitnessResponse{Rows: shapeFitnessRows(requestedRows, args.IncludeFull), Meta: fitnessMeta{ServerVersion: normalizeVersion(version), StartDate: args.StartDate, EndDate: args.EndDate, Timezone: timezone, Count: len(requestedRows), IncludeFull: args.IncludeFull}}
		if args.IncludePerSportLoadTrends {
			trends, meta := shapePerSportLoadTrends(rows, args.StartDate, args.EndDate, args.IncludeFull)
			payload.PerSportLoadTrends = trends
			payload.Meta.PerSportLoadTrends = &meta
		}
		return encodeShaped(payload, args.IncludeFull, []string{"fitness"}, version, debugMetadata, getFitnessName, unitSystem, shapeCfg)
	}
}

func decodeGetFitnessRequest(raw json.RawMessage) (getFitnessRequest, error) {
	var args getFitnessRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[getFitnessRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	if !validDate(args.StartDate) || !validDate(args.EndDate) {
		return args, errors.New("start_date and end_date must be YYYY-MM-DD")
	}
	if args.EndDate < args.StartDate {
		return args, errors.New("end_date must be on or after start_date")
	}
	return args, nil
}

func getFitnessInputSchema() map[string]any {
	schema := dateRangeInputSchema("local start date for fitness rows")
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return schema
	}
	properties["include_per_sport_load_trends"] = map[string]any{"type": "boolean", "default": false, "description": "When true, include computed per-sport CTL/ATL/TSB-style load trend estimates for running, cycling, swimming, and other from athlete-summary byCategory training_load. Existing combined fitness rows remain unchanged."}
	return schema
}

func shapeFitnessRows(rows []intervals.SummaryWithCats, includeFull bool) []fitnessRow {
	out := make([]fitnessRow, 0, len(rows))
	for _, row := range rows {
		ctl, atl, tsb := roundPtr(row.Fitness), roundPtr(row.Fatigue), roundPtr(row.Form)
		shaped := fitnessRow{Date: row.Date, CTL: ctl, ATL: atl, TSB: tsb}
		if includeFull {
			shaped.Full = row.Raw
		}
		out = append(out, shaped)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out
}

func filterFitnessRowsByDate(rows []intervals.SummaryWithCats, startDate string, endDate string) []intervals.SummaryWithCats {
	out := make([]intervals.SummaryWithCats, 0, len(rows))
	for _, row := range rows {
		if row.Date >= startDate && row.Date <= endDate {
			out = append(out, row)
		}
	}
	return out
}

func perSportLoadTrendLookbackStart(startDate string) string {
	start, _ := time.Parse(time.DateOnly, startDate)
	return start.AddDate(0, 0, -perSportLoadTrendWarmupDays).Format(time.DateOnly)
}

func shapePerSportLoadTrends(rows []intervals.SummaryWithCats, startDate string, endDate string, includeFull bool) ([]perSportLoadTrendBucket, perSportLoadTrendMeta) {
	loadsByDate := map[string]map[string]float64{}
	categoriesByDateBucket := map[string]map[string]map[string]bool{}
	allCategoriesByBucket := map[string]map[string]bool{}
	var hasNonZeroCategoryLoad bool
	presentDates := map[string]bool{}
	warmupDates := map[string]bool{}
	caveatSet := map[string]bool{}

	for _, row := range rows {
		if !validDate(row.Date) || row.Date > endDate {
			continue
		}
		presentDates[row.Date] = true
		if row.Date < startDate {
			warmupDates[row.Date] = true
		}
		rowCategoryLoad := 0
		if len(row.ByCategory) == 0 && row.TrainingLoad > 0 {
			caveatSet["some summary rows have training_load but no byCategory sport breakdown; that load cannot be assigned to a per-sport trend"] = true
		}
		for _, category := range row.ByCategory {
			if categoryTrainingLoadMissing(category) {
				caveatSet["some byCategory rows omit training_load; missing category loads are treated as 0"] = true
			}
			bucket := sportLoadBucket(category.Category)
			label := strings.TrimSpace(category.Category)
			if label == "" {
				label = "unknown"
			}
			rowCategoryLoad += category.TrainingLoad
			if category.TrainingLoad > 0 {
				hasNonZeroCategoryLoad = true
			}
			ensureDateBucket(loadsByDate, row.Date)[bucket] += float64(category.TrainingLoad)
			ensureDateBucketCategories(categoriesByDateBucket, row.Date, bucket)[label] = true
			ensureBucketCategories(allCategoriesByBucket, bucket)[label] = true
		}
		if row.TrainingLoad != 0 && rowCategoryLoad != row.TrainingLoad {
			caveatSet["some byCategory training_load totals differ from the combined daily training_load; per-sport trends use category totals only"] = true
		}
	}

	for _, date := range dateRange(startDate, endDate) {
		if !presentDates[date] {
			caveatSet["some requested calendar dates were absent from upstream summary rows; per-sport state advances with zero load for absent dates"] = true
		}
	}
	if len(warmupDates) < perSportLoadTrendWarmupDays {
		caveatSet["fewer than 84 warm-up summary days were available; early computed per-sport CTL/ATL/TSB estimates may be understated"] = true
	}
	if !hasNonZeroCategoryLoad {
		caveatSet["no non-zero per-sport category load was available in the fetched summary rows"] = true
	}

	buckets := []string{"running", "cycling", "swimming", "other"}
	trends := make([]perSportLoadTrendBucket, 0, len(buckets))
	states := map[string]struct{ ctl, atl float64 }{}
	for _, bucket := range buckets {
		trends = append(trends, perSportLoadTrendBucket{Sport: bucket})
	}
	trendIndex := map[string]int{"running": 0, "cycling": 1, "swimming": 2, "other": 3}
	for _, date := range dateRange(perSportLoadTrendLookbackStart(startDate), endDate) {
		for _, bucket := range buckets {
			load := loadsByDate[date][bucket]
			state := states[bucket]
			state.ctl = state.ctl + (load-state.ctl)/analysis.FitnessProjectionCTLTimeConstantDays
			state.atl = state.atl + (load-state.atl)/analysis.FitnessProjectionATLTimeConstantDays
			states[bucket] = state
			if date < startDate {
				continue
			}
			row := perSportLoadTrendRow{Date: date, TrainingLoad: round(load, 3), CTL: round(state.ctl, 3), ATL: round(state.atl, 3), TSB: round(state.ctl-state.atl, 3)}
			if includeFull {
				row.SourceCategories = sortedStringSet(categoriesByDateBucket[date][bucket])
			}
			idx := trendIndex[bucket]
			trends[idx].Rows = append(trends[idx].Rows, row)
		}
	}

	meta := perSportLoadTrendMeta{
		Method:                     perSportLoadTrendMethod,
		SourceEndpoint:             "athlete-summary.json",
		SourceField:                "byCategory[].training_load",
		WarmupDaysRequested:        perSportLoadTrendWarmupDays,
		WarmupSummaryDaysAvailable: len(warmupDates),
		CTLTimeConstantDays:        analysis.FitnessProjectionCTLTimeConstantDays,
		ATLTimeConstantDays:        analysis.FitnessProjectionATLTimeConstantDays,
		BucketMapping:              sportLoadBucketMapping(),
		SourceCategoriesByBucket:   sortedCategoryBuckets(allCategoriesByBucket),
		MissingRequestedDates:      missingRequestedDates(presentDates, startDate, endDate),
		Caveats:                    sortedCaveats(caveatSet),
	}
	return trends, meta
}

func ensureDateBucket(values map[string]map[string]float64, date string) map[string]float64 {
	if values[date] == nil {
		values[date] = map[string]float64{}
	}
	return values[date]
}

func ensureDateBucketCategories(values map[string]map[string]map[string]bool, date string, bucket string) map[string]bool {
	if values[date] == nil {
		values[date] = map[string]map[string]bool{}
	}
	return ensureBucketCategories(values[date], bucket)
}

func ensureBucketCategories(values map[string]map[string]bool, bucket string) map[string]bool {
	if values[bucket] == nil {
		values[bucket] = map[string]bool{}
	}
	return values[bucket]
}

func categoryTrainingLoadMissing(category intervals.CategorySummary) bool {
	if category.Raw == nil {
		return false
	}
	_, ok := category.Raw["training_load"]
	return !ok
}

func sportLoadBucket(category string) string {
	normalized := normalizeSportLoadCategory(category)
	switch normalized {
	case "run", "trailrun", "virtualrun", "treadmill", "treadmillrun":
		return "running"
	case "ride", "virtualride", "bike", "bikeride", "cycling", "cycle", "indoorcycling", "indoorride", "gravelride", "mountainbike", "mountainbikeride", "mtb", "ebikeride":
		return "cycling"
	case "swim", "openwaterswim", "poolswim":
		return "swimming"
	default:
		return "other"
	}
}

func normalizeSportLoadCategory(category string) string {
	replacer := strings.NewReplacer(" ", "", "_", "", "-", "")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(category)))
}

func sportLoadBucketMapping() map[string][]string {
	return map[string][]string{
		"running":  {"run", "trailrun", "virtualrun", "treadmill", "treadmillrun"},
		"cycling":  {"ride", "virtualride", "bike", "bikeride", "cycling", "cycle", "indoorcycling", "indoorride", "gravelride", "mountainbike", "mountainbikeride", "mtb", "ebikeride"},
		"swimming": {"swim", "openwaterswim", "poolswim"},
		"other":    {"fallback for empty, unknown, or unsupported category labels"},
	}
}

func dateRange(startDate string, endDate string) []string {
	start, _ := time.Parse(time.DateOnly, startDate)
	end, _ := time.Parse(time.DateOnly, endDate)
	if end.Before(start) {
		return nil
	}
	days := int(math.Round(end.Sub(start).Hours()/24)) + 1
	out := make([]string, 0, days)
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		out = append(out, date.Format(time.DateOnly))
	}
	return out
}

func missingRequestedDates(present map[string]bool, startDate string, endDate string) []string {
	missing := []string{}
	for _, date := range dateRange(startDate, endDate) {
		if !present[date] {
			missing = append(missing, date)
		}
	}
	return missing
}

func sortedCategoryBuckets(values map[string]map[string]bool) map[string][]string {
	out := map[string][]string{}
	for _, bucket := range []string{"running", "cycling", "swimming", "other"} {
		out[bucket] = sortedStringSet(values[bucket])
	}
	return out
}

func sortedStringSet(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func sortedCaveats(values map[string]bool) []string {
	out := sortedStringSet(values)
	if len(out) == 0 {
		return []string{"per-sport CTL/ATL/TSB are computed estimates from visible summary category load, not upstream-native per-sport fitness values"}
	}
	return append([]string{"per-sport CTL/ATL/TSB are computed estimates from visible summary category load, not upstream-native per-sport fitness values"}, out...)
}
