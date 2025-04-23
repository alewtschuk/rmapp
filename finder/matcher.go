package finder

import (
	"fmt"
	"strings"

	"github.com/alewtschuk/pfmt"
	"github.com/alewtschuk/rmapp/options"
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

// Decide if a directory should be skipped based on context
func shouldSkipDir(name string, depth int, ctx ScanContext) bool {
	if depth > ctx.SearchDepth {
		return true
	}
	if ctx.SearchDepth == STANDARD_DEPTH && depth < ctx.SearchDepth {
		if strings.Contains(name, ctx.DomainHint) && !isMatch(name, ctx.AppName, ctx.BundleID) {
			return true
		}
	}
	return false
}

// Helper function to print and send matches
func emitMatch(name, path string, matchesChan chan string, opts options.Options) {
	if opts.Verbosity && !opts.Peek {
		fmt.Printf("Match %s FOUND at: %s\n", pfmt.ApplyColor(name, 2), pfmt.ApplyColor(path, 3))
	}

	matchesChan <- path
}

// Extract domain hint from bundleID (e.g. "company.thebrowser.Browser" → "thebrowser")
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
