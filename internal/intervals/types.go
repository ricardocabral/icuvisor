package intervals

// AthleteWithSportSettings contains the v0.1 athlete profile fields returned by /athlete/{id}.
type AthleteWithSportSettings struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	FirstName             string          `json:"firstname"`
	LastName              string          `json:"lastname"`
	MeasurementPreference string          `json:"measurement_preference"`
	PreferredUnits        string          `json:"preferred_units"`
	WeightPrefLB          bool            `json:"weight_pref_lb"`
	Fahrenheit            bool            `json:"fahrenheit"`
	Timezone              string          `json:"timezone"`
	Locale                string          `json:"locale"`
	SportSettings         []SportSettings `json:"sportSettings"`
}

// SportSettings contains stable threshold and zone fields used by get_athlete_profile.
type SportSettings struct {
	ID             int       `json:"id"`
	AthleteID      string    `json:"athlete_id"`
	Type           string    `json:"type"`
	Types          []string  `json:"types"`
	FTP            int       `json:"ftp"`
	IndoorFTP      int       `json:"indoor_ftp"`
	WPrime         int       `json:"w_prime"`
	PMax           int       `json:"p_max"`
	PowerZones     []int     `json:"power_zones"`
	PowerZoneNames []string  `json:"power_zone_names"`
	FTHR           int       `json:"fthr"`
	LTHR           int       `json:"lthr"`
	MaxHR          int       `json:"max_hr"`
	HRZones        []int     `json:"hr_zones"`
	HRZoneNames    []string  `json:"hr_zone_names"`
	PaceThreshold  float64   `json:"pace_threshold"`
	ThresholdPace  float64   `json:"threshold_pace"`
	PaceUnits      string    `json:"pace_units"`
	PaceLoadType   string    `json:"pace_load_type"`
	WorkoutOrder   string    `json:"workout_order"`
	PaceZones      []float64 `json:"pace_zones"`
	PaceZoneNames  []string  `json:"pace_zone_names"`
}
