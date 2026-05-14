package tools

import (
	"context"
	"encoding/json"
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
}

func (f *fakeCustomItemsClient) ListCustomItems(context.Context) ([]intervals.CustomItem, error) {
	f.listCalls++
	return append([]intervals.CustomItem(nil), f.items...), nil
}

func (f *fakeCustomItemsClient) GetCustomItem(_ context.Context, itemID string) (intervals.CustomItem, error) {
	f.detailCalls = append(f.detailCalls, itemID)
	return f.detail, nil
}

func TestCustomItemsRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if len(registrar.tools) != 6 {
		t.Fatalf("registered tool count = %d, want profile + custom item read/write tools + advanced capabilities", len(registrar.tools))
	}
	listTool := findTool(t, registrar.tools, getCustomItemsName)
	if !strings.Contains(listTool.Description, "id, name, and item_type") {
		t.Fatalf("list description = %q, want terse row language", listTool.Description)
	}
	detailTool := findTool(t, registrar.tools, getCustomItemByIDName)
	if !strings.Contains(detailTool.Description, "icuvisor://custom-item-schemas") {
		t.Fatalf("detail description = %q, want v0.4 resource note", detailTool.Description)
	}
}

func TestGetCustomItemsListsMultipleItemTypeVariants(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
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
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
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
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		detail:            decodeToolCustomItem(t, `{"id":7,"type":"FITNESS_CHART","name":"CTL Chart","content":{"series":[{"field":"ctl","color":"blue"}],"layout":{"height":240}},"from_athlete":{"id":"i999"}}`),
	}
	tool := newGetCustomItemByIDTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_id":"7"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.detailCalls) != 1 || client.detailCalls[0] != "7" {
		t.Fatalf("detail calls = %#v, want item ID 7", client.detailCalls)
	}
	out := resultMap(t, result)
	item := out["custom_item"].(map[string]any)
	if item["id"] != "7" || item["item_type"] != "FITNESS_CHART" {
		t.Fatalf("custom_item identity = %#v, want normalized id/item_type", item)
	}
	content := item["content"].(map[string]any)
	series := content["series"].([]any)[0].(map[string]any)
	if series["field"] != "ctl" || content["layout"].(map[string]any)["height"] != float64(240) {
		t.Fatalf("content = %#v, want verbatim nested payload", content)
	}
	meta := out["_meta"].(map[string]any)
	if meta["content_preserved"] != true || meta["schema_documentation"] != "icuvisor://custom-item-schemas" {
		t.Fatalf("meta = %#v, want content preservation and resource note", meta)
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
