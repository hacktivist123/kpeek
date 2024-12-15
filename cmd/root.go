/*
Copyright © 2024 hacktivist123
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hacktivist123/kpeek/pkg/kube"
	"github.com/spf13/cobra"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var (
	namespace string
	jsonOut   bool
	noLogs    bool
)

type OutputData struct {
	ResourceType string    `json:"resourceType"`
	ResourceName string    `json:"resourceName"`
	Namespace    string    `json:"namespace"`
	Pods         []PodInfo `json:"pods"`
}

type PodInfo struct {
	PodName      string              `json:"podName"`
	Containers   []ContainerInfo     `json:"containers"`
	ContainerLog []kube.ContainerLog `json:"logs,omitempty"`
}

type ContainerInfo struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kpeek [resource/type-name]",
	Short: "kpeek fetches and displays debug information for a Kubernetes resource",
	Long: `kpeek aggregates describe output, logs, and events for a given K8s resource like 
a Deployment or Pod into a single, easy-to-read report.`,
	Args: cobra.ExactArgs(1), // Expect exactly one argument, like "deploy/my-app"
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		parts := strings.SplitN(input, "/", 2)

		if len(parts) != 2 {
			fmt.Println("Invalid input. Expected format: <resource-type>/<resource-name>, e.g deploy/my-app")
			os.Exit(1)
		}

		resourceType := parts[0]
		resourceName := parts[1]

		// Get the Kubernetes client
		client, err := kube.GetClient()
		if err != nil {
			fmt.Printf("Error creating Kubernetes client: %v\n", err)
			os.Exit(1)
		}

		resourceData, err := kube.FetchResource(client, namespace, resourceType, resourceName)
		if err != nil {
			fmt.Printf("Error fetch rosource: %v\n", err)
			os.Exit(1)
		}

		var pods []corev1.Pod
		switch resourceData.Kind {
		case "Deployment":
			deploy := resourceData.Obj.(*appsv1.Deployment)
			pods, err = kube.GetDeploymentPods(client, namespace, deploy)
			if err != nil {
				fmt.Printf("Error fetching deployment pods: %v\n", err)
				os.Exit(1)
			}
		case "Pod":
			pod := resourceData.Obj.(*corev1.Pod)
			pods = []corev1.Pod{*pod}
		default:
			fmt.Printf("Unsupported resource kind: %s\n", resourceData.Kind)
			os.Exit(1)
		}

		var output OutputData
		output.ResourceType = resourceData.Kind
		output.ResourceName = resourceName
		output.Namespace = namespace

		for _, p := range pods {
			podInfo := PodInfo{
				PodName: p.Name,
			}

			// Basic Container Info
			for _, c := range p.Spec.Containers {
				podInfo.Containers = append(podInfo.Containers, ContainerInfo{
					Name:  c.Name,
					Image: c.Image,
				})
			}

			// Fetch Logs if not --no-logs
			if !noLogs {
				logs, err := kube.FetchPodLogs(client, namespace, p)
				if err != nil {
					fmt.Printf("Error fetching logs for pod %s: %v\n", p.Name, err)
					os.Exit(1)
				}
				podInfo.ContainerLog = logs
			}

			output.Pods = append(output.Pods, podInfo)
		}

		// Print output
		if jsonOut {
			data, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Resource: %s/%s (Namespace: %s)\n", resourceType, resourceName, namespace)
			for _, p := range output.Pods {
				fmt.Printf("Pod: %s\n", p.PodName)
				for _, c := range p.Containers {
					fmt.Printf("  Container: %s (Image: %s)\n", c.Name, c.Image)
				}
				if !noLogs {
					for _, log := range p.ContainerLog {
						fmt.Printf("  Logs from %s/%s:\n%s\n", log.PodName, log.ContainerName, log.Logs)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace of the resource")
	rootCmd.Flags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
	rootCmd.Flags().BoolVar(&noLogs, "no-logs", false, "Skip retrieving container logs")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
