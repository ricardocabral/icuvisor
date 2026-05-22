package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type activitiesPageToken struct {
	Version              int      `json:"v"`
	Oldest               string   `json:"oldest"`
	Newest               string   `json:"newest,omitempty"`
	RouteID              int64    `json:"route_id,omitempty"`
	IncludeUnnamed       bool     `json:"include_unnamed"`
	IncludeFull          bool     `json:"include_full"`
	PageSize             int      `json:"page_size"`
	AthleteID            string   `json:"athlete_id,omitempty"`
	Fields               []string `json:"fields,omitempty"`
	BeforeStartDateLocal string   `json:"before_start_date_local,omitempty"`
	BeforeID             string   `json:"before_id,omitempty"`
	SkipIDsAtBoundary    []string `json:"skip_ids_at_boundary,omitempty"`
}

func decodeGetActivitiesRequest(raw json.RawMessage) (GetActivitiesRequest, *activitiesPageToken, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return GetActivitiesRequest{}, nil, errors.New("oldest is required unless next_page_token is supplied")
	}
	if trimmed[0] != '{' {
		return GetActivitiesRequest{}, nil, errors.New("arguments must be a JSON object")
	}
	args, err := DecodeStrict[GetActivitiesRequest](trimmed)
	if err != nil {
		return GetActivitiesRequest{}, nil, err
	}
	var supplied activitiesTokenArgs
	if err := json.Unmarshal(trimmed, &supplied); err != nil {
		return GetActivitiesRequest{}, nil, err
	}
	args.PageSize = normalizeActivitiesPageSize(args.PageSize)
	if strings.TrimSpace(args.NextPageToken) == "" {
		if strings.TrimSpace(args.Oldest) == "" {
			return GetActivitiesRequest{}, nil, errors.New("oldest is required")
		}
		return args, nil, nil
	}
	token, err := parseActivitiesPageToken(args.NextPageToken)
	if err != nil {
		return GetActivitiesRequest{}, nil, err
	}
	if err := validateActivitiesTokenArgs(args, token, supplied); err != nil {
		return GetActivitiesRequest{}, nil, err
	}
	args.Oldest = token.Oldest
	args.Newest = token.Newest
	args.RouteID = token.RouteID
	args.IncludeUnnamed = token.IncludeUnnamed
	args.IncludeFull = token.IncludeFull
	args.PageSize = token.PageSize
	return args, token, nil
}

func normalizeActivitiesPageSize(pageSize int) int {
	if pageSize <= 0 {
		return defaultActivitiesPageSize
	}
	if pageSize > maxActivitiesPageSize {
		return maxActivitiesPageSize
	}
	return pageSize
}

type activitiesTokenArgs struct {
	Oldest         *string `json:"oldest"`
	Newest         *string `json:"newest"`
	RouteID        *int64  `json:"route_id"`
	IncludeUnnamed *bool   `json:"include_unnamed"`
	PageSize       *int    `json:"page_size"`
	IncludeFull    *bool   `json:"include_full"`
}

func validateActivitiesTokenArgs(args GetActivitiesRequest, token *activitiesPageToken, supplied activitiesTokenArgs) error {
	if token.Version != 1 {
		return fmt.Errorf("%w: unsupported token version %d", ErrInvalidInput, token.Version)
	}
	if token.PageSize <= 0 || token.PageSize > maxActivitiesPageSize || strings.TrimSpace(token.Oldest) == "" {
		return errors.New("invalid token payload")
	}
	if supplied.Oldest != nil && strings.TrimSpace(args.Oldest) != strings.TrimSpace(token.Oldest) {
		return errors.New("oldest does not match next_page_token")
	}
	if supplied.Newest != nil && strings.TrimSpace(args.Newest) != strings.TrimSpace(token.Newest) {
		return errors.New("newest does not match next_page_token")
	}
	if supplied.RouteID != nil && args.RouteID != token.RouteID {
		return errors.New("route_id does not match next_page_token")
	}
	if supplied.PageSize != nil && args.PageSize != token.PageSize {
		return errors.New("page_size does not match next_page_token")
	}
	if supplied.IncludeUnnamed != nil && args.IncludeUnnamed != token.IncludeUnnamed {
		return errors.New("include_unnamed does not match next_page_token")
	}
	if supplied.IncludeFull != nil && args.IncludeFull != token.IncludeFull {
		return errors.New("include_full does not match next_page_token")
	}
	return nil
}

func parseActivitiesPageToken(value string) (*activitiesPageToken, error) {
	data, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("decoding next_page_token: %w", err)
	}
	var token activitiesPageToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("decoding next_page_token JSON: %w", err)
	}
	return &token, nil
}

func encodeActivitiesPageToken(token activitiesPageToken) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func fetchActivitiesPage(ctx context.Context, client ActivitiesClient, args GetActivitiesRequest, token *activitiesPageToken, targetAthleteID string, customFieldCodes []string) ([]intervals.Activity, string, error) {
	cursor := newPageCursor(args, token, targetAthleteID, customFieldCodes)
	page := make([]intervals.Activity, 0, args.PageSize)
	for {
		candidates, done, err := iteratePages(ctx, client, args, &cursor)
		if err != nil {
			return nil, "", err
		}
		if done {
			break
		}
		for _, activity := range candidates {
			if !args.IncludeUnnamed && strings.TrimSpace(stringValue(activity.Name)) == "" && !isStravaBlocked(activity) {
				cursor.advancePast(activity)
				continue
			}
			if len(page) == args.PageSize {
				nextToken, err := cursor.encodeTokenIfAdvanced()
				return page, nextToken, err
			}
			page = append(page, activity)
			cursor.advancePast(activity)
		}
		if !cursor.fullWindow {
			break
		}
		if len(page) == args.PageSize {
			nextToken, err := cursor.encodeTokenIfAdvanced()
			return page, nextToken, err
		}
	}
	if cursor.fullWindow && cursor.advanced {
		nextToken, err := encodeActivitiesPageToken(cursor.token)
		if err != nil {
			return nil, "", err
		}
		return page, nextToken, nil
	}
	return page, "", nil
}

func iteratePages(ctx context.Context, client ActivitiesClient, args GetActivitiesRequest, cursor *pageCursor) ([]intervals.Activity, bool, error) {
	for cursor.canFetch() {
		cursor.beginIteration()
		newest := effectiveNewest(args.Newest, cursor.token)
		if activityBoundaryBefore(newest, args.Oldest) {
			cursor.fullWindow = false
			return nil, true, nil
		}
		params := intervals.ListActivitiesParams{Oldest: args.Oldest, Newest: newest, RouteID: args.RouteID, Limit: cursor.fetchLimit}
		if !args.IncludeFull {
			params.Fields = cursor.token.Fields
		}
		activities, err := client.ListActivities(ctx, params)
		cursor.fetches++
		if err != nil {
			return nil, true, err
		}
		if len(activities) == 0 {
			cursor.fullWindow = false
			return nil, true, nil
		}
		cursor.fullWindow = len(activities) == cursor.fetchLimit
		sortActivities(activities)
		candidates := activitiesAfterCursor(activities, cursor.token)
		if len(candidates) > 0 {
			return candidates, false, nil
		}
		if !cursor.advancePast(activities[len(activities)-1]) && cursor.fullWindow && cursor.fetchLimit < maxActivityFetchLimit {
			cursor.fetchLimit = maxActivityFetchLimit
			continue
		}
		if !cursor.advancedThisIteration && cursor.fullWindow && cursor.fetchLimit >= maxActivityFetchLimit {
			return nil, true, errActivitiesPaginationBoundary
		}
		if !cursor.advancedThisIteration {
			cursor.advanceBeforeBoundary()
		}
		if !cursor.advancedThisIteration {
			cursor.fullWindow = false
			return nil, true, nil
		}
	}
	return nil, true, nil
}

type pageCursor struct {
	token                 activitiesPageToken
	fetchLimit            int
	fetches               int
	fullWindow            bool
	advanced              bool
	advancedThisIteration bool
}

func newPageCursor(args GetActivitiesRequest, token *activitiesPageToken, targetAthleteID string, customFieldCodes []string) pageCursor {
	cursor := pageCursor{
		token:      activitiesPageToken{Version: 1, Oldest: args.Oldest, Newest: args.Newest, RouteID: args.RouteID, IncludeUnnamed: args.IncludeUnnamed, IncludeFull: args.IncludeFull, PageSize: args.PageSize, AthleteID: targetAthleteID},
		fetchLimit: min(args.PageSize*2+1, maxActivityFetchLimit),
	}
	if !args.IncludeFull {
		cursor.token.Fields = terseActivityFieldsWithCustom(customFieldCodes)
	}
	if token != nil {
		cursor.token.BeforeStartDateLocal = token.BeforeStartDateLocal
		cursor.token.BeforeID = token.BeforeID
		cursor.token.SkipIDsAtBoundary = append([]string(nil), token.SkipIDsAtBoundary...)
		cursor.token.Fields = append([]string(nil), token.Fields...)
	}
	return cursor
}

func (c *pageCursor) canFetch() bool {
	return c.fetches < maxActivityPageFetches
}

func (c *pageCursor) beginIteration() {
	c.advancedThisIteration = false
}

func (c *pageCursor) advancePast(activity intervals.Activity) bool {
	return c.markAdvanced(advanceCursorPast(&c.token, activity))
}

func (c *pageCursor) advanceBeforeBoundary() bool {
	return c.markAdvanced(advanceCursorBeforeBoundary(&c.token))
}

func (c *pageCursor) markAdvanced(advanced bool) bool {
	if advanced {
		c.advanced = true
		c.advancedThisIteration = true
	}
	return advanced
}

func (c pageCursor) encodeTokenIfAdvanced() (string, error) {
	if !c.advanced {
		return "", nil
	}
	return encodeActivitiesPageToken(c.token)
}

func effectiveNewest(newest string, cursor activitiesPageToken) string {
	if cursor.BeforeStartDateLocal != "" {
		return cursor.BeforeStartDateLocal
	}
	return newest
}

func sortActivities(activities []intervals.Activity) {
	sort.SliceStable(activities, func(i, j int) bool {
		leftDate := activitySortDate(activities[i])
		rightDate := activitySortDate(activities[j])
		if leftDate != rightDate {
			return leftDate > rightDate
		}
		return activities[i].ID > activities[j].ID
	})
}

func activitiesAfterCursor(activities []intervals.Activity, cursor activitiesPageToken) []intervals.Activity {
	if cursor.BeforeStartDateLocal == "" {
		return activities
	}
	out := make([]intervals.Activity, 0, len(activities))
	skips := make(map[string]bool, len(cursor.SkipIDsAtBoundary))
	for _, id := range cursor.SkipIDsAtBoundary {
		skips[id] = true
	}
	for _, activity := range activities {
		date := activitySortDate(activity)
		if date > cursor.BeforeStartDateLocal {
			continue
		}
		if date == cursor.BeforeStartDateLocal {
			if skips[activity.ID] || (cursor.BeforeID != "" && activity.ID >= cursor.BeforeID) {
				continue
			}
		}
		out = append(out, activity)
	}
	return out
}

func advanceCursorBeforeBoundary(cursor *activitiesPageToken) bool {
	if cursor.BeforeStartDateLocal == "" {
		return false
	}
	before := justBeforeActivityTimestamp(cursor.BeforeStartDateLocal)
	if before == "" || before == cursor.BeforeStartDateLocal {
		return false
	}
	cursor.BeforeStartDateLocal = before
	cursor.BeforeID = ""
	cursor.SkipIDsAtBoundary = nil
	return true
}

func activityBoundaryBefore(left string, right string) bool {
	leftTime, leftOK := parseActivityBoundary(left)
	rightTime, rightOK := parseActivityBoundary(right)
	return leftOK && rightOK && leftTime.Before(rightTime)
}

func parseActivityBoundary(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}
	layouts := []string{"2006-01-02T15:04:05", "2006-01-02T15:04", time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func justBeforeActivityTimestamp(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	layouts := []string{"2006-01-02T15:04:05", "2006-01-02T15:04", time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err != nil {
			continue
		}
		if layout == "2006-01-02" {
			return parsed.AddDate(0, 0, -1).Format(layout)
		}
		return parsed.Add(-time.Second).Format(layout)
	}
	return ""
}

func advanceCursorPast(cursor *activitiesPageToken, activity intervals.Activity) bool {
	date := activitySortDate(activity)
	if date == "" {
		return false
	}
	if cursor.BeforeStartDateLocal != date {
		cursor.BeforeStartDateLocal = date
		cursor.BeforeID = activity.ID
		cursor.SkipIDsAtBoundary = []string{activity.ID}
		return true
	}
	if slices.Contains(cursor.SkipIDsAtBoundary, activity.ID) {
		return false
	}
	cursor.BeforeID = activity.ID
	cursor.SkipIDsAtBoundary = append(cursor.SkipIDsAtBoundary, activity.ID)
	return true
}

func activitySortDate(activity intervals.Activity) string {
	if value := stringValue(activity.StartDateLocal); value != "" {
		return value
	}
	return stringValue(activity.StartDate)
}
