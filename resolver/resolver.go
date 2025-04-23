package resolver

/*
Resolver.go holds the logic for the resolver which contains the associated
application bundle data
*/

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alewtschuk/pfmt"
	"github.com/alewtschuk/rmapp/deleter"
	"github.com/alewtschuk/rmapp/finder"
	"github.com/alewtschuk/rmapp/options"
)

// Resolver holds all the information reagarding the application's info
type Resolver struct {
	AppName       string          // full .app name to be deleted
	MdlsReturnStr string          // full return string of the mlds command call
	BundleID      string          // app's bundle ID
	Finder        finder.Finder   // finder to look for files using app info
	Options       options.Options // resolver options
	Deleter       deleter.Deleter // deleter struct for handling file removal
	Peeked        bool            // resolved in peek mode
}

// Creates resolver struct and populates fields
func NewResolver(app string, opts options.Options) (*Resolver, bool) {
	appName := getDotApp(app)
	mdlsReturnStr := getMdlsIdentifier(appName)
	if opts.Verbosity {
		fmt.Println("Application to delete: ", pfmt.ApplyColor(app, 2))
		fmt.Print("Resolved Bundle ID: ", pfmt.ApplyColor(getBundleID(mdlsReturnStr), 2), "\n\n")
	}
	finder, peeked := finder.NewFinder(app, getBundleID((mdlsReturnStr)), opts) // uses app name over .app to ensure propper name based searching
	resolver := &Resolver{
		AppName:       appName,
		MdlsReturnStr: mdlsReturnStr,
		BundleID:      getBundleID(mdlsReturnStr),
		Finder:        finder,
		Options:       opts,
		Deleter:       deleter.NewDeleter(finder.MatchedPaths, opts),
		Peeked:        peeked,
	}
	return resolver, peeked
}

// Calls mlds to retrieve the bundle identifier
// and converts the bundle identifier to a string
func getMdlsIdentifier(appName string) string {
	if !strings.HasPrefix(appName, "/") {
		appName = fmt.Sprintf("/Applications/%s", appName)
	}

	out, err := exec.Command("mdls", appName, "-name", "kMDItemCFBundleIdentifier").Output()
	if err != nil {
		appName = strings.TrimSuffix(strings.TrimPrefix(appName, "/Applications/"), ".app")
		fmt.Printf("App %s not found.\n", pfmt.ApplyColor(appName, 2))
		os.Exit(0)
	}
	// Set full mlds output to string
	mdlsReturnStr := string(out)

	return mdlsReturnStr
}

// Takes mlds returned kMDItemCFBundleIdentifier
// string and extracts the bundle id
func getBundleID(mdlsReturnStr string) string {
	bundleID := extractQuotedSubstring(mdlsReturnStr)

	return bundleID
}

// Extracts substring between " delimiter
//
// For use to trim kMDItemCFBundleIdentifier string
func extractQuotedSubstring(str string) string {
	strs := strings.Split(str, "\"")
	if len(strs) >= 2 {
		return strs[1]
	}

	return ""
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
