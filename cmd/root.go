/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/pratyushtiwary/toki/toki"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "toki",
	Short: "Auth orchestrator",
	Long:  "Toki is an auth orchestrator that makes it easy to run long running programs on short lived auth tokens",
	Args:  cobra.ExactArgs(1),
	Run:   toki.Run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("project-config", "p", "", "Specifies location of project config")
	rootCmd.Flags().BoolP("dev", "d", false, "Run toki in development mode")
	rootCmd.Flags().BoolP("verbose", "v", false, "[WIP] Prints all the information to stdout, including stdout/stderr of all the process executed")
}
