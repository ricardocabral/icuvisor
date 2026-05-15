package tools

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type decodeStrictSample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestDecodeStrict(t *testing.T) {
	tests := []struct {
		name    string
		raw     json.RawMessage
		want    decodeStrictSample
		wantErr string
	}{
		{
			name: "zero input returns zero value",
		},
		{
			name: "whitespace input returns zero value",
			raw:  json.RawMessage(" \n\t "),
		},
		{
			name:    "non object is rejected",
			raw:     json.RawMessage(`[]`),
			wantErr: "arguments must be a JSON object",
		},
		{
			name:    "unknown field is rejected",
			raw:     json.RawMessage(`{"name":"ride","extra":true}`),
			wantErr: `decoding arguments: json: unknown field "extra"`,
		},
		{
			name:    "malformed JSON is rejected",
			raw:     json.RawMessage(`{"name":`),
			wantErr: "decoding arguments:",
		},
		{
			name:    "trailing JSON is rejected",
			raw:     json.RawMessage(`{"name":"ride"} {}`),
			wantErr: "unexpected trailing JSON",
		},
		{
			name: "happy path decodes strictly",
			raw:  json.RawMessage(`{"name":"ride","count":2}`),
			want: decodeStrictSample{Name: "ride", Count: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeStrict[decodeStrictSample](tt.raw)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("DecodeStrict() error = nil, want containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("DecodeStrict() error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("DecodeStrict() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("DecodeStrict() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
