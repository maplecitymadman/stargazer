package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/maplecitymadman/stargazer/internal/k8s"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestHandleHealth tests the health endpoint
func TestHandleHealth(t *testing.T) {
	tests := []struct {
		name           string
		k8sClient      *k8s.Client
		expectedStatus int
		checkBody      func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "without k8s client",
			k8sClient:      nil,
			expectedStatus: http.StatusServiceUnavailable,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				if body["status"] != "unhealthy" {
					t.Errorf("expected status unhealthy, got %v", body["status"])
				}
				if body["cluster"] != "disconnected" {
					t.Errorf("expected cluster disconnected, got %v", body["cluster"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				router:    gin.New(),
				k8sClient: tt.k8sClient,
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/health", nil)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			server.handleHealth(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			tt.checkBody(t, body)
		})
	}
}

// TestHandleGetContexts tests the contexts endpoint
func TestHandleGetContexts(t *testing.T) {
	tests := []struct {
		name           string
		k8sClient      *k8s.Client
		expectedStatus int
		checkBody      func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "without k8s client",
			k8sClient:      nil,
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				contexts, ok := body["contexts"].([]interface{})
				if !ok || len(contexts) != 0 {
					t.Errorf("expected empty contexts array, got %v", body["contexts"])
				}
				if body["total"].(float64) != 0 {
					t.Errorf("expected total 0, got %v", body["total"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				router:    gin.New(),
				k8sClient: tt.k8sClient,
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/contexts", nil)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			server.handleGetContexts(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			tt.checkBody(t, body)
		})
	}
}

// TestHandleGetConfig tests the config endpoint
func TestHandleGetConfig(t *testing.T) {
	server := &Server{
		router: gin.New(),
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/config", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleGetConfig(c)

	// Should return config (may be default if not found)
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500, got %d", w.Code)
	}
}

// TestHandleSwitchContext tests context switching
func TestHandleSwitchContext(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
	}{
		{
			name:           "missing context",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty context",
			requestBody:    map[string]string{"context": ""},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				router: gin.New(),
			}

			body, _ := json.Marshal(tt.requestBody)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			server.handleSwitchContext(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestHandleGetNamespaces tests the namespaces endpoint
func TestHandleGetNamespaces(t *testing.T) {
	server := &Server{
		router:    gin.New(),
		k8sClient: nil, // No client
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/namespaces", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleGetNamespaces(c)

	// Should return 200 with error field when client not initialized
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should have namespaces array (empty)
	if _, ok := body["namespaces"]; !ok {
		t.Error("expected namespaces field in response")
	}
}

// TestHandleGetIssuesEmpty tests the empty issues endpoint
func TestHandleGetIssuesEmpty(t *testing.T) {
	server := &Server{
		router: gin.New(),
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/cluster/issues", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleGetIssuesEmpty(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	issues, ok := body["issues"].([]interface{})
	if !ok || len(issues) != 0 {
		t.Errorf("expected empty issues array, got %v", body["issues"])
	}
}

// TestHandleGetPodsEmpty tests the empty pods endpoint
func TestHandleGetPodsEmpty(t *testing.T) {
	server := &Server{
		router: gin.New(),
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/pods", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleGetPodsEmpty(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	pods, ok := body["pods"].([]interface{})
	if !ok || len(pods) != 0 {
		t.Errorf("expected empty pods array, got %v", body["pods"])
	}
}

// TestHandleGetDeploymentsEmpty tests the empty deployments endpoint
func TestHandleGetDeploymentsEmpty(t *testing.T) {
	server := &Server{
		router: gin.New(),
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/deployments", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleGetDeploymentsEmpty(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	deployments, ok := body["deployments"].([]interface{})
	if !ok || len(deployments) != 0 {
		t.Errorf("expected empty deployments array, got %v", body["deployments"])
	}
}

// TestHandleMetrics tests the metrics endpoint
func TestHandleMetrics(t *testing.T) {
	server := &Server{
		router: gin.New(),
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/metrics", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	server.handleMetrics(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Metrics endpoint returns Prometheus format (text), not JSON
	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty metrics response")
	}

	// Check for Prometheus metrics format
	if !contains(body, "stargazer_requests_total") {
		t.Error("expected stargazer_requests_total metric")
	}
}
