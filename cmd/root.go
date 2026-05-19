package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "db-manager",
	Short: "Test the db connection",
	Long:  "Cli tool to test the database connection and running tests",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(incrementCmd)
	rootCmd.AddCommand(decrementCmd)
}
