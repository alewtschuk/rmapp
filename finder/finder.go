package finder

import (
	"fmt"
	"os"
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
func NewFinder(appName string, bundleID string, opts options.Options) (Finder, bool) {
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
func (f *Finder) FindMatches(appName, bundleID string, opts options.Options) ([]string, bool, error) {
	var (
		err     error
		matches []string
	)
	matchesChan := make(chan string)
	wg := sync.WaitGroup{}

	domainHint := GetDomainHint(bundleID)

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
				f.FindMatchesApp(rootPath, ctx)
				return
			}
			// For all other scanned directories we need to walk
			f.FindMatchesWalk(rootPath, ctx, opts)
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

	if opts.Peek {
		GeneratePeekReport(matches, appName, opts)
	}

	return matches, opts.Peek, err
}
