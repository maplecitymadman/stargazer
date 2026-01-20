package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
)

// GetPodLogs retrieves logs from a pod
func (c *Client) GetPodLogs(ctx context.Context, namespace, podName string, tail int, follow bool) (string, error) {
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
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &tailLines,
		Follow:    follow,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	defer stream.Close()

	// Fix Issue #6: Add 10MB limit using io.LimitReader to prevent buffer overflow
	const maxLogSize = 10 * 1024 * 1024 // 10MB limit
	limitedReader := io.LimitReader(stream, maxLogSize)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, limitedReader)
	if err != nil {
		return "", fmt.Errorf("error reading logs: %w", err)
	}

	return buf.String(), nil
}