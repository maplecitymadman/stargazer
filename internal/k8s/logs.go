package k8s

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
)

// GetPodLogs retrieves logs from a pod
// GetPodLogs retrieves logs from a pod as a stream
func (c *Client) GetPodLogs(ctx context.Context, namespace, podName string, tail int, follow bool) (io.ReadCloser, error) {
	if namespace == "" {
		namespace = c.namespace
		if namespace == "" {
			namespace = "default"
		}
	}

	// Build log options
	tailLines := int64(tail)
	logOptions := &corev1.PodLogOptions{
		TailLines: &tailLines,
	}

	if follow {
		logOptions.Follow = true
	}

	// Get logs
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)

	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return stream, nil
}
