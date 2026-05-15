package tools

import "encoding/json"

// TextResult returns a text MCP result with the same shaped structured content.
func TextResult(shaped any) Result {
	text, _ := json.Marshal(shaped)
	return Result{Content: []Content{{Type: ContentTypeText, Text: string(text)}}, StructuredContent: shaped}
}
