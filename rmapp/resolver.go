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
	AppName       string // app to be deleted
	MldsReturnStr string // full return string of the mlds command call
	BundleID      string // app's bundle ID
	//Finder        Finder // finder to look for files using app info
}

// Creates resolver struct and populates fields
func NewResolver(app string) *Resolver {
	appName := getDotApp(app)
	mldsReturnStr := getMldsIdentifier(appName)
	resolver := &Resolver{
		AppName:       appName,
		MldsReturnStr: mldsReturnStr,
		BundleID:      getBundleID(mldsReturnStr),
	}
	return resolver
}

// Calls mlds to retrieve the bundle identifier
// and converts the bundle identifier to a string
func getMldsIdentifier(appName string) string {
	out, err := exec.Command("mdls", fmt.Sprintf("/Applications/%s", appName), "-name", "kMDItemCFBundleIdentifier").Output()
	if err != nil {
		fmt.Printf("Error running mdls command: %v\n", err)
		return ""
	}
	// Set full mlds output to string
	mldsReturnStr := string(out)

	return mldsReturnStr
}

// Takes mlds returned kMDItemCFBundleIdentifier
// string and extracts the bundle id
func getBundleID(mldsReturnStr string) string {
	fmt.Println(mldsReturnStr)
	bundleID := extractQuotedSubstring(mldsReturnStr)

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
