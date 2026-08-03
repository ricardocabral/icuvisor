package response

var defaultScaleLabels = map[string]string{
	"feel":         "1-5 (athlete-reported feel)",
	"fatigue":      "1-5 (athlete-reported fatigue)",
	"mood":         "1-5 (athlete-reported mood)",
	"motivation":   "1-5 (athlete-reported motivation)",
	"rpe":          "1-10 (rating of perceived exertion)",
	"session_rpe":  "1-10 (session rating of perceived exertion)",
	"sleepQuality": "1-4 (athlete-entered, 1=poor 4=great)",
	"hydration":    "1-4 (athlete-reported hydration)",
	"sleepScore":   "0-100 (device-imported nightly score)",
	"soreness":     "1-5 (athlete-reported soreness)",
	"stress":       "1-5 (athlete-reported stress)",
}

// RegisteredScaleLabels returns the central field-name to scale-label registry.
func RegisteredScaleLabels() map[string]string {
	out := make(map[string]string, len(defaultScaleLabels))
	for field, label := range defaultScaleLabels {
		out[field] = label
	}
	return out
}
