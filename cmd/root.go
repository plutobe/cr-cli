package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cr-cli",
	Short: "AI-powered code review tool",
	Long:  "CR-CLI is a CLI tool for AI-powered code review.",
}

func Execute() error {
	return rootCmd.Execute()
}
