package kube

import (
	"context"
	"fmt"

	// appsv1 "k8s.io/api/apps/v1"
	// corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceData is a simple struct that can hold the retrieved object and its kind.
type ResourceData struct {
	Kind string
	Obj  interface{}
}

func FetchResource(client kubernetes.Interface, namespace, resourceType, resourceName string) (*ResourceData, error) {
	ctx := context.Background()

	switch resourceType {
	case "deploy", "deployment":
		deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment: %w", err)
		}
		return &ResourceData{Kind: "Deployment", Obj: deploy}, nil

	case "pod":
		pod, err := client.CoreV1().Pods(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod: %w", err)
		}
		return &ResourceData{Kind: "Pod", Obj: pod}, nil

	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}
