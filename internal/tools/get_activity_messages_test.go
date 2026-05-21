package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func (f *fakeActivityReadClient) GetActivityMessages(ctx context.Context, params intervals.ActivityMessagesParams) ([]intervals.ActivityMessage, error) {
	return f.messages, f.messageErr
}

func TestGetActivityMessagesTerseAndTimezone(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "America/Sao_Paulo"}}, messages: decodeMessageFixtures(t, `{"id":1,"name":"Coach","created":"2026-01-02T12:00:00Z","type":"comment","content":"nice","activity_id":"a1","deleted":false,"seen":false,"extra":null}`)}
	tool := newGetActivityMessagesTool(client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	if got := payload["_meta"].(map[string]any)["limit"]; got != float64(defaultActivityMessagesLimit) {
		t.Fatalf("limit meta = %v, want default", got)
	}
	row := payload["messages"].([]any)[0].(map[string]any)
	if got := row["created"]; got != "2026-01-02T09:00:00-03:00" {
		t.Fatalf("created = %v, want rendered Sao Paulo time", got)
	}
	if row["deleted"] != false || row["seen"] != false {
		t.Fatalf("row = %#v, want explicit false booleans preserved", row)
	}
	if value, ok := row["full"].(map[string]any)["extra"]; !ok || value != nil {
		t.Fatalf("full extra = %#v present %v, want preserved nil", value, ok)
	}
}

func TestGetActivityMessagesRejectsNegativeSinceID(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newGetActivityMessagesTool(client, client, client, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","since_id":-1}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != invalidActivityReadArgumentsMessage {
		t.Fatalf("PublicErrorMessage = %q, %v; err = %v", message, ok, err)
	}
}

func TestGetActivityMessagesRejectsInvalidLimit(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newGetActivityMessagesTool(client, client, client, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","limit":999}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != invalidActivityReadArgumentsMessage {
		t.Fatalf("PublicErrorMessage = %q, %v; err = %v", message, ok, err)
	}
}

func TestGetActivityMessagesPreservesCancellation(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}, messageErr: context.Canceled}
	tool := newGetActivityMessagesTool(client, client, client, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1"}`)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handler() error = %v, want context.Canceled", err)
	}
}

func TestGetActivityMessagesPropagatesFallbackCancellation(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}, messageErr: intervals.ErrNotFound, activityErr: context.Canceled}
	tool := newGetActivityMessagesTool(client, client, client, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handler() error = %v, want context.Canceled", err)
	}
}

func TestGetActivityMessagesFallbacksToStravaUnavailable(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}, activity: decodeActivityFixture(t, `{"id":"stub1","source":"Strava","_note":"hidden","external_id":"wahoo-synthetic-12345"}`), messageErr: intervals.ErrNotFound}
	tool := newGetActivityMessagesTool(client, client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	assertUnavailableReasonAndWorkaround(t, payload, "strava_tos", wantWahooStravaWorkaround)
}

func decodeMessageFixtures(t *testing.T, raws ...string) []intervals.ActivityMessage {
	t.Helper()
	out := make([]intervals.ActivityMessage, 0, len(raws))
	for _, raw := range raws {
		var message intervals.ActivityMessage
		if err := json.Unmarshal([]byte(raw), &message); err != nil {
			t.Fatalf("decode message fixture: %v", err)
		}
		out = append(out, message)
	}
	return out
}
