package rmapp

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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
	MatchesChan chan string
}

// PeekContext encapsulates all info needed durring handling '--peek' related activities
type PeekContext struct {
	matches   []string
	opts      ResolverOptions
	totalSize int64
	numFiles  int
	appName   string
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
	SystemSupportFilesPath      string
	SystemCrashReports          string
	SystemCaches                string
	SystemExtensions            string
	SystemInternetPlugIns       string
	SystemLaunchAgents          string
	SystemLaunchDaemons         string
	SystemLogs                  string
	SystemPrivilegedHelperTools string
	SystemRecipts               string
	SystemBin                   string
	SystemOpt                   string
	SystemSbin                  string
	SystemShare                 string
	SystemVar                   string
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
	GroupedContainers   string
	InternetPlugIns     string
	LaunchAgents        string
	Logs                string
	WebKit              string
	ApplicationScripts  string
}

// Creates and loads a new Finder with all needed fields
func NewFinder(appName string, bundleID string, opts ResolverOptions) (Finder, bool) {
	finder := Finder{
		OSMain: OSMainPaths{
			RootApplicationsPath: "/Applications",
			UserApplicationsPath: fmt.Sprintf("/Users/%s/Applications", os.Getenv("USER")),
		},
		System: SystemPaths{
			SystemSupportFilesPath:      "/Library/Application Support",
			SystemCrashReports:          "/Library/Application Support/CrashReporter",
			SystemCaches:                "/Library/Caches",
			SystemExtensions:            "/Library/Extensions",
			SystemInternetPlugIns:       "/Library/Internet Plug-Ins",
			SystemLaunchAgents:          "/Library/LaunchAgents",
			SystemLaunchDaemons:         "/Library/LaunchDaemons",
			SystemLogs:                  "/Library/Logs",
			SystemPrivilegedHelperTools: "/Library/PrivilegedHelperTools",
			SystemRecipts:               "/private/var/db/recipts",
			SystemBin:                   "/usr/local/bin",
			SystemOpt:                   "/usr/local/opt",
			SystemSbin:                  "/usr/local/sbin",
			SystemShare:                 "/usr/local/share",
			SystemVar:                   "/usr/local/var",
		},
		UserPaths: UserPaths{
			AppSupportFilesPath: fmt.Sprintf("/Users/%s/Library/Application Support", os.Getenv("USER")),
			PreferencesPath:     fmt.Sprintf("/Users/%s/Library/Preferences", os.Getenv("USER")),
			CachesPath:          fmt.Sprintf("/Users/%s/Library/Caches", os.Getenv("USER")),
			ContainersPath:      fmt.Sprintf("/Users/%s/Library/Containers", os.Getenv("USER")),
			SavedStatePath:      fmt.Sprintf("/Users/%s/Library/Saved Application State", os.Getenv("USER")),
			HTTPStorages:        fmt.Sprintf("/Users/%s/Library/HTTPStorages", os.Getenv("USER")),
			GroupedContainers:   fmt.Sprintf("/Users/%s/Library/Grouped Containers", os.Getenv("USER")),
			InternetPlugIns:     fmt.Sprintf("/Users/%s/Library/Internet Plug-Ins", os.Getenv("USER")),
			LaunchAgents:        fmt.Sprintf("/Users/%s/Library/LaunchAgents", os.Getenv("USER")),
			Logs:                fmt.Sprintf("/Users/%s/Library/Logs", os.Getenv("USER")),
			WebKit:              fmt.Sprintf("/Users/%s/Library/WebKit", os.Getenv("USER")),
			ApplicationScripts:  fmt.Sprintf("/Users/%s/Library/Application Scripts", os.Getenv("USER")),
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
		f.System.SystemSupportFilesPath,
		f.System.SystemCaches,
		f.System.SystemCrashReports,
		f.System.SystemSupportFilesPath,
		f.UserPaths.AppSupportFilesPath,
		f.UserPaths.PreferencesPath,
		f.UserPaths.CachesPath,
		f.UserPaths.ContainersPath,
		f.UserPaths.SavedStatePath,
		f.UserPaths.HTTPStorages,
		f.UserPaths.GroupedContainers,
		f.UserPaths.InternetPlugIns,
		f.UserPaths.LaunchAgents,
		f.UserPaths.Logs,
		f.UserPaths.WebKit,
		f.UserPaths.ApplicationScripts,
	}
}

// Walks the filepath for each path available and checks if each path contains a match
// to the bundleID or the appname.
//
// Internal WalkDir function passes matches to a channel which will be read from to
// build a string slice of matched paths that will be flagged for deletion
func (f *Finder) FindMatches(appName, bundleID string, opts ResolverOptions) ([]string, bool, error) {
	var (
		err       error
		numFiles  int
		totalSize int64
		matches   []string
	)
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

			// Create context struct for passing context to other functions
			ctx := ScanContext{
				AppName:     appName,
				BundleID:    bundleID,
				DomainHint:  domainHint,
				SearchDepth: searchDepth,
				MatchesChan: matchesChan,
			}

			// Check if root Applications directories hold the .app
			if rootPath == f.OSMain.RootApplicationsPath || rootPath == f.OSMain.UserApplicationsPath {
				f.findMatchesApp(rootPath, ctx)
				return
			}
			// For all other scanned directories we need to walk
			f.findMatchesWalk(rootPath, ctx, opts)
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

	return matches, handlePeek(matches,
			opts,
			totalSize,
			numFiles,
			appName),
		err
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
func (f *Finder) handleScan(d fs.DirEntry, subPath, rootPath string, ctx ScanContext, opts ResolverOptions) error {
	name := d.Name()

	// If type is a file
	if d.Type().IsRegular() && isMatch(name, ctx.AppName, ctx.BundleID) {
		emitMatch(name, subPath, ctx.MatchesChan, opts)
		if !opts.Peek {
			fmt.Println()
		}

		return nil
	}

	// If type is a symlink
	if d.Type()&os.ModeSymlink != 0 {
		return fs.SkipDir
	}

	// If type is a directory
	if d.Type().IsDir() {
		relPath, err := filepath.Rel(rootPath, subPath)
		if err != nil {
			return nil
		}
		pathSeg := strings.Split(relPath, string(os.PathSeparator))
		depth := len(pathSeg)

		if isMatch(name, ctx.AppName, ctx.BundleID) {
			emitMatch(name, subPath, ctx.MatchesChan, opts)
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

// formatSize takes a size in bytes and returns a human-readable string
func formatSize(size int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	floatSize := float64(size)

	switch {
	case floatSize >= TB:
		return pfmt.ApplyColor(fmt.Sprintf("%.1f TB", floatSize/TB), 196)
	case floatSize >= GB:
		return pfmt.ApplyColor(fmt.Sprintf("%.1f GB", floatSize/GB), 6)
	case floatSize >= MB:
		return pfmt.ApplyColor(fmt.Sprintf("%.1f MB", floatSize/MB), 150)
	case floatSize >= KB:
		return pfmt.ApplyColor(fmt.Sprintf("%.1f KB", floatSize/KB), 70)
	default:
		return pfmt.ApplyColor(fmt.Sprintf("%d B", size), 205)
	}
}

// Strips any ANSI color codes from input string
func stripColor(s string) string {
	var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiEscape.ReplaceAllString(s, "")
}

// Recursively walks filepath and sums up full logical sizes of the files in the dir
func getLogicalSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			// Ignore permission errors, etc.
			return nil
		}
		size += info.Size()
		return nil
	})
	return size, err
}

func getDiskSize(path string) int64 {
	return GetDiskUsageAtPath(path)
}

func (f *Finder) findMatchesApp(rootPath string, ctx ScanContext) {
	entries, err := os.ReadDir(rootPath)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if isMatch(name, ctx.AppName, ctx.BundleID) {
				ctx.MatchesChan <- filepath.Join(rootPath, name)
			}
		}
	}
}

func (f *Finder) findMatchesWalk(rootPath string, ctx ScanContext, opts ResolverOptions) {
	err := filepath.WalkDir(rootPath,
		func(subPath string, d fs.DirEntry, err error) error {
			if err == nil {
				return f.handleScan(d, subPath, rootPath, ctx, opts)
			}

			if os.IsNotExist(err) && f.verbosity {
				fmt.Fprintf(os.Stderr, "Skipped nonexistent path: %s\n", pfmt.ApplyColor(subPath, 3))
			}
			return nil
		})

	if err != nil {
		fmt.Println(" Error on path:", rootPath, err)
	}
}

// Handles '--peek' specific logic incluing
func handlePeek(matches []string, opts ResolverOptions, totalSize int64, numFiles int, appName string) bool {
	// pctx := PeekContext{
	// 	matches:   matches,
	// 	opts:      opts,
	// 	totalSize: totalSize,
	// 	numFiles:  numFiles,
	// 	appName:   appName,
	// }

	// handleSize(pctx)

	maxLen := 0
	fileSizes := make(map[string]string)

	// For all match in matches
	for _, match := range matches {
		fileInfo, err := os.Stat(match) //get info for the file
		if err != nil {
			return opts.Peek
		}
		totalSize += fileInfo.Size() // add file size info to the totalsize
		numFiles++

		// Set human readable string
		humanSize := formatSize(fileInfo.Size())
		fileSizes[match] = humanSize //Pair the human readable size with the path of the match
		if len(stripColor(humanSize)) > maxLen {
			maxLen = len(stripColor(humanSize)) //get max length minus the coloring
		}
	}

	if opts.Peek {
		fmt.Printf("\nFound %s files for %s\n", pfmt.ApplyColor(fmt.Sprintf("%d", numFiles), 3), appName)
		// Holds the match metadata
		type MatchMeta struct {
			Path      string
			SizeStr   string
			PrintLine string
		}

		var metas []MatchMeta
		maxLineWidth := 0

		for _, match := range matches {
			size := getDiskSize(match)
			totalSize += size
			numFiles++

			sizeStr := formatSize(size)
			appColored := pfmt.ApplyColor(appName, 2)
			pathColored := pfmt.ApplyColor(match, 3)

			printLine := fmt.Sprintf("• Match %s FOUND at: %s", appColored, pathColored)
			printLineStripped := fmt.Sprintf("• Match %s FOUND at: %s", appName, match)

			if len(printLineStripped) > maxLineWidth {
				maxLineWidth = len(printLineStripped)
			}

			metas = append(metas, MatchMeta{
				Path:      match,
				SizeStr:   sizeStr,
				PrintLine: printLine,
			})
		}

		// Print all formatted
		for _, meta := range metas {
			lineStripped := stripColor(meta.PrintLine)
			padding := maxLineWidth - len(lineStripped)
			fmt.Printf("%s%s %s\n", meta.PrintLine, strings.Repeat(" ", padding), meta.SizeStr)
		}

		fmt.Printf("→ Total: %s would be freed\n\n", formatSize(totalSize))
		fmt.Println("Run again without -p '--peek' to Trash files or with -f '--force' to delete files")
	}

	// If --peek is enabled send back signal to exit to calling function
	if opts.Peek {
		return true
	} else {
		return false
	}
}
