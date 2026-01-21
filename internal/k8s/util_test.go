package k8s

import (
	"testing"
)

func TestToJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	jsonStr := toJSON(data)
	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}
	if jsonStr == "{}" {
		t.Error("Expected populated JSON string")
	}
}
