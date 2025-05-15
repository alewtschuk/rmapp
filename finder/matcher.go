package finder

import (
	"strings"
	"unicode"
)

// Checks if the file/directory name contains the appName or bundleID
func isMatch(name, appName, bundleID string) bool {
	name = strings.ToLower(name)
	appName = strings.ToLower(appName)
	bundleID = strings.ToLower(bundleID)

	// Match full bundleID anywhere in the name
	if strings.Contains(name, bundleID) {
		return true
	}

	// Handle numeric suffix variations in bundle ID
	// For example: com.microsoft.teams2 should match com.microsoft.teams (detected edge case)
	bundleIDBase := strings.TrimRightFunc(bundleID, unicode.IsDigit)
	if bundleIDBase != bundleID && strings.Contains(name, bundleIDBase) {
		return true
	}

	// Only match exact or prefix match
	if strings.HasSuffix(name, ".app") {
		base := strings.TrimSuffix(name, ".app")
		if base == appName || strings.HasPrefix(base, appName) {
			return true
		}
	}

	// Otherwise fallback to token check
	for _, token := range tokenize(name) {
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

// Tokenize the file/directory name based on delimiters
//
// Mitigates incorrect matches
func tokenize(name string) []string {
	return strings.FieldsFunc(name, func(r rune) bool {
		return r == '.' || r == '-' || r == '_' || r == ' ' || r == '/'
	})
}
