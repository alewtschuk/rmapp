package resolver

/*
Resolver.go holds the logic for the resolver which contains the associated
application bundle data
*/

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alewtschuk/pfmt"
	"github.com/alewtschuk/rmapp/darwin"
	"github.com/alewtschuk/rmapp/deleter"
	"github.com/alewtschuk/rmapp/finder"
	"github.com/alewtschuk/rmapp/options"
)

// Resolver holds all the information reagarding the application's info
type Resolver struct {
	AppName    string          // full .app name to be deleted
	BundleID   string          // app's bundle ID
	Finder     finder.Finder   // finder to look for files using app info
	Options    options.Options // resolver options
	Deleter    deleter.Deleter // deleter struct for handling file removal
	Reported   bool            // resolved if peek or size is true
	BundleOnly bool            // holds if only the bundle will be removed
}

// Creates resolver struct and populates fields
func NewResolver(app string, opts options.Options) *Resolver {
	appName := getDotApp(app)
	bundleID := getBundleID(getAppPath(appName))
	if opts.Verbosity {
		log.Println("\nApplication to delete: ", pfmt.ApplyColor(app, 2))
		log.Print("Resolved Bundle ID: ", pfmt.ApplyColor(bundleID, 2), "\n\n")
	}

	// Sets if a report and exit is needed
	var isReported bool = false
	if opts.Peek || opts.Size {
		isReported = true
	}

	finder := finder.NewFinder(app, bundleID, opts) // uses app name over .app to ensure propper name based searching

	resolver := &Resolver{
		AppName:    appName,
		BundleID:   bundleID,
		Finder:     finder,
		Options:    opts,
		Deleter:    deleter.NewDeleter(finder.MatchedPaths, finder.MatchedSizes, opts),
		Reported:   isReported,
		BundleOnly: opts.BundleOnly,
	}

	return resolver
}

var (
	errAppNotFound = errors.New("app not found")
	errNoBundleID  = errors.New("bundleid value is empty")
)

// Resolves the full path to the .app bundle, relative names are
// resolved against /Applications
func getAppPath(appName string) string {
	if !strings.HasPrefix(appName, "/") {
		appName = fmt.Sprintf("/Applications/%s", appName)
	}

	return appName
}

// Retrieves the bundle identifier for the app at appPath, exiting if
// the app does not exist or has no bundle identifier
func getBundleID(appPath string) string {
	bundleID, err := lookupBundleID(appPath)
	switch err {
	case errAppNotFound:
		appName := strings.TrimSuffix(strings.TrimPrefix(appPath, "/Applications/"), ".app")
		fmt.Printf("[rmapp] App %s not found.\n", pfmt.ApplyColor(appName, 2))
		os.Exit(1)
	case errNoBundleID:
		fmt.Println(pfmt.ApplyColor("[rmapp] Error: BundleId is empty", 9))
		os.Exit(1)
	}

	return bundleID
}

// Reads the bundle identifier from the app bundle's Info.plist
//
// Reads the bundle directly rather than querying Spotlight via mdls, which
// spawns a process and fails if Spotlight has not indexed the app
func lookupBundleID(appPath string) (string, error) {
	if _, err := os.Stat(appPath); err != nil {
		return "", errAppNotFound
	}

	bundleID, ok := darwin.GetBundleIdentifier(appPath)
	if !ok || bundleID == "" {
		return "", errNoBundleID
	}

	return bundleID, nil
}

// Checks if the input app name contains ".app"
// if not append to app name
func getDotApp(name string) string {
	if strings.Contains(name, ".app") {
		return name
	} else {
		name += ".app"
		return name
	}
}
