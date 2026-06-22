package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeCustomItemsClient struct {
	fakeProfileClient
	items       []intervals.CustomItem
	detail      intervals.CustomItem
	createdItem intervals.CustomItem
	updatedItem intervals.CustomItem
	listCalls   int
	detailCalls []string
	created     []intervals.WriteCustomItemParams
	updated     []intervals.WriteCustomItemParams
	detailErr   error
	createErr   error
	updateErr   error
}

func (f *fakeCustomItemsClient) ListCustomItems(context.Context) ([]intervals.CustomItem, error) {
	f.listCalls++
	return append([]intervals.CustomItem(nil), f.items...), nil
}

func (f *fakeCustomItemsClient) GetCustomItem(_ context.Context, itemID string) (intervals.CustomItem, error) {
	f.detailCalls = append(f.detailCalls, itemID)
	if f.detailErr != nil {
		return intervals.CustomItem{}, f.detailErr
	}
	return f.detail, nil
}

func TestCustomItemsRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	listTool := newGetCustomItemsTool(client, client, "test", "UTC", false)
	if !strings.Contains(listTool.Description, "id, name, and item_type") {
		t.Fatalf("list description = %q, want terse row language", listTool.Description)
	}
	detailTool := newGetCustomItemByIDTool(client, client, "test", "UTC", false)
	if !strings.Contains(detailTool.Description, "icuvisor://custom-item-schemas") {
		t.Fatalf("detail description = %q, want v0.4 resource note", detailTool.Description)
	}
}

func TestGetCustomItemsListsMultipleItemTypeVariants(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		items: decodeToolCustomItems(t,
			`{"id":2,"type":"ZONES","name":"Run Zones","visibility":"PRIVATE","usage_count":3,"index":2,"content":{"zones":[{"name":"Z1"}]}}`,
			`{"id":1,"type":"FITNESS_CHART","name":"CTL Chart","visibility":"PUBLIC","usage_count":9,"index":1,"content":{"series":[{"field":"ctl"}]}}`,
			`{"id":3,"type":"INPUT_FIELD","name":"Shoe","visibility":"PRIVATE","index":3,"content":{"field":"shoe"}}`,
		),
	}
	tool := newGetCustomItemsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	rows := out["custom_items"].([]any)
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	first := rows[0].(map[string]any)
	if first["item_type"] != "FITNESS_CHART" || first["id"] != "1" || first["name"] != "CTL Chart" {
		t.Fatalf("first row = %#v, want sorted terse chart row", first)
	}
	if _, ok := first["content"]; ok {
		t.Fatalf("list row exposed content: %#v", first)
	}
	meta := out["_meta"].(map[string]any)
	counts := meta["counts_by_item_type"].(map[string]any)
	if counts["FITNESS_CHART"] != float64(1) || counts["INPUT_FIELD"] != float64(1) || counts["ZONES"] != float64(1) {
		t.Fatalf("counts = %#v, want one per item_type", counts)
	}
}

func TestGetCustomItemsFiltersByItemType(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		items: decodeToolCustomItems(t,
			`{"id":1,"type":"FITNESS_CHART","name":"CTL Chart"}`,
			`{"id":3,"type":"INPUT_FIELD","name":"Shoe"}`,
		),
	}
	tool := newGetCustomItemsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_type":"INPUT_FIELD"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["custom_items"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["item_type"] != "INPUT_FIELD" {
		t.Fatalf("filtered rows = %#v, want only INPUT_FIELD", rows)
	}
}

func TestGetCustomItemByIDReturnsFullContentPayload(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		detail:            decodeToolCustomItem(t, `{"id":7,"type":"FITNESS_CHART","name":"CTL Chart","content":{"series":[{"field":"ctl","color":"blue","future_metric":"ramp"}],"layout":{"height":240}},"from_athlete":{"id":"i999"},"future_top_level":{"nested":true}}`),
	}
	tool := newGetCustomItemByIDTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_id":"7"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.detailCalls) != 1 || client.detailCalls[0] != "7" || client.listCalls != 0 {
		t.Fatalf("calls = detail %#v/list %d, want by-ID lookup only", client.detailCalls, client.listCalls)
	}
	out := resultMap(t, result)
	item := out["custom_item"].(map[string]any)
	if item["id"] != "7" || item["item_type"] != "FITNESS_CHART" {
		t.Fatalf("custom_item identity = %#v, want normalized id/item_type", item)
	}
	content := item["content"].(map[string]any)
	series := content["series"].([]any)[0].(map[string]any)
	if series["field"] != "ctl" || series["future_metric"] != "ramp" || content["layout"].(map[string]any)["height"] != float64(240) {
		t.Fatalf("content = %#v, want verbatim nested payload", content)
	}
	if item["future_top_level"].(map[string]any)["nested"] != true || item["from_athlete"].(map[string]any)["id"] != "i999" {
		t.Fatalf("custom_item = %#v, want unknown fields preserved", item)
	}
	meta := out["_meta"].(map[string]any)
	if meta["content_preserved"] != true || meta["schema_documentation"] != "icuvisor://custom-item-schemas" || meta["source_endpoint"] != customItemByIDEndpoint {
		t.Fatalf("meta = %#v, want content preservation, endpoint, and resource note", meta)
	}
}

func TestGetCustomItemByIDTrimsItemIDAndRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		detail:            decodeToolCustomItem(t, `{"id":7,"type":"FITNESS_CHART","name":"CTL Chart"}`),
	}
	tool := newGetCustomItemByIDTool(client, client, "test", "UTC", false)

	if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_id":" 7 "}`)}); err != nil {
		t.Fatalf("Handler() trim error = %v", err)
	}
	if len(client.detailCalls) != 1 || client.detailCalls[0] != "7" {
		t.Fatalf("detail calls = %#v, want trimmed item ID", client.detailCalls)
	}

	tests := []struct {
		name string
		args string
	}{
		{name: "missing object", args: ``},
		{name: "blank item id", args: `{"item_id":"   "}`},
		{name: "unknown argument", args: `{"item_id":"7","query":"fallback"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err == nil {
				t.Fatalf("Handler() error = nil, want user-facing argument error")
			}
			if message, ok := PublicErrorMessage(err); !ok || message != invalidGetCustomItemByIDArgumentsMessage {
				t.Fatalf("PublicErrorMessage(%v) = %q/%v, want %q", err, message, ok, invalidGetCustomItemByIDArgumentsMessage)
			}
		})
	}
}

func TestGetCustomItemByIDReturnsSanitizedPublicErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{name: "not found", err: intervals.ErrNotFound},
		{name: "unauthorized", err: intervals.ErrUnauthorized},
		{name: "rate limited", err: intervals.ErrRateLimited},
		{name: "raw upstream detail", err: errors.New("upstream 401 Authorization: Bearer secret-token")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeCustomItemsClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
				detailErr:         tc.err,
			}
			tool := newGetCustomItemByIDTool(client, client, "test", "UTC", false)

			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_id":"7"}`)})
			if err == nil {
				t.Fatalf("Handler() error = nil, want sanitized public error")
			}
			message, ok := PublicErrorMessage(err)
			if !ok || message != fetchCustomItemByIDMessage {
				t.Fatalf("PublicErrorMessage(%v) = %q/%v, want %q", err, message, ok, fetchCustomItemByIDMessage)
			}
			if strings.Contains(message, "secret-token") || strings.Contains(message, "Authorization") {
				t.Fatalf("public message leaked raw upstream detail: %q", message)
			}
		})
	}
}

func decodeToolCustomItems(t *testing.T, raws ...string) []intervals.CustomItem {
	t.Helper()
	items := make([]intervals.CustomItem, 0, len(raws))
	for _, raw := range raws {
		items = append(items, decodeToolCustomItem(t, raw))
	}
	return items
}

func decodeToolCustomItem(t *testing.T, raw string) intervals.CustomItem {
	t.Helper()
	var item intervals.CustomItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatalf("decode custom item: %v", err)
	}
	return item
}
