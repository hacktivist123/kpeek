package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

var version = "v0.7.0"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update kpeek to the latest version",
	Long: `Check GitHub for a newer release of kpeek and update the binary automatically.

This command fetches the latest release from GitHub (repository "hacktivist123/kpeek") and
replaces the current binary if a newer version is available. It will retry a few times if
the network is unreachable.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")

		rawVersion := strings.TrimPrefix(version, "v")
		vLocal, err := semver.ParseTolerant(rawVersion)
		if err != nil {
			fmt.Printf("Could not parse local version (%s): %v\n", version, err)
			fmt.Println("Falling back to version 0.0.0 for update checks.")
			vLocal = semver.MustParse("0.0.0")
		}

		const maxRetries = 3
		var success bool
		var updateErr error

		for attempt := 1; attempt <= maxRetries; attempt++ {
			updateErr = attemptUpdate(vLocal)
			if updateErr == nil {
				success = true
				break
			}

			if isOfflineOrNetworkError(updateErr) {
				fmt.Printf("Network error (attempt %d/%d): %v\n", attempt, maxRetries, updateErr)
				if attempt < maxRetries {
					fmt.Println("Retrying in 2s...")
					time.Sleep(2 * time.Second)
				}
			} else {
				break
			}
		}

		if !success {
			fmt.Printf("Update failed after %d attempt(s): %v\n", maxRetries, updateErr)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func attemptUpdate(current semver.Version) error {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{})
	if err != nil {
		return fmt.Errorf("error creating updater: %w", err)
	}

	res, err := updater.UpdateSelf(current, "hacktivist123/kpeek")
	if err != nil {
		return err
	}

	if res.Version.Equals(current) {
		fmt.Println("You are already using the latest version!")
	} else {
		fmt.Printf("Successfully updated to version %s\n", res.Version)
	}
	return nil
}

// isOfflineOrNetworkError checks if the error is likely a network issue, prompting a retry.
func isOfflineOrNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	lowerMsg := strings.ToLower(err.Error())
	if strings.Contains(lowerMsg, "connection refused") ||
		strings.Contains(lowerMsg, "dial tcp") {
		return true
	}

	return false
}
