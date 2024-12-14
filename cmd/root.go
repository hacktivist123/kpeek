/*
Copyright © 2024 hacktivist123
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kpeek [resource/type-name]",
	Short: "kpeek fetches and displays debug information for a Kubernetes resource",
	Long:  `kpeek aggregates describe output, logs, and events for a given K8s resource like a Deployment or Pod.`,
	Args:  cobra.ExactArgs(1), // Expect exactly one argument, like "deploy/my-app"
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		parts := strings.SplitN(input, "/", 2)

		if len(parts) != 2 {
			fmt.Println("Invalid input. Expected format: <resource-type>/<resource-name>, e.g deploy/my-app")
			os.Exit(1)
		}

		resourceType := parts[0]
		resourceName := parts[1]

		fmt.Printf("Welcome to kpeek! Resource Type: %s, Resource Name: %s\n", resourceType, resourceName)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// func init() {
// 	// Here you will define your flags and configuration settings.
// 	// Cobra supports persistent flags, which, if defined here,
// 	// will be global for your application.

// 	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kpeek.yaml)")

// 	// Cobra also supports local flags, which will only run
// 	// when this action is called directly.
// 	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
