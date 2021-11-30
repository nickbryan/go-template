package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "gotemplate",
	Short: "A simple tempalte application for Go microservices.",
	Long:  "This is a simple template application/demo for Go microservices.",
}

func Execute() {
	// Our application errors are passed through cobra back to here so we can report and exit accordingly.
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
