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
	// Extract home directory for use in user identification if ran as sudo
	home := os.Getenv("HOME")
	fmt.Printf("Current HOME: %s\n", home)
	finder := Finder{
		OSMain: OSMainPaths{
			RootApplicationsPath: "/Applications",
			UserApplicationsPath: fmt.Sprintf("%s/Applications", home),
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
			AppSupportFilesPath: fmt.Sprintf("%s/Library/Application Support", home),
			PreferencesPath:     fmt.Sprintf("%s/Library/Preferences", home),
			CachesPath:          fmt.Sprintf("%s/Library/Caches", home),
			ContainersPath:      fmt.Sprintf("%s/Library/Containers", home),
			SavedStatePath:      fmt.Sprintf("%s/Library/Saved Application State", home),
			HTTPStorages:        fmt.Sprintf("%s/Library/HTTPStorages", home),
			GroupContainers:     fmt.Sprintf("%s/Library/Group Containers", home),
			InternetPlugIns:     fmt.Sprintf("%s/Library/Internet Plug-Ins", home),
			LaunchAgents:        fmt.Sprintf("%s/Library/LaunchAgents", home),
			Logs:                fmt.Sprintf("%s/Library/Logs", home),
			WebKit:              fmt.Sprintf("%s/Library/WebKit", home),
			ApplicationScripts:  fmt.Sprintf("%s/Library/Application Scripts", home),
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
		err         error
		matches     []string
		searchDepth int
	)
	matchesChan := make(chan string)
	wg := sync.WaitGroup{}

	for _, rootPath := range f.AllSearchPaths() {
		wg.Add(1)

		go func(rootPath string) {
			defer wg.Done()
			if rootPath == f.UserPaths.PreferencesPath {
				searchDepth = PREFERENCES_DEPTH
			} else {
				searchDepth = STANDARD_DEPTH
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
