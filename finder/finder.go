package finder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alewtschuk/rmapp/options"
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
	RootPath    string

	//KMP Additions
	TokenizedApp []string
	LpsArray     []int
}

// Whole Finder struct that holds everything related to finder
type Finder struct {
	OSMain       OSMainPaths
	System       SystemPaths
	UserPaths    UserPaths
	MatchedPaths []string
	Verbosity    bool
	Reported     bool
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
	SystemReceipts              string
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
	GroupContainers     string
	InternetPlugIns     string
	LaunchAgents        string
	Logs                string
	WebKit              string
	ApplicationScripts  string
}

// Creates and loads a new Finder with all needed fields
func NewFinder(appName string, bundleID string, opts options.Options) Finder {
	// Extract home directory for use in user identification if ran as sudo
	home := os.Getenv("HOME")
	finder := searchLayout("/", home)
	finder.Verbosity = opts.Verbosity

	if opts.Peek || opts.Size {
		finder.Reported = true
	} else {
		finder.Reported = false
	}

	matches, err := finder.FindMatches(appName, bundleID, opts)
	if err != nil {
		fmt.Println("NewFinder Error: ", err)
	}
	finder.MatchedPaths = matches
	return finder
}

// Returns a Finder whose search paths are laid out under the system root and home directory
//
// Allows the search paths to be relocated, e.g. into a fixture directory for testing
func searchLayout(root, home string) Finder {
	return Finder{
		OSMain: OSMainPaths{
			RootApplicationsPath: filepath.Join(root, "Applications"),
			UserApplicationsPath: filepath.Join(home, "Applications"),
		},
		System: SystemPaths{
			SystemSupportFilesPath:      filepath.Join(root, "Library/Application Support"),
			SystemCrashReports:          filepath.Join(root, "Library/Application Support/CrashReporter"),
			SystemCaches:                filepath.Join(root, "Library/Caches"),
			SystemExtensions:            filepath.Join(root, "Library/Extensions"),
			SystemInternetPlugIns:       filepath.Join(root, "Library/Internet Plug-Ins"),
			SystemLaunchAgents:          filepath.Join(root, "Library/LaunchAgents"),
			SystemLaunchDaemons:         filepath.Join(root, "Library/LaunchDaemons"),
			SystemLogs:                  filepath.Join(root, "Library/Logs"),
			SystemPrivilegedHelperTools: filepath.Join(root, "Library/PrivilegedHelperTools"),
			SystemReceipts:              filepath.Join(root, "var/db/receipts"),
			SystemBin:                   filepath.Join(root, "usr/local/bin"),
			SystemOpt:                   filepath.Join(root, "usr/local/opt"),
			SystemSbin:                  filepath.Join(root, "usr/local/sbin"),
			SystemShare:                 filepath.Join(root, "usr/local/share"),
			SystemVar:                   filepath.Join(root, "usr/local/var"),
		},
		UserPaths: UserPaths{
			AppSupportFilesPath: filepath.Join(home, "Library/Application Support"),
			PreferencesPath:     filepath.Join(home, "Library/Preferences"),
			CachesPath:          filepath.Join(home, "Library/Caches"),
			ContainersPath:      filepath.Join(home, "Library/Containers"),
			SavedStatePath:      filepath.Join(home, "Library/Saved Application State"),
			HTTPStorages:        filepath.Join(home, "Library/HTTPStorages"),
			GroupContainers:     filepath.Join(home, "Library/Group Containers"),
			InternetPlugIns:     filepath.Join(home, "Library/Internet Plug-Ins"),
			LaunchAgents:        filepath.Join(home, "Library/LaunchAgents"),
			Logs:                filepath.Join(home, "Library/Logs"),
			WebKit:              filepath.Join(home, "Library/WebKit"),
			ApplicationScripts:  filepath.Join(home, "Library/Application Scripts"),
		},
	}
}

// Returns a string of all available paths to search
func (f Finder) AllSearchPaths() []string {
	return []string{
		f.OSMain.RootApplicationsPath,
		f.OSMain.UserApplicationsPath,
		f.System.SystemSupportFilesPath,
		f.System.SystemCrashReports,
		f.System.SystemCaches,
		f.System.SystemExtensions,
		f.System.SystemInternetPlugIns,
		f.System.SystemLaunchAgents,
		f.System.SystemLaunchDaemons,
		f.System.SystemLogs,
		f.System.SystemPrivilegedHelperTools,
		f.System.SystemReceipts,
		f.System.SystemBin,
		f.System.SystemOpt,
		f.System.SystemSbin,
		f.System.SystemShare,
		f.System.SystemVar,
		f.UserPaths.AppSupportFilesPath,
		f.UserPaths.PreferencesPath,
		f.UserPaths.CachesPath,
		f.UserPaths.ContainersPath,
		f.UserPaths.SavedStatePath,
		f.UserPaths.HTTPStorages,
		f.UserPaths.GroupContainers,
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
func (f *Finder) FindMatches(appName, bundleID string, opts options.Options) ([]string, error) {
	var (
		err     error
		matches []string
	)
	tokenizedApp := tokenize(strings.ToLower(appName))
	matchesChan := make(chan string)
	wg := sync.WaitGroup{}

	searchPaths := f.AllSearchPaths()

	if opts.BundleOnly { // if only the bundle is going to be removed only search the main directories
		searchPaths = []string{f.OSMain.RootApplicationsPath, f.OSMain.UserApplicationsPath}
	}

	for _, rootPath := range searchPaths {
		wg.Add(1)

		go func(rootPath string) {
			defer wg.Done()
			// Declared per goroutine as each root has its own depth
			searchDepth := STANDARD_DEPTH
			if rootPath == f.UserPaths.PreferencesPath {
				searchDepth = PREFERENCES_DEPTH
			}

			// Create context struct for passing context to other functions
			ctx := ScanContext{
				AppName:      appName,
				BundleID:     bundleID,
				DomainHint:   GetDomainHint(bundleID),
				SearchDepth:  searchDepth,
				MatchesChan:  matchesChan,
				RootPath:     rootPath,
				TokenizedApp: tokenizedApp,
				LpsArray:     buildLPS(tokenizedApp),
			}

			// Check if root Applications directories hold the .app
			if rootPath == f.OSMain.RootApplicationsPath || rootPath == f.OSMain.UserApplicationsPath {
				f.FindApp(rootPath, ctx)
				return
			}
			f.FindAppFiles(rootPath, ctx, opts)
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

	if opts.Peek || opts.Size {
		GenerateReport(matches, appName, opts)
	}

	return matches, err
}
