/*
Copyright © 2025 Alex Lewtschuk
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/alewtschuk/pfmt"
	"github.com/alewtschuk/rmapp/options"
	"github.com/alewtschuk/rmapp/resolver"
	"github.com/spf13/cobra"
)

var version string = ""
var banner string = pfmt.ApplyColor(`

    ___       ___       ___       ___       ___   
   /\  \     /\__\     /\  \     /\  \     /\  \  
  /::\  \   /::L_L_   /::\  \   /::\  \   /::\  \ 
 /::\:\__\ /:/L:\__\ /::\:\__\ /::\:\__\ /::\:\__\
 \;:::/  / \/_/:/  / \/\::/  / \/\::/  / \/\::/  /
  |:\/__/    /:/  /    /:/  /     \/__/     \/__/ 
   \|__|     \/__/     \/__/              `, 33)

var (
	verbose    bool
	force      bool
	peek       bool
	logical    bool
	versionOpt bool
	size       bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rmapp app_name",
	Short: "Removes specified macOS apps and thier associated files",
	Long: banner + `        

rmapp is a macOS app removal tool for command line and power users.
It deletes both standard .app bundles and associated files stored elsewhere
in your system, securely, with file size reporting, and default safe trashing.`,
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
				fmt.Println("[rmapp] ⚠️ Detected multiple app name arguments. Did you forget to wrap the app name in quotes?")
				fmt.Printf("           Try: rmapp \"%s\"", joinWithSpaces(appNameParts))

				// Add any flags back to the suggestion
				for _, flag := range flags {
					fmt.Printf(" %s", flag)
				}
				fmt.Println()
				os.Exit(0)
			}
		}

		appName := args[0]
		var opts options.Options
		// Enables logical file size if peek is used
		switch {

		case peek:
			opts = options.Options{
				Verbosity: verbose,
				Mode:      force,
				Peek:      peek,
				Logical:   logical,
			}

		case size:
			opts = options.Options{
				Verbosity: verbose,
				Mode:      force,
				Peek:      peek,
				Logical:   logical,
				Size:      size,
			}
		default:
			opts = options.Options{
				Verbosity: verbose,
				Mode:      force,
				Peek:      peek,
			}

		}

		if !opts.Verbosity {
			log.SetOutput(io.Discard)
		} else {
			log.SetOutput(os.Stdout)
			log.SetFlags(0)
		}
		// Create and populate new resolver
		instance := resolver.NewResolver(appName, opts)
		if instance.Reported {
			os.Exit(0)
		}
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
	cobra.OnInitialize(getVersion)
	cobra.OnInitialize(checkArgs)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rmapp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&force, "force", "f", false,
		fmt.Sprintf("Sets program force between %s and %s",
			pfmt.ApplyColor("Trash (Default, Safe, RECOVERABLE)", 2),
			pfmt.ApplyColor("Force (Full file removal, Unsafe, UNRECOVERABLE)", 9)),
	)
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	rootCmd.Flags().BoolVarP(&peek, "peek", "p", false, "Peek matched files")
	rootCmd.Flags().BoolVarP(&logical, "logical", "l", false, "Show logical file size")
	rootCmd.Flags().BoolVar(&versionOpt, "version", false, "Show rmapp version")
	rootCmd.Flags().BoolVarP(&size, "size", "s", false, "Show the total size of the application's data")
}

// Prints version
func getVersion() {
	if versionOpt {
		fmt.Println("rmapp version: " + version)
		os.Exit(0)
	}
}

// Checks argument compatibility
func checkArgs() {
	if peek && force {
		pfmt.Printcln("[rmapp] Incompatible args '--force' and '--peek' please run again with one or the other...", 9)
		fmt.Println()
		os.Exit(0)
	}

	if logical && force {
		pfmt.Printcln("[rmapp] Incompatible args '--force' and '--logical'. '--logical' can only be run in peek context. \nPlease run again with '--force' alone or '--logical' and '--peek'...", 9)
		fmt.Println()
		os.Exit(0)
	}

	if !(peek || size) && logical {
		pfmt.Printcln("[rmapp] Incompatible args '--logical' must be used in '--peek' context. Please run again with '--peek' enabled...", 9)
		os.Exit(0)
	}

	if peek && size {
		pfmt.Printcln("[rmapp] Incompatible args '--size' is shown in '--peek'. Please choose one argument and run again...", 9)
		os.Exit(0)
	}
}
