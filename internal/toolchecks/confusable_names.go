package toolchecks

import (
	"context"
	"sort"
	"strings"
	"unicode"
)

const DefaultConfusableThreshold = 0.58

type ToolInfo struct {
	Name        string
	Description string
}

type ConfusableReport struct {
	Pairs []ConfusablePair
}

func (r ConfusableReport) OK() bool { return len(r.Pairs) == 0 }

type ConfusablePair struct {
	Cluster   string
	ToolA     string
	ToolB     string
	Score     float64
	SentenceA string
	SentenceB string
}

func GenerateToolCatalog(ctx context.Context) ([]ToolInfo, error) {
	toolCatalog, err := generateSchemaCatalogTools(ctx)
	if err != nil {
		return nil, err
	}

	catalog := make([]ToolInfo, 0, len(toolCatalog))
	for _, tool := range toolCatalog {
		catalog = append(catalog, ToolInfo{Name: tool.Name, Description: tool.Description})
	}
	sort.Slice(catalog, func(i, j int) bool { return catalog[i].Name < catalog[j].Name })
	return catalog, nil
}

func CheckConfusableCatalog(catalog []ToolInfo, threshold float64) ConfusableReport {
	clusters := map[string][]ToolInfo{}
	for _, tool := range catalog {
		cluster := toolCluster(tool.Name)
		if cluster == "" {
			continue
		}
		clusters[cluster] = append(clusters[cluster], tool)
	}
	report := ConfusableReport{}
	for _, clusterName := range sortedToolClusters(clusters) {
		cluster := clusters[clusterName]
		if len(cluster) < 2 {
			continue
		}
		sort.Slice(cluster, func(i, j int) bool { return cluster[i].Name < cluster[j].Name })
		for i := 0; i < len(cluster); i++ {
			for j := i + 1; j < len(cluster); j++ {
				sentenceA := FirstDescriptionSentence(cluster[i].Description)
				sentenceB := FirstDescriptionSentence(cluster[j].Description)
				score := tokenJaccard(sentenceA, sentenceB)
				if score >= threshold {
					report.Pairs = append(report.Pairs, ConfusablePair{Cluster: clusterName, ToolA: cluster[i].Name, ToolB: cluster[j].Name, Score: score, SentenceA: sentenceA, SentenceB: sentenceB})
				}
			}
		}
	}
	return report
}

func FirstDescriptionSentence(description string) string {
	description = strings.TrimSpace(description)
	for i, r := range description {
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		if r == '.' && hasAdjacentWordRunes(description, i) {
			continue
		}
		next := i + len(string(r))
		if next >= len(description) || unicode.IsSpace(rune(description[next])) {
			return strings.TrimSpace(description[:next])
		}
	}
	return description
}

func hasAdjacentWordRunes(value string, dotIndex int) bool {
	if dotIndex == 0 || dotIndex+1 >= len(value) {
		return false
	}
	before := rune(value[dotIndex-1])
	after := rune(value[dotIndex+1])
	return isTokenRune(before) && isTokenRune(after)
}

func tokenJaccard(a string, b string) float64 {
	aTokens := tokenSet(a)
	bTokens := tokenSet(b)
	if len(aTokens) == 0 && len(bTokens) == 0 {
		return 1
	}
	intersection := 0
	for token := range aTokens {
		if _, ok := bTokens[token]; ok {
			intersection++
		}
	}
	union := len(aTokens) + len(bTokens) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func tokenSet(value string) map[string]struct{} {
	var builder strings.Builder
	for _, r := range strings.ToLower(value) {
		if isTokenRune(r) {
			builder.WriteRune(r)
		} else {
			builder.WriteRune(' ')
		}
	}
	out := map[string]struct{}{}
	for _, token := range strings.Fields(builder.String()) {
		if _, skip := confusableStopWords[token]; skip {
			continue
		}
		out[token] = struct{}{}
	}
	return out
}

func isTokenRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

var confusableStopWords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "by": {}, "for": {}, "from": {}, "get": {}, "in": {}, "inside": {}, "is": {}, "list": {}, "of": {}, "one": {}, "or": {}, "the": {}, "to": {}, "with": {},
}

func toolCluster(name string) string {
	switch {
	case strings.HasPrefix(name, "analyze_") || strings.HasPrefix(name, "compute_") || name == "get_activity_histogram" || name == "get_fitness_projection":
		return "analyzers"
	case name == "get_activities" || strings.HasPrefix(name, "get_activity_") || name == "get_extended_metrics":
		return "activity"
	case strings.HasPrefix(name, "get_event") || name == "get_training_plan":
		return "event-calendar"
	case strings.HasPrefix(name, "get_workout"):
		return "workout-library"
	case strings.HasPrefix(name, "get_custom_item"):
		return "custom-items"
	case name == "get_fitness" || name == "get_training_summary" || name == "get_best_efforts" || name == "get_power_curves":
		return "fitness-performance"
	case strings.HasPrefix(name, "get_wellness"):
		return "wellness"
	default:
		return ""
	}
}

func sortedToolClusters(clusters map[string][]ToolInfo) []string {
	names := make([]string, 0, len(clusters))
	for name := range clusters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
