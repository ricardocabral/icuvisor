package tools

import (
	"context"
	"errors"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type activityUnavailable struct {
	ActivityID     string
	StravaImported bool
	Unavailable    *unavailableReason
	Full           map[string]any
}

func classifyActivityReadUnavailable(err error) *unavailableReason {
	switch {
	case errors.Is(err, intervals.ErrNotFound):
		return &unavailableReason{Reason: "not_found"}
	case errors.Is(err, intervals.ErrUnauthorized):
		return &unavailableReason{Reason: "unauthorized"}
	case errors.Is(err, intervals.ErrRateLimited):
		return &unavailableReason{Reason: "rate_limited"}
	default:
		return &unavailableReason{Reason: "upstream_unavailable"}
	}
}

func detectActivityUnavailable(ctx context.Context, detailsClient ActivityDetailsClient, activityID string, originalErr error) (activityUnavailable, error) {
	out := activityUnavailable{ActivityID: activityID, Unavailable: classifyActivityReadUnavailable(originalErr)}
	if !isActivityReadFallbackCandidate(originalErr) || detailsClient == nil {
		return out, nil
	}
	activity, activityErr := detailsClient.GetActivity(ctx, activityID)
	if activityErr == nil && isStravaBlocked(activity) {
		out.ActivityID = firstNonEmpty(activity.ID, activityID)
		out.StravaImported = true
		out.Unavailable = &unavailableReason{Reason: "strava_blocked", Workaround: stravaBlockedWorkaround(activity.Raw)}
		out.Full = activity.Raw
		return out, nil
	}
	if isContextError(activityErr) {
		return activityUnavailable{}, activityErr
	}
	return out, nil
}
