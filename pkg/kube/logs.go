package kube

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
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
func FetchPodLogs(client kubernetes.Interface, namespace string, pod corev1.Pod, tailLines int) ([]ContainerLog, error) {
	ctx := context.Background()
	var results []ContainerLog

	for _, container := range pod.Spec.Containers {
		podLogOptions := &corev1.PodLogOptions{
			Container: container.Name,
		}
		if tailLines > 0 {
			t := int64(tailLines)
			podLogOptions.TailLines = &t
		}

		req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, podLogOptions)
		stream, err := req.Stream(ctx)
		if err != nil {
			results = append(results, ContainerLog{
				PodName:       pod.Name,
				ContainerName: container.Name,
				Logs:          fmt.Sprintf("Error fetching logs: %v", err),
			})
			continue
		}

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

// filter logs provides advanced filtering for logs based on regex, log tail
func FilterLogs(logString string, logTail int, logRegex string) string {
	lines := strings.Split(logString, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if logTail > 0 {
		if logTail > len(lines) {
			logTail = len(lines)
		}
		lines = lines[len(lines)-logTail:]
	}

	if len(logRegex) > 0 {
		filteredLines := []string{}
		for _, line := range lines {
			regexMatched, err := regexp.MatchString(logRegex, line)
			if err != nil {
				fmt.Println(err)
			}
			if regexMatched {
				filteredLines = append(filteredLines, line)
			}
		}
		lines = filteredLines
	}
	return strings.Join(lines, "\n")
}
