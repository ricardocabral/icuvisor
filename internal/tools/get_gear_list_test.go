package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeGearListClient struct {
	gearByTarget map[string][]intervals.Gear
	gear         []intervals.Gear
	err          error
	calls        int
	targets      []string
}

func (f *fakeGearListClient) ListGear(ctx context.Context) ([]intervals.Gear, error) {
	f.calls++
	target, _ := intervals.TargetAthleteIDFromContext(ctx)
	f.targets = append(f.targets, target)
	if f.err != nil {
		return nil, f.err
	}
	if f.gearByTarget != nil {
		return f.gearByTarget[target], nil
	}
	return f.gear, nil
}

func TestGetGearListReturnsTerseRowsAndMeta(t *testing.T) {
	t.Parallel()

	client := &fakeGearListClient{gear: decodeToolGear(t,
		`{"id":"g-1","name":"Race Bike","type":"Bike","brand":"Cervelo","model":"S5","retired":false}`,
		`{"id":"g-2","type":"Shoes","retired":true}`,
	)}
	tool := newGetGearListTool(client, newGearListCache(), "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	rows := payload["gear"].([]any)
	if len(rows) != 2 {
		t.Fatalf("gear rows len = %d, want 2", len(rows))
	}
	first := rows[0].(map[string]any)
	if first["gear_id"] != "g-1" || first["name"] != "Race Bike" || first["type"] != "Bike" || first["full"] != nil {
		t.Fatalf("first row = %#v, want terse named gear", first)
	}
	second := rows[1].(map[string]any)
	if second["gear_id"] != "g-2" || second["name_missing"] != true {
		t.Fatalf("second row = %#v, want explicit missing name", second)
	}
	meta := payload["_meta"].(map[string]any)
	if meta["count"] != float64(2) || meta["unnamed_count"] != float64(1) || meta["refreshed"] != true || meta["cached"] != false || meta["include_full"] != false {
		t.Fatalf("meta = %#v, want counts and refreshed state", meta)
	}
}

func TestGetGearListIncludeFull(t *testing.T) {
	t.Parallel()

	client := &fakeGearListClient{gear: decodeToolGear(t, `{"id":"g-1","name":"Race Bike","custom":"kept"}`)}
	tool := newGetGearListTool(client, newGearListCache(), "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["gear"].([]any)[0].(map[string]any)
	full := row["full"].(map[string]any)
	if full["custom"] != "kept" {
		t.Fatalf("full = %#v, want raw upstream fields", full)
	}
}

func TestGetGearListNormalizesAthleteCacheKey(t *testing.T) {
	t.Parallel()

	client := &fakeGearListClient{gear: decodeToolGear(t, `{"id":"g-1","name":"Race Bike"}`)}
	tool := newGetGearListTool(client, newGearListCache(), "test", false)

	ctxUpper := intervals.WithTargetAthleteID(context.Background(), "I12345")
	if _, err := tool.Handler(ctxUpper, Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)}); err != nil {
		t.Fatalf("first Handler() error = %v", err)
	}
	ctxNormalized := intervals.WithTargetAthleteID(context.Background(), "i12345")
	result, err := tool.Handler(ctxNormalized, Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("second Handler() error = %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("ListGear calls = %d, want normalized athlete IDs to share one cache entry", client.calls)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["cached"] != true || meta["refreshed"] != false {
		t.Fatalf("meta = %#v, want cached second response", meta)
	}
}

func TestGetGearListCacheIsPerTargetAthleteAndRefreshable(t *testing.T) {
	t.Parallel()

	client := &fakeGearListClient{gearByTarget: map[string][]intervals.Gear{
		"i111": decodeToolGear(t, `{"id":"g-111","name":"A Bike"}`),
		"i222": decodeToolGear(t, `{"id":"g-222","name":"B Bike"}`),
	}}
	tool := newGetGearListTool(client, newGearListCache(), "test", false)
	ctx111 := intervals.WithTargetAthleteID(context.Background(), "i111")
	ctx222 := intervals.WithTargetAthleteID(context.Background(), "i222")

	if _, err := tool.Handler(ctx111, Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)}); err != nil {
		t.Fatalf("athlete 111 first Handler() error = %v", err)
	}
	if _, err := tool.Handler(ctx222, Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)}); err != nil {
		t.Fatalf("athlete 222 first Handler() error = %v", err)
	}
	cached111, err := tool.Handler(ctx111, Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("athlete 111 cached Handler() error = %v", err)
	}
	if client.calls != 2 {
		t.Fatalf("ListGear calls = %d, want separate fetches for two target athletes and cached third call", client.calls)
	}
	row := resultMap(t, cached111)["gear"].([]any)[0].(map[string]any)
	if row["gear_id"] != "g-111" {
		t.Fatalf("cached athlete 111 row = %#v, want no cross-athlete cache leak", row)
	}

	refreshed, err := tool.Handler(ctx111, Request{Name: tool.Name, Arguments: json.RawMessage(`{"refresh":true}`)})
	if err != nil {
		t.Fatalf("refresh Handler() error = %v", err)
	}
	if client.calls != 3 {
		t.Fatalf("ListGear calls = %d, want refresh to bypass cache", client.calls)
	}
	meta := resultMap(t, refreshed)["_meta"].(map[string]any)
	if meta["refreshed"] != true || meta["cached"] != false {
		t.Fatalf("refresh meta = %#v, want refreshed response", meta)
	}
}

func TestGetGearListFailedRefreshDoesNotReplaceCachedEntry(t *testing.T) {
	t.Parallel()

	client := &fakeGearListClient{gear: decodeToolGear(t, `{"id":"g-1","name":"Race Bike"}`)}
	tool := newGetGearListTool(client, newGearListCache(), "test", false)
	if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)}); err != nil {
		t.Fatalf("initial Handler() error = %v", err)
	}
	client.err = errors.New("upstream down")
	if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"refresh":true}`)}); err == nil {
		t.Fatal("refresh Handler() error = nil, want error")
	}
	client.err = nil
	client.gear = decodeToolGear(t, `{"id":"g-2","name":"Replacement"}`)
	cached, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("cached Handler() error = %v", err)
	}
	row := resultMap(t, cached)["gear"].([]any)[0].(map[string]any)
	if row["gear_id"] != "g-1" {
		t.Fatalf("cached row = %#v, want failed refresh to preserve previous cache entry", row)
	}
}

func TestGetGearListEmptyList(t *testing.T) {
	t.Parallel()

	client := &fakeGearListClient{}
	tool := newGetGearListTool(client, newGearListCache(), "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	if len(payload["gear"].([]any)) != 0 || payload["_meta"].(map[string]any)["count"] != float64(0) {
		t.Fatalf("payload = %#v, want empty gear list with count 0", payload)
	}
}

func decodeToolGear(t *testing.T, raws ...string) []intervals.Gear {
	t.Helper()
	gear := make([]intervals.Gear, 0, len(raws))
	for _, raw := range raws {
		var item intervals.Gear
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			t.Fatalf("unmarshal gear fixture: %v", err)
		}
		gear = append(gear, item)
	}
	return gear
}
