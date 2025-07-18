package finder

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Checks if the file/directory name contains the appName or bundleID
func (f Finder) isMatch(filename string, ctx ScanContext) bool {
	filename = strings.ToLower(filename)
	appName := strings.ToLower(ctx.AppName)
	bundleID := strings.ToLower(ctx.BundleID)

	// appNameTokenzied := tokenize(appName)
	// ttoken := []string{"microsoft", "excel"}

	// fmt.Println(slices.Equal(ttoken, appNameTokenzied))

	// Match full bundleID anywhere in the filename
	if strings.Contains(filename, bundleID) {
		return true
	}

	// Handle numeric suffix variations in bundle ID
	// For example: com.microsoft.teams2 should match com.microsoft.teams (detected edge case)
	bundleIDBase := strings.TrimRightFunc(bundleID, unicode.IsDigit)
	if bundleIDBase != bundleID && strings.Contains(filename, bundleIDBase) {
		return true
	}

	// Only match exact or prefix match
	if strings.HasSuffix(filename, ".app") {
		base := strings.TrimSuffix(filename, ".app")
		if base == appName || strings.HasPrefix(base, appName) {
			return true
		}
	}

	// Otherwise fallback to token check
	tokenizedFilename := tokenize(filename)
	if ctx.RootPath == f.System.SystemReceipts {
		fmt.Println("Name: " + filename)
		fmt.Print("Tokenized Name: ")
		fmt.Println(tokenizedFilename)
		fmt.Println()
		if containsName(appName, tokenizedFilename) {
			return true
		}
	}

	for _, token := range tokenizedFilename {
		if token == appName {
			return true
		}
	}
	return false
}

// Extract domain hint from bundleID (e.g. "com.theapp.App" to "theapp")
func GetDomainHint(bundleID string) string {
	parts := strings.Split(bundleID, ".")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// Tokenize the input based on specific delimiters
//
// Mitigates incorrect matches
func tokenize(name string) []string {
	// Used to tokenize Application name
	if strings.Contains(name, " ") {
		return strings.FieldsFunc(name, func(r rune) bool {
			return r == ' '
		})
	} else { // Tokenizes the file names based on special chars
		return strings.FieldsFunc(name, func(r rune) bool {
			return r == '.' || r == '-' || r == '_' || r == ' ' || r == '/'
		})
	}
}

func containsName(appName string, tokenizedFilename []string) bool {
	appNameTokenzied := tokenize(appName) //tokenize the app name to get individual substrings
	appNameLen := len(appNameTokenzied)   //get length of the tokenized app name array
	leftPointer := 0                      //initialize left to the start of the filename array
	rightPointer := appNameLen - 1        //initialize right to the size of the app name array, length of app name will be window size

	//Edge case, the window can't be created if the filename is shorter than the app name
	if len(tokenizedFilename) < appNameLen {
		return false
	}

	//For the range of the tokens in the filename array slide the window
	for i := 0; i < len(tokenizedFilename); i++ {

		if rightPointer >= len(tokenizedFilename) {
			leftPointer = 0
			rightPointer = appNameLen - 1
			return false
		}
		fmt.Println("Left Pointer index is " + strconv.Itoa(leftPointer) + " and value is " + tokenizedFilename[leftPointer])
		fmt.Println("Right Pointer value is " + strconv.Itoa(rightPointer) + " and value is " + tokenizedFilename[rightPointer])
		if (tokenizedFilename[leftPointer] == appNameTokenzied[0]) && (tokenizedFilename[rightPointer] == appNameTokenzied[appNameLen-1]) {
			return true
		}

		leftPointer++
		rightPointer++
		fmt.Println("Left Pointer is now: " + strconv.Itoa(leftPointer))
		fmt.Println("Right Pointer is now: " + strconv.Itoa(rightPointer))
	}

	// fmt.Print("App Name: ")
	// fmt.Print(appNameTokenzied)
	// fmt.Println("\tApp Name Length: " + strconv.Itoa(appNameLen))
	return false
}
