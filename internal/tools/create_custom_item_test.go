package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func (f *fakeCustomItemsClient) CreateCustomItem(_ context.Context, params intervals.WriteCustomItemParams) (intervals.CustomItem, error) {
	f.created = append(f.created, params)
	if f.createdItem.Raw != nil {
		return f.createdItem, nil
	}
	return decodeToolCustomItem(nil, customItemCreatedJSON(params)), nil
}

func TestCreateCustomItemCreatesPerReadableSchema(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		items: decodeToolCustomItems(t,
			`{"id":1,"type":"FITNESS_CHART","name":"Schema","content":{"series":[{"field":"ctl","color":"blue"}],"layout":{"height":240}}}`,
		),
		createdItem: decodeToolCustomItem(t, `{"id":9,"type":"FITNESS_CHART","name":"New CTL","visibility":"PRIVATE","content":{"series":[{"field":"atl","color":"red"}],"layout":{"height":260}}}`),
	}
	tool := newCreateCustomItemTool(client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_type":"FITNESS_CHART","name":"New CTL","visibility":"PRIVATE","content":{"series":[{"field":"atl","color":"red"}],"layout":{"height":260}}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.listCalls != 1 {
		t.Fatalf("listCalls = %d, want schema lookup before upload", client.listCalls)
	}
	if len(client.created) != 1 {
		t.Fatalf("created calls = %d, want one", len(client.created))
	}
	params := client.created[0]
	if params.ItemType != "FITNESS_CHART" || params.Name != "New CTL" || !params.ContentSet || params.Content["layout"] == nil {
		t.Fatalf("params = %+v, want per-schema create body", params)
	}
	out := resultMap(t, result)
	item := out["custom_item"].(map[string]any)
	if item["id"] != "9" || item["item_type"] != "FITNESS_CHART" || item["content"].(map[string]any)["layout"] == nil {
		t.Fatalf("custom_item = %#v, want full read shape", item)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "create" || meta["schema_source_count"] != float64(1) || meta["content_preserved"] != true {
		t.Fatalf("meta = %#v, want create/schema/read metadata", meta)
	}
}

func TestCreateCustomItemDefaultStripsSparseNullsAndPreservesMapValues(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		items: decodeToolCustomItems(t,
			`{"id":1,"type":"FITNESS_CHART","name":"Schema","content":{"series":[{"field":"ctl","color":"blue"}],"layout":{"height":240}}}`,
		),
		createdItem: decodeToolCustomItem(t, `{"id":9,"type":"FITNESS_CHART","name":"Sparse","description":"","image":null,"index":0,"hide_script":false,"content":{"series":[{"field":"atl","color":"red","label":null}],"layout":{"height":260,"note":null}}}`),
	}
	tool := newCreateCustomItemTool(client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_type":"FITNESS_CHART","name":"Sparse","description":"","index":0,"hide_script":false,"content":{"series":[{"field":"atl","color":"red"}],"layout":{"height":260}}}`)})
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

func TestCreateCustomItemFetchesDetailWhenListOmitsContentSchema(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		items:             decodeToolCustomItems(t, `{"id":1,"type":"FITNESS_CHART","name":"Schema"}`),
		detail:            decodeToolCustomItem(t, `{"id":1,"type":"FITNESS_CHART","name":"Schema","content":{"series":[{"field":"ctl","color":"blue"}],"layout":{"height":240}}}`),
		createdItem:       decodeToolCustomItem(t, `{"id":9,"type":"FITNESS_CHART","name":"New CTL","content":{"series":[{"field":"atl","color":"red"}],"layout":{"height":260}}}`),
	}
	tool := newCreateCustomItemTool(client, client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_type":"FITNESS_CHART","name":"New CTL","content":{"series":[{"field":"atl","color":"red"}],"layout":{"height":260}}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if got := strings.Join(client.detailCalls, ","); got != "1" {
		t.Fatalf("detailCalls = %q, want schema detail fetch for list row without content", got)
	}
	if len(client.created) != 1 {
		t.Fatalf("created calls = %d, want one", len(client.created))
	}
}

func TestCreateCustomItemRejectsSchemaViolationsBeforeUpload(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
		items: decodeToolCustomItems(t,
			`{"id":1,"type":"FITNESS_CHART","name":"Schema","content":{"series":[{"field":"ctl"}],"layout":{"height":240}}}`,
		),
	}
	tool := newCreateCustomItemTool(client, client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"item_type":"FITNESS_CHART","name":"Bad","content":{"series":[{"field":"ctl"}],"layout":{"height":"tall"}}}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want schema violation")
	}
	var userErr *UserError
	if !errors.As(err, &userErr) || userErr.Err == nil || !strings.Contains(userErr.Err.Error(), "content.layout.height must be number") {
		t.Fatalf("error = %v, cause = %v, want schema-path validation", err, userErr)
	}
	if len(client.created) != 0 {
		t.Fatalf("created calls = %d, want validation before upload", len(client.created))
	}
}

func TestCreateCustomItemRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeCustomItemsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newCreateCustomItemTool(client, client, client, "test", "UTC", false)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	if strings.Contains(strings.ToLower(tool.Description), "confirm") || !strings.Contains(tool.Description, "validated against readable samples") || !strings.Contains(tool.Description, "icuvisor://custom-item-schemas") {
		t.Fatalf("description = %q, want validation/resource language and no confirm", tool.Description)
	}
}

func customItemCreatedJSON(params intervals.WriteCustomItemParams) string {
	payload := map[string]any{"id": "created", "type": params.ItemType, "name": params.Name, "content": params.Content}
	encoded, _ := json.Marshal(payload)
	return string(encoded)
}
