package tools

import (
	"encoding/json"
	"reflect"
	"testing"
)

type textResultSample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestTextResult(t *testing.T) {
	tests := []struct {
		name   string
		shaped any
	}{
		{
			name:   "struct shaped content",
			shaped: textResultSample{Name: "ride", Count: 2},
		},
		{
			name: "map shaped content",
			shaped: map[string]any{
				"name":  "ride",
				"count": float64(2),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, _ := json.Marshal(tt.shaped)
			want := Result{Content: []Content{{Type: ContentTypeText, Text: string(text)}}, StructuredContent: tt.shaped}

			got := TextResult(tt.shaped)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("TextResult() = %#v, want %#v", got, want)
			}
			if len(got.Content) != 1 {
				t.Fatalf("content count = %d, want 1", len(got.Content))
			}
			if got.Content[0].Text != string(text) {
				t.Fatalf("content text = %q, want %q", got.Content[0].Text, string(text))
			}
			if !reflect.DeepEqual(got.StructuredContent, tt.shaped) {
				t.Fatalf("structured content = %#v, want %#v", got.StructuredContent, tt.shaped)
			}
		})
	}
}
