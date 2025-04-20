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

var (
	verbose bool
	mode    bool
	peek    bool
)

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
		// Check if we're dealing with multiple arguments that might be an unquoted app name
		if len(args) > 1 {
			// Get the flags separately - any argument starting with '-'
			flags := []string{}
			appNameParts := []string{}

			for _, arg := range args {
				if len(arg) > 0 && arg[0] == '-' {
					flags = append(flags, arg)
				} else {
					appNameParts = append(appNameParts, arg)
				}
			}

			// Only suggest combining if we found actual app name parts
			if len(appNameParts) > 1 {
				fmt.Println("⚠️  Detected multiple app name arguments. Did you forget to wrap the app name in quotes?")
				fmt.Printf("    Try: rmapp \"%s\"", joinWithSpaces(appNameParts))

				// Add any flags back to the suggestion
				for _, flag := range flags {
					fmt.Printf(" %s", flag)
				}
				fmt.Println()
				os.Exit(0)
			}
		}

		appName := args[0]
		opts := rmapp.ResolverOptions{
			Verbosity: verbose,
			Mode:      mode,
			Peek:      peek,
		}
		// Create and populate new resolver
		instance, peeked := rmapp.NewResolver(appName, opts)
		if peeked {
			os.Exit(0)
		}
		//time.Sleep(50 * time.Minute)
		instance.Deleter.Delete()

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

// Helper function to join strings with spaces
func joinWithSpaces(parts []string) string {
	result := ""
	for i, part := range parts {
		result += part
		if i < len(parts)-1 {
			result += " "
		}
	}
	return result
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rmapp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&mode, "force", "f", false, "Sets program mode between Trash (Default, Safe, RECOVERABLE) and Force (Full file removal, Unsafe, UNRECOVERABLE)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	rootCmd.Flags().BoolVarP(&peek, "peek", "p", false, "Peek matched files")
}
