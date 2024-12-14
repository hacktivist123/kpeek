package kube

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
)

func GetClient() (kubernetes.Interface, error) {
	// try to use KUBECONFIG if available
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// check default location if not set
		homeDir, err := os.UserHomeDir()
		if err == nil {
			kubeconfig = filepath.Join(homeDir, ".kube", "config")
		}
	}

	config, err := buildConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create clientset: %w, err")
	}
	return clientset, nil
}

func buildConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		if _, err := os.Stat(kubeconfigPath); err == nil {
			// if KUBECONFIG file exists, use it
			return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		}
	}
	// if no kubeconfig file, try in-cluster config
	return rest.InClusterConfig()
} 
