package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func (f *fakeCustomItemsClient) UpdateCustomItem(_ context.Context, params intervals.WriteCustomItemParams) (intervals.CustomItem, error) {
	f.updated = append(f.updated, params)
	if f.updatedItem.Raw != nil {
		return f.updatedItem, nil
	}
	return decodeToolCustomItem(nil, customItemUpdatedJSON(params)), nil
}

func TestUpdateCustomItemUpdatesSingleSparseField(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		updatedItem:       decodeToolCustomItem(t, `{"id":9,"type":"FITNESS_CHART","name":"Renamed","content":{"series":[{"field":"ctl"}],"layout":{"height":240}}}`),
	}
	tool := newUpdateCustomItemTool(client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_id":" 9 ","name":" Renamed "}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.updated) != 1 {
		t.Fatalf("updated calls = %d, want one", len(client.updated))
	}
	params := client.updated[0]
	if params.ItemID != "9" || !params.NameSet || params.Name != "Renamed" || params.ContentSet {
		t.Fatalf("params = %+v, want sparse name-only update", params)
	}
	out := resultMap(t, result)
	item := out["custom_item"].(map[string]any)
	if item["id"] != "9" || item["name"] != "Renamed" || item["content"].(map[string]any)["layout"] == nil {
		t.Fatalf("custom_item = %#v, want full read shape", item)
	}
	meta := out["_meta"].(map[string]any)
	fields := meta["fields_updated"].([]any)
	if meta["operation"] != "update" || len(fields) != 1 || fields[0] != "name" {
		t.Fatalf("meta = %#v, want update fields metadata", meta)
	}
}

func TestUpdateCustomItemDefaultStripsSparseNullsAndPreservesMapValues(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		updatedItem:       decodeToolCustomItem(t, `{"id":9,"type":"FITNESS_CHART","name":"Sparse","description":"","image":null,"index":0,"hide_script":false,"content":{"series":[{"field":"ctl","label":null}],"layout":{"height":240,"note":null}}}`),
	}
	tool := newUpdateCustomItemTool(client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_id":"9","name":"Sparse"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	item := resultMap(t, result)["custom_item"].(map[string]any)
	assertKeyAbsent(t, item, "image")
	if item["description"] != "" || item["index"] != float64(0) || item["hide_script"] != false {
		t.Fatalf("custom_item = %#v, want empty description, index=0, hide_script=false preserved", item)
	}
	content := item["content"].(map[string]any)
	layout := content["layout"].(map[string]any)
	assertKeyAbsent(t, layout, "note")
	series := content["series"].([]any)[0].(map[string]any)
	assertKeyAbsent(t, series, "label")
}

func TestUpdateCustomItemMergesContentPatchAndRejectsSchemaViolation(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
		detail:            decodeToolCustomItem(t, `{"id":9,"type":"FITNESS_CHART","name":"Chart","content":{"series":[{"field":"ctl","color":"blue"}],"layout":{"height":240,"width":600}}}`),
		updatedItem:       decodeToolCustomItem(t, `{"id":9,"type":"FITNESS_CHART","name":"Chart","content":{"series":[{"field":"ctl","color":"blue"}],"layout":{"height":260,"width":600}}}`),
	}
	tool := newUpdateCustomItemTool(client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_id":"9","content":{"layout":{"height":260}}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.detailCalls) != 1 || client.detailCalls[0] != "9" {
		t.Fatalf("detail calls = %#v, want existing item schema lookup", client.detailCalls)
	}
	if len(client.updated) != 1 || !client.updated[0].ContentSet {
		t.Fatalf("updated = %#v, want merged content upload", client.updated)
	}
	layout := client.updated[0].Content["layout"].(map[string]any)
	if layout["height"] != float64(260) || layout["width"] != float64(600) {
		t.Fatalf("merged layout = %#v, want sparse patch preserving width", layout)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["schema_source_count"] != float64(1) {
		t.Fatalf("meta = %#v, want schema source count", meta)
	}

	_, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_id":"9","content":{"layout":{"height":"high"}}}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want schema violation")
	}
	var userErr *UserError
	if !errors.As(err, &userErr) || userErr.Err == nil || !strings.Contains(userErr.Err.Error(), "content.layout.height must be number") {
		t.Fatalf("error = %v, cause = %v, want schema-path validation", err, userErr)
	}
	if len(client.updated) != 1 {
		t.Fatalf("updated calls = %d, want rejection before second upload", len(client.updated))
	}
}

func TestUpdateCustomItemRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newUpdateCustomItemTool(client, client, client, "test", "UTC", false)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	if strings.Contains(strings.ToLower(tool.Description), "confirm") || !strings.Contains(tool.Description, "sparse fields") {
		t.Fatalf("description = %q, want sparse update language and no confirm", tool.Description)
	}
}

func customItemUpdatedJSON(params intervals.WriteCustomItemParams) string {
	payload := map[string]any{"id": params.ItemID, "name": params.Name, "content": params.Content}
	encoded, _ := json.Marshal(payload)
	return string(encoded)
}
