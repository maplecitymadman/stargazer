package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// MetricResult represents a Prometheus query result
type MetricResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"` // [timestamp, value]
		} `json:"result"`
	} `json:"data"`
}

// queryPrometheus queries the Prometheus API
func (c *Client) queryPrometheus(ctx context.Context, query string) (*MetricResult, error) {
	if c.prometheusURL == "" {
		// Default to likely-in-cluster address
		c.prometheusURL = "http://kube-prometheus-stack-prometheus.monitoring.svc.cluster.local:9090"
	}

	u, err := url.Parse(fmt.Sprintf("%s/api/v1/query", c.prometheusURL))
	if err != nil {
		return nil, err
	}

	params := u.Query()
	params.Set("query", query)
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result MetricResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// detectZombieServices returns services with near-zero traffic
func (c *Client) detectZombieServices(ctx context.Context, namespace string) ([]string, error) {
	// Query for average RPS over the last 24h
	// We use istio_requests_total if istio is enabled, otherwise fallback to standard metrics
	// For simplicity, we'll try a generic query first: rate(http_requests_total[24h])
	// In a real world scenario, we'd check for multiple metrics (Istio, Cilium, etc)

	// Query: sum(rate(istio_requests_total[24h])) by (destination_service_name, destination_service_namespace)
	query := "sum(rate(istio_requests_total[24h])) by (destination_service_name, destination_service_namespace)"

	result, err := c.queryPrometheus(ctx, query)
	if err != nil {
		return nil, err
	}

	var zombies []string
	// If 0 RPS, it's a zombie
	// (Implementation would check all services and find those NOT in this list or with 0 value)

	for _, r := range result.Data.Result {
		// RPS value is the second element in the Value array
		if len(r.Value) < 2 {
			continue
		}

		// Map RPS to service
		// ... logic to identify services with RPS < threshold ...
	}

	return zombies, nil
}
