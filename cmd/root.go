package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/hacktivist123/kpeek/pkg/kube"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var (
	namespace     string
	jsonOut       bool
	noLogs        bool
	includeEvents bool
)

// OutputData is your final struct holding everything fetched.
type OutputData struct {
	ResourceType string      `json:"resourceType"`
	ResourceName string      `json:"resourceName"`
	Namespace    string      `json:"namespace"`
	Pods         []PodInfo   `json:"pods"`
	Events       []EventInfo `json:"events,omitempty"`    // Resource-level events
	PodEvents    []EventInfo `json:"podEvents,omitempty"` // Pod-level events
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

type EventInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Reason      string `json:"reason"`
	Message     string `json:"message"`
	InvolvedObj string `json:"involvedObj"`
}

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "kpeek [resource/type-name]",
	Short: "kpeek fetches and displays debug information for a Kubernetes resource",
	Long: `kpeek aggregates describe output, logs, and optionally events for a given K8s 
resource like a Deployment or Pod into a single, colorized report.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		parts := strings.SplitN(input, "/", 2)
		if len(parts) != 2 {
			fmt.Println(color.RedString("Invalid input. Expected format: <resource-type>/<resource-name>, e.g. deploy/my-app"))
			os.Exit(1)
		}

		resourceType := parts[0]
		resourceName := parts[1]

		// Get the Kubernetes client
		client, err := kube.GetClient()
		if err != nil {
			fmt.Println(color.RedString("Error creating Kubernetes client: %v", err))
			os.Exit(1)
		}

		resourceData, err := kube.FetchResource(client, namespace, resourceType, resourceName)
		if err != nil {
			fmt.Println(color.RedString("Error fetching resource: %v", err))
			os.Exit(1)
		}

		// If resource is a Deployment, fetch pods. If it's a Pod, just store that Pod.
		var pods []corev1.Pod
		switch resourceData.Kind {
		case "Deployment":
			deploy := resourceData.Obj.(*appsv1.Deployment)
			pods, err = kube.GetDeploymentPods(client, namespace, deploy)
			if err != nil {
				fmt.Println(color.RedString("Error fetching deployment pods: %v", err))
				os.Exit(1)
			}
		case "Pod":
			p := resourceData.Obj.(*corev1.Pod)
			pods = []corev1.Pod{*p}
		default:
			fmt.Println(color.RedString("Unsupported resource kind: %s", resourceData.Kind))
			os.Exit(1)
		}

		// Build the output data
		var output OutputData
		output.ResourceType = resourceData.Kind
		output.ResourceName = resourceName
		output.Namespace = namespace

		// Collect Pod and container info, plus logs (if not noLogs)
		var podNames []string
		for _, p := range pods {
			podNames = append(podNames, p.Name)
			podInfo := PodInfo{PodName: p.Name}

			for _, c := range p.Spec.Containers {
				podInfo.Containers = append(podInfo.Containers, ContainerInfo{
					Name:  c.Name,
					Image: c.Image,
				})
			}

			if !noLogs {
				logs, err := kube.FetchPodLogs(client, namespace, p)
				if err != nil {
					fmt.Println(color.RedString("Error fetching logs for pod %s: %v", p.Name, err))
					os.Exit(1)
				}
				podInfo.ContainerLog = logs
			}

			output.Pods = append(output.Pods, podInfo)
		}

		// Collect events if requested
		if includeEvents {
			allEvents, err := kube.ListAllEventsInNamespace(client, namespace)
			if err != nil {
				fmt.Println(color.RedString("Error listing events: %v", err))
				os.Exit(1)
			}

			resourceEvents, podEvents := kube.FilterEvents(allEvents, output.ResourceType, output.ResourceName, podNames)

			for _, e := range resourceEvents {
				output.Events = append(output.Events, EventInfo{
					Name:        e.Name,
					Type:        e.Type,
					Reason:      e.Reason,
					Message:     e.Message,
					InvolvedObj: fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
				})
			}

			for _, e := range podEvents {
				output.PodEvents = append(output.PodEvents, EventInfo{
					Name:        e.Name,
					Type:        e.Type,
					Reason:      e.Reason,
					Message:     e.Message,
					InvolvedObj: fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
				})
			}
		}

		// If JSON output requested, print JSON. Otherwise, pretty-print with color.
		if jsonOut {
			data, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(data))
		} else {
			Output(resourceType, resourceName, output)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace of the resource")
	rootCmd.Flags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
	rootCmd.Flags().BoolVar(&noLogs, "no-logs", false, "Skip retrieving container logs")
	rootCmd.Flags().BoolVar(&includeEvents, "include-events", false, "Include events in the output")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(color.RedString("%v", err))
		os.Exit(1)
	}
}

func Output(resourceType, resourceName string, out OutputData) {
	// Styles
	boldCyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	boldWhite := color.New(color.FgWhite, color.Bold).SprintFunc()
	boldBlue := color.New(color.FgBlue, color.Bold).SprintFunc()
	yellowText := color.New(color.FgYellow).SprintFunc()

	// Header
	fmt.Println(boldCyan("===================================================="))
	fmt.Printf("%s: %s/%s\n", boldWhite("Resource"), resourceType, resourceName)
	fmt.Printf("%s: %s\n", boldWhite("Namespace"), out.Namespace)
	fmt.Println(boldCyan("===================================================="))

	// PODS & CONTAINERS TABLE
	fmt.Println()
	fmt.Println(boldBlue("Pods & Containers:"))

	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.AppendHeader(table.Row{"Pod Name", "Container", "Image", "Logs?"})

	for _, pod := range out.Pods {
		// If no containers, just note it
		if len(pod.Containers) == 0 {
			tw.AppendRow(table.Row{pod.PodName, "-", "-", "None"})
			tw.AppendSeparator()
			continue
		}
		first := true
		for _, c := range pod.Containers {
			logStatus := yellowText("Skipped (--no-logs)")
			if len(pod.ContainerLog) > 0 {
				// Check if we have logs for this container
				foundLogs := false
				for _, l := range pod.ContainerLog {
					if l.ContainerName == c.Name {
						foundLogs = true
						break
					}
				}
				if foundLogs {
					logStatus = color.GreenString("Fetched")
				} else {
					logStatus = color.RedString("Not Found")
				}
			}
			if first {
				tw.AppendRow(table.Row{pod.PodName, c.Name, c.Image, logStatus})
				tw.AppendSeparator()
				first = false
			} else {
				tw.AppendRow(table.Row{"", c.Name, c.Image, logStatus})
				tw.AppendSeparator()
			}
		}
	}

	fmt.Println(tw.Render())

	// LOGS SECTION
	if noLogs {
		fmt.Println(color.YellowString("\nLogs not included (use without --no-logs to see logs)."))
	} else {
		for _, pod := range out.Pods {
			for _, l := range pod.ContainerLog {
				if l.Logs == "" {
					continue
				}

				fmt.Println(color.YellowString("----------------------------------------------------"))
				fmt.Printf("%s: %s / %s\n", boldBlue("Logs for"), color.GreenString(l.PodName), color.GreenString(l.ContainerName))

				// Highlight lines containing ERROR or WARN
				lines := strings.Split(l.Logs, "\n")
				for _, line := range lines {
					switch {
					case strings.Contains(line, "ERROR"):
						fmt.Println(color.RedString("  %s", line))
					case strings.Contains(line, "WARN"):
						fmt.Println(color.YellowString("  %s", line))
					default:
						fmt.Println("  " + line)
					}
				}
			}
		}
	}

	// EVENTS SECTION
	if len(out.Events)+len(out.PodEvents) == 0 {
		if !includeEvents {
			fmt.Println(color.YellowString("\nEvents not included (use --include-events to see events)."))
		} else {
			fmt.Println(color.YellowString("\nNo events found for this resource or its pods."))
		}
		return
	}

	fmt.Println(boldCyan("\n----------------------------------------------------"))
	fmt.Println(boldBlue("Events:"))

	// Resource-level events table
	if len(out.Events) > 0 {
		fmt.Println(boldWhite("Resource-Level Events:"))
		resTbl := table.NewWriter()
		resTbl.SetStyle(table.StyleLight)
		resTbl.AppendHeader(table.Row{"Type", "Reason", "Message"})
		for _, e := range out.Events {
			resTbl.AppendRow(table.Row{e.Type, e.Reason, e.Message})
			resTbl.AppendSeparator()
		}
		fmt.Println(resTbl.Render())
	}

	// Pod-level events table
	if len(out.PodEvents) > 0 {
		fmt.Println(boldWhite("Pod-Level Events:"))
		podTbl := table.NewWriter()
		podTbl.SetStyle(table.StyleLight)
		podTbl.AppendHeader(table.Row{"Type", "Reason", "Message", "Pod"})
		for _, e := range out.PodEvents {
			splitName := strings.Split(e.InvolvedObj, "/")
			podName := e.InvolvedObj
			if len(splitName) > 1 {
				podName = splitName[1]
			}
			podTbl.AppendRow(table.Row{e.Type, e.Reason, e.Message, podName})
			podTbl.AppendSeparator()
		}
		fmt.Println(podTbl.Render())
	}
}
