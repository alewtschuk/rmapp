package rmapp

/*
Resolver.go holds the logic for the resolver which contains the associated
application bundle data
*/

import (
	"fmt"
	"os/exec"
	"strings"
)

// Resolver holds all the information reagarding the application's info
type Resolver struct {
	AppName       string // full .app name to be deleted
	MdlsReturnStr string // full return string of the mlds command call
	BundleID      string // app's bundle ID
	Finder        Finder // finder to look for files using app info
	verbosity     bool   // is verbose flag set
}

// Creates resolver struct and populates fields
func NewResolver(app string, verbose bool) *Resolver {
	appName := getDotApp(app)
	mdlsReturnStr := getMdlsIdentifier(appName)
	if verbose {
		fmt.Println("Application to delete: ", app)
		fmt.Println("Resolved Bundle ID:", getBundleID(mdlsReturnStr), "\n")
	}
	resolver := &Resolver{
		AppName:       appName,
		MdlsReturnStr: mdlsReturnStr,
		BundleID:      getBundleID(mdlsReturnStr),
		Finder:        NewFinder(app, getBundleID((mdlsReturnStr)), verbose), // uses app name over .app to ensure propper name based searching
		verbosity:     verbose,
	}
	return resolver
}

// Calls mlds to retrieve the bundle identifier
// and converts the bundle identifier to a string
func getMdlsIdentifier(appName string) string {
	if !strings.HasPrefix(appName, "/") {
		appName = fmt.Sprintf("/Applications/%s", appName)
	}

	out, err := exec.Command("mdls", appName, "-name", "kMDItemCFBundleIdentifier").Output()
	if err != nil {
		fmt.Printf("App %s not found.\n", appName)
		return ""
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
