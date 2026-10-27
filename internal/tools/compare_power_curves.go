package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	comparePowerCurvesName        = "compare_power_curves"
	comparePowerCurvesDescription = "Compare current and previous upstream training-season power bests with separate watts and W/kg effort attribution. Supply an athlete-local as_of_date; up to four seasons return compact duration buckets, available fitted models, and after-kJ curves."
	maxComparedPowerSeasons       = 4
	maxComparedPowerBuckets       = 16
)

type comparePowerCurvesRequest struct {
	AsOfDate        string `json:"as_of_date"`
	SeasonOffsets   []int  `json:"season_offsets,omitempty"`
	Sport           string `json:"sport,omitempty"`
	DurationSeconds []int  `json:"duration_seconds,omitempty"`
	IncludeFull     bool   `json:"include_full,omitempty"`
}

type comparedPowerSeason struct {
	SeasonOffset     int                    `json:"season_offset"`
	CurveSpec        string                 `json:"curve_spec"`
	CurveID          string                 `json:"curve_id,omitempty"`
	Label            string                 `json:"label,omitempty"`
	Oldest           string                 `json:"oldest,omitempty"`
	Newest           string                 `json:"newest,omitempty"`
	Status           string                 `json:"status"`
	Points           []powerCurvePoint      `json:"points"`
	MissingBuckets   []int                  `json:"missing_buckets,omitempty"`
	PowerModels      []powerModelSummary    `json:"power_models,omitempty"`
	DurabilityCurves []powerDurabilityCurve `json:"durability_curves,omitempty"`
}

type comparePowerCurvesResponse struct {
	Sport   string                 `json:"sport"`
	Seasons []comparedPowerSeason  `json:"seasons"`
	Full    map[string]any         `json:"full,omitempty"`
	Meta    comparePowerCurvesMeta `json:"_meta"`
}

type comparePowerCurvesMeta struct {
	AsOfDate        string `json:"as_of_date"`
	SeasonSource    string `json:"season_source"`
	DateBasis       string `json:"date_basis"`
	DurationSeconds []int  `json:"duration_seconds"`
	IncludeFull     bool   `json:"include_full"`
}

func newComparePowerCurvesTool(client PowerCurvesClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: comparePowerCurvesName, Description: comparePowerCurvesDescription, InputSchema: comparePowerCurvesInputSchema(), OutputSchema: genericOutputSchema("Bounded upstream season power curves with missing-data status, local dates, W/kg attribution, and optional upstream fitted power models."), Handler: func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeComparePowerCurvesRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError("invalid power comparison; provide athlete-local as_of_date, 1-4 unique season_offsets from 0-10, and up to 16 positive duration_seconds", err)
		}
		payload := comparePowerCurvesResponse{Sport: args.Sport, Seasons: []comparedPowerSeason{}, Meta: comparePowerCurvesMeta{AsOfDate: args.AsOfDate, SeasonSource: "upstream_season_selectors", DateBasis: "athlete_local", DurationSeconds: args.DurationSeconds, IncludeFull: args.IncludeFull}}
		if args.IncludeFull {
			payload.Full = map[string]any{}
		}
		for _, offset := range args.SeasonOffsets {
			spec := "s" + strconv.Itoa(offset)
			set, err := client.ListAthletePowerCurves(ctx, intervals.CurveParams{Sport: args.Sport, Now: args.AsOfDate, CurveSpec: powerCurveRequestSpec(spec), DurationSeconds: args.DurationSeconds})
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return Result{}, err
				}
				return Result{}, NewUserError(fetchPowerCurvesMessage, err)
			}
			curve := standardPowerCurve(set)
			points, missing := bucketPowerCurve(curve, args.DurationSeconds)
			addPowerCurveDates(points, set.Activities)
			season := comparedPowerSeason{SeasonOffset: offset, CurveSpec: spec, CurveID: curve.ID, Label: curve.Label, Oldest: curve.StartDateLocal, Newest: curve.EndDateLocal, Status: "unavailable", Points: points, MissingBuckets: missing, PowerModels: summarizePowerModels(curve.PowerModels), DurabilityCurves: bucketDurabilityPowerCurves(set, args.DurationSeconds)}
			if len(points) > 0 {
				season.Status = "available"
			}
			payload.Seasons = append(payload.Seasons, season)
			if args.IncludeFull {
				payload.Full[spec] = set.Raw
			}
		}
		return encodeShaped(payload, args.IncludeFull, []string{"seasons"}, version, debugMetadata, comparePowerCurvesName, response.UnitSystemMetric, shapeCfg)
	}})
}

func decodeComparePowerCurvesRequest(raw json.RawMessage) (comparePowerCurvesRequest, error) {
	args, err := DecodeStrict[comparePowerCurvesRequest](raw)
	if err != nil {
		return args, err
	}
	args.AsOfDate = strings.TrimSpace(args.AsOfDate)
	if !validDate(args.AsOfDate) {
		return args, fmt.Errorf("as_of_date must be athlete-local YYYY-MM-DD")
	}
	args.Sport = firstNonEmpty(strings.TrimSpace(args.Sport), defaultPowerCurveSport)
	if args.SeasonOffsets == nil {
		args.SeasonOffsets = []int{0, 1}
	}
	if len(args.SeasonOffsets) < 1 || len(args.SeasonOffsets) > maxComparedPowerSeasons {
		return args, fmt.Errorf("season_offsets must contain 1-4 entries")
	}
	seen := map[int]bool{}
	for _, offset := range args.SeasonOffsets {
		if offset < 0 || offset > 10 || seen[offset] {
			return args, fmt.Errorf("season_offsets must be unique integers from 0-10")
		}
		seen[offset] = true
	}
	if len(args.DurationSeconds) > maxComparedPowerBuckets {
		return args, fmt.Errorf("duration_seconds must contain at most 16 entries")
	}
	for _, duration := range args.DurationSeconds {
		if duration <= 0 {
			return args, fmt.Errorf("duration_seconds must be positive")
		}
	}
	args.DurationSeconds = normalizePositiveInts(args.DurationSeconds, defaultDurationBuckets)
	return args, nil
}

func comparePowerCurvesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"as_of_date"}, "properties": map[string]any{
		"as_of_date":       map[string]any{"type": "string", "description": "Current athlete-local date YYYY-MM-DD passed to upstream now to resolve season boundaries; never a UTC timestamp."},
		"season_offsets":   map[string]any{"type": "array", "minItems": 1, "maxItems": maxComparedPowerSeasons, "uniqueItems": true, "items": map[string]any{"type": "integer", "minimum": 0, "maximum": 10}, "description": "Season offsets: 0=current upstream season, 1=previous, up to 10. Defaults to [0,1]; at most four unique offsets. Missing seasons remain explicitly unavailable."},
		"sport":            map[string]any{"type": "string", "default": defaultPowerCurveSport, "description": "intervals.icu sport/type; defaults to Ride. Every compared curve uses this sport."},
		"duration_seconds": map[string]any{"type": "array", "maxItems": maxComparedPowerBuckets, "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Up to 16 duration buckets in seconds, including long efforts. Defaults to 5,15,30,60,300,1200,3600; only upstream values are returned."},
		"include_full":     map[string]any{"type": "boolean", "default": false, "description": "Include raw upstream arrays and activity maps keyed by season selector when true."},
	}}
}
