/*
Copyright © 2025 Alex Lewtschuk
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/alewtschuk/rmapp/rmapp"
	"github.com/spf13/cobra"
)

var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rmapp",
	Short: "Removes a macOS app and all associated files",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1), // Set minimum required args to 1 (app to remove)
	Run: func(cmd *cobra.Command, args []string) {
		appName := os.Args[1]
		// Create and populate new resolver
		resolver := rmapp.NewResolver(appName)
		fmt.Println(resolver)
		if verbose {
			fmt.Println("Application to delete", appName)
			fmt.Println("Resolved Bundle ID:", resolver.BundleID)
		}

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

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rmapp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
}
