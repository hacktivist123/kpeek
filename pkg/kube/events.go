package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// FilterEvents filters events for a given resource kind/name and optional pod names.
func FilterEvents(events []corev1.Event, resourceKind, resourceName string, podNames []string) ([]corev1.Event, []corev1.Event) {
	var resourceEvents, podEvents []corev1.Event
	podNameSet := make(map[string]struct{})

	for _, p := range podNames {
		podNameSet[p] = struct{}{}
	}

	for _, e := range events {
		obj := e.InvolvedObject
		if obj.Kind == resourceKind && obj.Name == resourceName {
			resourceEvents = append(resourceEvents, e)
		} else if obj.Kind == "Pod" {
			// Check if this event belongs to one of our pods
			if _, found := podNameSet[obj.Name]; found {
				podEvents = append(podEvents, e)
			}
		}
	}

	return resourceEvents, podEvents
}

// ListAllEventsInNamespace fetches all events in the given namespace.
func ListAllEventsInNamespace(client kubernetes.Interface, namespace string) ([]corev1.Event, error) {
	ctx := context.Background()
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events in namespace %s: %w", namespace, err)
	}
	return events.Items, nil
}


