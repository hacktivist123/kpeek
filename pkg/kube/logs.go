package kube

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ContainerLog holds logs for a single container.
type ContainerLog struct {
	PodName       string
	ContainerName string
	Logs          string
}

// FetchPodLogs retrieves logs for all containers in a given pod.
func FetchPodLogs(client kubernetes.Interface, namespace string, pod corev1.Pod) ([]ContainerLog, error) {
	ctx := context.Background()
	var results []ContainerLog

	for _, container := range pod.Spec.Containers {
		req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name})
		stream, err := req.Stream(ctx)
		if err != nil {
			// If logs can't be fetched for this container, continue with others but record the error.
			results = append(results, ContainerLog{
				PodName:       pod.Name,
				ContainerName: container.Name,
				Logs:          fmt.Sprintf("Error fetching logs: %v", err),
			})
			continue
		}
		defer stream.Close()

		// Read logs into a string
		logData, err := readStream(stream)
		if err != nil {
			results = append(results, ContainerLog{
				PodName:       pod.Name,
				ContainerName: container.Name,
				Logs:          fmt.Sprintf("Error reading logs: %v", err),
			})
			continue
		}

		results = append(results, ContainerLog{
			PodName:       pod.Name,
			ContainerName: container.Name,
			Logs:          logData,
		})
	}

	return results, nil
}

// readStream reads all lines from an io.ReadCloser into a single string.
func readStream(stream io.ReadCloser) (string, error) {
	defer stream.Close()
	scanner := bufio.NewScanner(stream)
	var sb strings.Builder

	for scanner.Scan() {
		sb.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return sb.String(), nil
}
