package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// detectZombieServices returns services with traffic stats
func (c *Client) detectZombieServices(ctx context.Context, namespace string) (map[string]CostStats, error) {
	stats := make(map[string]CostStats)

	// Check if Istio is enabled by looking for the namespace
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, "istio-system", metav1.GetOptions{})
	if err != nil {
		// Istio not found, return empty stats
		return stats, nil
	}

	query := "sum(rate(istio_requests_total[24h])) by (destination_service_name, destination_service_namespace)"
	if namespace != "" {
		query = fmt.Sprintf("sum(rate(istio_requests_total{destination_service_namespace=\"%s\"}[24h])) by (destination_service_name, destination_service_namespace)", namespace)
	}

	result, err := c.queryPrometheus(ctx, query)
	if err != nil {
		return stats, err
	}

	for _, r := range result.Data.Result {
		svcName := r.Metric["destination_service_name"]
		svcNs := r.Metric["destination_service_namespace"]
		if svcName == "" || svcNs == "" {
			continue
		}

		key := fmt.Sprintf("%s/%s", svcNs, svcName)

		var rps float64
		if len(r.Value) >= 2 {
			if rpsStr, ok := r.Value[1].(string); ok {
				fmt.Sscanf(rpsStr, "%f", &rps)
			}
		}

		stats[key] = CostStats{
			RPS:      rps,
			IsZombie: rps < 0.001,
		}
	}

	return stats, nil
}
