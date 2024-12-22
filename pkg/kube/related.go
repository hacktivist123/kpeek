package kube

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetDeploymentPods returns the pods managed by a given deployment.
func GetDeploymentPods(client kubernetes.Interface, namespace string, deploy *appsv1.Deployment) ([]corev1.Pod, error) {
	ctx := context.Background()
	selector := deploy.Spec.Selector.MatchLabels
	if selector == nil {
		return nil, fmt.Errorf("deployment has no selector")
	}

	// Convert the map into a label selector string
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: selector})

	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for deployment: %w", err)
	}

	return pods.Items, nil
}
