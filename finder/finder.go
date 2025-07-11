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
	user := os.Getenv("USER")
	finder := Finder{
		OSMain: OSMainPaths{
			RootApplicationsPath: "/Applications",
			UserApplicationsPath: fmt.Sprintf("/Users/%s/Applications", user),
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
			SystemReceipts:              "/var/db/receipts",
			SystemBin:                   "/usr/local/bin",
			SystemOpt:                   "/usr/local/opt",
			SystemSbin:                  "/usr/local/sbin",
			SystemShare:                 "/usr/local/share",
			SystemVar:                   "/usr/local/var",
		},
		UserPaths: UserPaths{
			AppSupportFilesPath: fmt.Sprintf("/Users/%s/Library/Application Support", user),
			PreferencesPath:     fmt.Sprintf("/Users/%s/Library/Preferences", user),
			CachesPath:          fmt.Sprintf("/Users/%s/Library/Caches", user),
			ContainersPath:      fmt.Sprintf("/Users/%s/Library/Containers", user),
			SavedStatePath:      fmt.Sprintf("/Users/%s/Library/Saved Application State", user),
			HTTPStorages:        fmt.Sprintf("/Users/%s/Library/HTTPStorages", user),
			GroupContainers:     fmt.Sprintf("/Users/%s/Library/Group Containers", user),
			InternetPlugIns:     fmt.Sprintf("/Users/%s/Library/Internet Plug-Ins", user),
			LaunchAgents:        fmt.Sprintf("/Users/%s/Library/LaunchAgents", user),
			Logs:                fmt.Sprintf("/Users/%s/Library/Logs", user),
			WebKit:              fmt.Sprintf("/Users/%s/Library/WebKit", user),
			ApplicationScripts:  fmt.Sprintf("/Users/%s/Library/Application Scripts", user),
		},
		verbosity: opts.Verbosity,
	}
	matches, err := finder.FindMatches(appName, bundleID, opts)
	if err != nil {
		fmt.Println("NewFinder Error: ", err)
	}

	finder.MatchedPaths = matches
	return finder
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
	matchesChan := make(chan string)
	wg := sync.WaitGroup{}

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
				DomainHint:  GetDomainHint(bundleID),
				SearchDepth: searchDepth,
				MatchesChan: matchesChan,
			}

			// Check if root Applications directories hold the .app
			if rootPath == f.OSMain.RootApplicationsPath || rootPath == f.OSMain.UserApplicationsPath {
				f.FindApp(rootPath, ctx)
				return
			}
			// For all other scanned directories we need to walk
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

	if opts.Peek {
		GeneratePeekReport(matches, appName, opts)
	}

	return matches, err
}
