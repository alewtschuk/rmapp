package finder

import (
	"strings"
	"unicode"
)

// Checks if the file/directory name contains the appName or bundleID
func (f Finder) isMatch(filename string, ctx ScanContext) bool {
	filename = strings.ToLower(filename)
	bundleID := strings.ToLower(ctx.BundleID)

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

	// Otherwise fallback to token check
	return searchName(ctx, filename)
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
	// Tokenizes the file names based on special chars
	return strings.FieldsFunc(name, func(r rune) bool {
		return r == '.' || r == '-' || r == '_' || r == ' ' || r == '/'
	})
}

// Utilizes KNP search algorithm to find match occurences
// of the app name inside the file name.
func searchName(ctx ScanContext, filename string) bool {

	//Tokenize files and build lps
	filename = strings.TrimRightFunc(filename, unicode.IsDigit) //trim any numeric suffix off filename
	tokenizedFile := tokenize(filename)
	lps := ctx.LpsArray

	//Initalize length and pointer values
	n := len(tokenizedFile)
	m := len(ctx.TokenizedApp)
	i := 0
	j := 0

	//While i < length of tokenizedFile
	for i < n {
		//Match occurs, move pointers forward
		if tokenizedFile[i] == ctx.TokenizedApp[j] {
			i++
			j++
			//Complete match found return true
			if j == m {
				return true
			}
			//Mismatch occurs
		} else {
			//As long as j isnt 0 move j back to previous prefix-suffix match
			if j != 0 {
				j = lps[j-1]
				//When j is 0 increment i forwards as no part of prefix-suffix can be used
			} else {
				i++
			}
		}
	}

	return false
}

// Build the LPS array for the KNP search algorithm
func buildLPS(pattern []string) []int {

	//Set length of the current longest prefix == suffix
	length := 0

	//Create lps array to size of pattern array and set first lps index
	lps := make([]int, len(pattern))
	lps[0] = 0 //will always be zero as no prefix-sufix can exist yet

	//Current index to build in lps
	i := 1
	for i < len(pattern) {
		//Length will be increased as the prefix that is also a suffix expands
		//Lps positions at i will be filled with updated length values
		if pattern[i] == pattern[length] {
			length++
			lps[i] = length
			i++
			//Backtrack to smaller prefix-suffix border
		} else {
			if length != 0 {
				length = lps[length-1]
				//If there is no usable prefix-suffix border
			} else {
				lps[i] = 0
				i++
			}
		}
	}
	return lps
}
