package intervals

import "context"

type targetAthleteIDContextKey struct{}

// WithTargetAthleteID returns a context carrying the request-scoped target athlete.
func WithTargetAthleteID(ctx context.Context, athleteID string) context.Context {
	if athleteID == "" {
		return ctx
	}
	return context.WithValue(ctx, targetAthleteIDContextKey{}, athleteID)
}

// TargetAthleteIDFromContext returns a request-scoped target athlete, if one is set.
func TargetAthleteIDFromContext(ctx context.Context) (string, bool) {
	athleteID, ok := ctx.Value(targetAthleteIDContextKey{}).(string)
	return athleteID, ok && athleteID != ""
}

func targetAthleteIDFromContext(ctx context.Context) (string, bool) {
	return TargetAthleteIDFromContext(ctx)
}
