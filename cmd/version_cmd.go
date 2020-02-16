package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// BuildCommit lastest build commit (set by Makefile)
	BuildCommit = ""
	// BuildTag if the `BuildCommit` matches a tag
	BuildTag = ""
	// BuildTime set by build script (set by Makefile)
	BuildTime = ""
)

// NewVersionCmd returns the root command
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and build info",
		Run: func(cmd *cobra.Command, args []string) {
			if BuildTag != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "version:    %s\n", BuildTag)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "commit:     %s\n", BuildCommit)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "build time: %s\n", BuildTime)
		},
	}
}
