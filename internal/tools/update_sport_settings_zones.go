package tools

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type updateSportSettingsZoneRequest struct {
	Kind       string    `json:"kind"`
	Boundaries []float64 `json:"boundaries"`
	Names      []string  `json:"names,omitempty"`
}

type updateSportSettingsZoneEcho struct {
	Kind       string    `json:"kind"`
	Boundaries []float64 `json:"boundaries"`
	Names      []string  `json:"names,omitempty"`
}

func ensureSportSettingsZonesAllowed(zonesProvided bool, capability safety.Capability) error {
	if !zonesProvided || capability.CanDelete() {
		return nil
	}
	return NewUserError(zoneOverwriteGateMessage, errors.New("zone overwrite requires delete capability"))
}

func validateSportSettingsZones(zones []updateSportSettingsZoneRequest) error {
	if len(zones) == 0 {
		return errors.New("zones must contain at least one zone definition when supplied")
	}
	seen := map[string]bool{}
	for _, zone := range zones {
		kind := normalizeZoneKind(zone.Kind)
		if kind == "" {
			return errors.New("zone kind must be power, hr, or pace")
		}
		if seen[kind] {
			return fmt.Errorf("duplicate %s zone definition", kind)
		}
		seen[kind] = true
		if len(zone.Boundaries) == 0 {
			return fmt.Errorf("%s zone boundaries are required", kind)
		}
		if len(zone.Names) > 0 && len(zone.Names) != len(zone.Boundaries) {
			return fmt.Errorf("%s zone names must match boundaries length", kind)
		}
		for index, boundary := range zone.Boundaries {
			if kind == "pace" {
				if boundary <= 0 || boundary > 200 || math.IsNaN(boundary) || math.IsInf(boundary, 0) {
					return errors.New("pace zone boundaries must be finite percentages in (0, 200]")
				}
				if index > 0 && boundary <= zone.Boundaries[index-1] {
					return errors.New("pace zone boundaries must be strictly increasing percentages")
				}
				continue
			}
			if boundary < 0 {
				return fmt.Errorf("%s zone boundaries must be >= 0", kind)
			}
		}
	}
	return nil
}

func sportSettingsZoneDefinitions(zones []updateSportSettingsZoneRequest) []intervals.SportSettingsZoneDefinition {
	definitions := make([]intervals.SportSettingsZoneDefinition, 0, len(zones))
	for _, zone := range zones {
		definitions = append(definitions, intervals.SportSettingsZoneDefinition{Kind: normalizeZoneKind(zone.Kind), Boundaries: append([]float64(nil), zone.Boundaries...), Names: append([]string(nil), zone.Names...)})
	}
	return definitions
}

func sportSettingsZoneEchoes(zones []intervals.SportSettingsZoneDefinition) []updateSportSettingsZoneEcho {
	echoes := make([]updateSportSettingsZoneEcho, 0, len(zones))
	for _, zone := range zones {
		echoes = append(echoes, updateSportSettingsZoneEcho{Kind: zone.Kind, Boundaries: append([]float64(nil), zone.Boundaries...), Names: append([]string(nil), zone.Names...)})
	}
	return echoes
}

func normalizeZoneKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "power":
		return "power"
	case "hr", "heart_rate", "heart-rate":
		return "hr"
	case "pace":
		return "pace"
	default:
		return ""
	}
}
