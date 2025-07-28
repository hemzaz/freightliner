// Package cmd provides the command-line interface commands for freightliner.
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// newVersionCmd creates a new version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Displays the version and build information for this installation of Freightliner`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Freightliner v0.1.0")
			fmt.Println("Go Version:", runtime.Version())
			fmt.Println("OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH)
		},
	}
}
