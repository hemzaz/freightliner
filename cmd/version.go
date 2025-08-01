// Package cmd provides the command-line interface commands for freightliner.
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information set at build time via ldflags
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// newVersionCmd creates a new version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Displays the version and build information for this installation of Freightliner`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Freightliner %s\n", version)
			fmt.Printf("Git Commit: %s\n", gitCommit)
			fmt.Printf("Build Time: %s\n", buildTime)
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}

// newHealthCheckCmd creates a new health-check command for containers
func newHealthCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health-check",
		Short: "Perform health check",
		Long:  `Performs a health check suitable for container health checks`,
		Run: func(cmd *cobra.Command, args []string) {
			// Simple health check - in a real implementation this would
			// check database connectivity, external services, etc.
			fmt.Println("OK")
		},
	}
}
