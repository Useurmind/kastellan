package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "Kastellan Agent - A CLI application for managing agent operations",
	Long: `Kastellan Agent is a CLI application that provides commands for managing
agent operations within the Kastellan system.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you can define flags and configuration for the root command
}
