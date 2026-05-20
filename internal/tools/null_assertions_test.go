package tools

import "testing"

func assertKeyAbsent(t *testing.T, object map[string]any, key string) {
	t.Helper()
	if _, ok := object[key]; ok {
		t.Fatalf("%s present in %#v, want absent", key, object)
	}
}

func assertKeyPresentNil(t *testing.T, object map[string]any, key string) {
	t.Helper()
	value, ok := object[key]
	if !ok {
		t.Fatalf("%s absent from %#v, want present with null", key, object)
	}
	if value != nil {
		t.Fatalf("%s = %#v, want null", key, value)
	}
}
