package api

import (
	"testing"
)

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantError bool
	}{
		{
			name:      "empty namespace is valid",
			namespace: "",
			wantError: false,
		},
		{
			name:      "all namespaces is valid",
			namespace: "all",
			wantError: false,
		},
		{
			name:      "valid namespace",
			namespace: "default",
			wantError: false,
		},
		{
			name:      "valid namespace with dash",
			namespace: "kube-system",
			wantError: false,
		},
		{
			name:      "valid namespace with numbers",
			namespace: "namespace-123",
			wantError: false,
		},
		{
			name:      "invalid - uppercase",
			namespace: "Default",
			wantError: true,
		},
		{
			name:      "invalid - starts with dash",
			namespace: "-invalid",
			wantError: true,
		},
		{
			name:      "invalid - ends with dash",
			namespace: "invalid-",
			wantError: true,
		},
		{
			name:      "invalid - special characters",
			namespace: "invalid_namespace",
			wantError: true,
		},
		{
			name:      "invalid - too long",
			namespace: string(make([]byte, 254)),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespace(tt.namespace)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateNamespace() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantError bool
	}{
		{
			name:      "valid value",
			value:     "test",
			fieldName: "field",
			wantError: false,
		},
		{
			name:      "empty value",
			value:     "",
			fieldName: "field",
			wantError: true,
		},
		{
			name:      "whitespace only",
			value:     "   ",
			fieldName: "field",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRequired() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateContext(t *testing.T) {
	tests := []struct {
		name      string
		context   string
		wantError bool
	}{
		{
			name:      "empty context is valid",
			context:   "",
			wantError: false,
		},
		{
			name:      "valid context",
			context:   "minikube",
			wantError: false,
		},
		{
			name:      "valid context with special chars",
			context:   "gke_project_us-central1_cluster",
			wantError: false,
		},
		{
			name:      "invalid - too long",
			context:   string(make([]byte, 254)),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContext(tt.context)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateContext() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
