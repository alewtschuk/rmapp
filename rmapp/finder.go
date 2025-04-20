package rmapp

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alewtschuk/pfmt"
)

// Declare constants
const (
	STANDARD_DEPTH    int = 1
	PREFERENCES_DEPTH int = 2
)

// ScanContext encapsulates all info needed during directory walking
type ScanContext struct {
	AppName     string
	BundleID    string
	DomainHint  string
	SearchDepth int
}

// Whole Finder struct that holds everything related to finder
type Finder struct {
	OSMain       OSMainPaths
	System       SystemPaths
	UserPaths    UserPaths
	MatchedPaths []string
	verbosity    bool
}

// The default os directories where the .app file should exist
type OSMainPaths struct {
	RootApplicationsPath string // default os applications path
	UserApplicationsPath string // default user applications path
}

// Directories where the system wide paths is stored
type SystemPaths struct {
	GlobalSupportFilesPath string
	//GlobalPreferencesFilesPath string //NOTE: Dir doesn't seem to hold user installed app data. Disabling for now
}

// Directories holding user specific paths
type UserPaths struct {
	AppSupportFilesPath string
	PreferencesPath     string
	CachesPath          string
	ContainersPath      string
	SavedStatePath      string
	HTTPStorages        string
}

// Creates and loads a new Finder with all needed fields
func NewFinder(appName string, bundleID string, opts ResolverOptions) (Finder, bool) {
	finder := Finder{
		OSMain: OSMainPaths{
			RootApplicationsPath: "/Applications",
			UserApplicationsPath: fmt.Sprintf("/Users/%s/Applications", os.Getenv("USER")),
		},
		System: SystemPaths{
			GlobalSupportFilesPath: "/Library/Application Support",
		},
		UserPaths: UserPaths{
			AppSupportFilesPath: fmt.Sprintf("/Users/%s/Library/Application Support", os.Getenv("USER")),
			PreferencesPath:     fmt.Sprintf("/Users/%s/Library/Preferences", os.Getenv("USER")),
			CachesPath:          fmt.Sprintf("/Users/%s/Library/Caches", os.Getenv("USER")),
			ContainersPath:      fmt.Sprintf("/Users/%s/Library/Containers", os.Getenv("USER")),
			SavedStatePath:      fmt.Sprintf("/Users/%s/Library/Saved Application State", os.Getenv("USER")),
			HTTPStorages:        fmt.Sprintf("/Users/%s/Library/HTTPStorages", os.Getenv("USER")),
		},
		verbosity: opts.Verbosity,
	}
	matches, peeked, err := finder.FindMatches(appName, bundleID, opts)
	if err != nil {
		fmt.Println("NewFinder Error: ", err)
	}

	finder.MatchedPaths = matches
	return finder, peeked
}

// Returns a string of all available paths to search
func (f Finder) AllSearchPaths() []string {
	return []string{
		f.OSMain.RootApplicationsPath,
		f.OSMain.UserApplicationsPath,
		f.System.GlobalSupportFilesPath,
		f.UserPaths.AppSupportFilesPath,
		f.UserPaths.PreferencesPath,
		f.UserPaths.CachesPath,
		f.UserPaths.ContainersPath,
		f.UserPaths.SavedStatePath,
		f.UserPaths.HTTPStorages,
	}
}

// Walks the filepath for each path available and checks if each path contains a match
// to the bundleID or the appname.
//
// Internal WalkDir function passes matches to a channel which will be read from to
// build a string slice of matched paths that will be flagged for deletion
func (f *Finder) FindMatches(appName, bundleID string, opts ResolverOptions) ([]string, bool, error) {
	var err error
	var matches []string
	matchesChan := make(chan string)
	wg := sync.WaitGroup{}

	domainHint := getDomainHint(bundleID)

	for _, rootPath := range f.AllSearchPaths() {
		wg.Add(1)

		go func(rootPath string) {
			defer wg.Done()
			searchDepth := STANDARD_DEPTH
			if rootPath == f.UserPaths.PreferencesPath {
				searchDepth = PREFERENCES_DEPTH
			}

			ctx := ScanContext{
				AppName:     appName,
				BundleID:    bundleID,
				DomainHint:  domainHint,
				SearchDepth: searchDepth,
			}

			err := filepath.WalkDir(rootPath, func(subPath string, d fs.DirEntry, err error) error {
				if err != nil {
					if f.verbosity {
						fmt.Fprintf(os.Stderr, "Skipping %s due to WalkDir error: %v\n", pfmt.ApplyColor(subPath, 3), err)
					}
					return nil // or return err if you want to stop
				}
				return f.handleScan(d, subPath, rootPath, matchesChan, ctx, opts)
			})

			if err != nil {
				fmt.Println(" Error on path:", rootPath, err)
			}
		}(rootPath)
	}

	// Go routine to close the channel
	go func() {
		wg.Wait()
		close(matchesChan)
	}()

	// Append match to matches for all matches in channel
	for match := range matchesChan {
		matches = append(matches, match)
	}

	for _, match := range matches {
		if opts.Peek {
			fmt.Printf("Match %s FOUND at: %s\n", pfmt.ApplyColor(appName, 2), pfmt.ApplyColor(match, 3))
		}
	}

	peeked := false
	// If --peek is enabled send back signal to exit to calling function
	if opts.Peek {
		peeked = true
	}

	return matches, peeked, err
}

// Extract domain hint from bundleID (e.g. "company.thebrowser.Browser" → "thebrowser")
func getDomainHint(bundleID string) string {
	parts := strings.Split(bundleID, ".")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

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

// Tokenize the file/directory name based on delimiters
//
// Mitigates incorrect matches
func tokenize(name string) []string {
	return strings.FieldsFunc(name, func(r rune) bool {
		return r == '.' || r == '-' || r == '_' || r == ' ' || r == '/'
	})
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
func emitMatch(name, path string, matchesChan chan string, opts ResolverOptions) {
	if opts.Verbosity && !opts.Peek {
		fmt.Printf("Match %s FOUND at: %s\n", pfmt.ApplyColor(name, 2), pfmt.ApplyColor(path, 3))
	}

	matchesChan <- path
}

// Handles the files/directories if there is a match
//
// Sends all matches to a channel for shared goroutine communication
func (f *Finder) handleScan(d fs.DirEntry, subPath, rootPath string, matchesChan chan string, ctx ScanContext, opts ResolverOptions) error {
	name := d.Name()

	if d.Type().IsRegular() && isMatch(name, ctx.AppName, ctx.BundleID) {
		emitMatch(name, subPath, matchesChan, opts)
		if !opts.Peek {
			fmt.Println()
		}

		return nil
	}

	if d.Type()&os.ModeSymlink != 0 {
		return fs.SkipDir
	}

	if d.Type().IsDir() {
		relPath, err := filepath.Rel(rootPath, subPath)
		if err != nil {
			return nil
		}
		pathSeg := strings.Split(relPath, string(os.PathSeparator))
		depth := len(pathSeg)

		if isMatch(name, ctx.AppName, ctx.BundleID) {
			emitMatch(name, subPath, matchesChan, opts)
			if !opts.Peek {
				fmt.Println()
			}
			return nil
		}

		if shouldSkipDir(name, depth, ctx) {
			return fs.SkipDir
		}
	}
	return nil
}
