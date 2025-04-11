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
}

// Creates and loads a new Finder with all needed fields
func NewFinder(appName string, bundleID string, verbose bool) Finder {
	finder := Finder{
		OSMain: OSMainPaths{
			RootApplicationsPath: "/Applications",
			UserApplicationsPath: fmt.Sprintf("/Users/%s/Applications", os.Getenv("USER")),
		},
		System: SystemPaths{
			GlobalSupportFilesPath: "/Library/Application Support",
			//GlobalPreferencesFilesPath: "/Library/Preferences", //NOTE: Dir doesn't seem to hold user installed app data. Disabling for now
		},
		UserPaths: UserPaths{
			AppSupportFilesPath: fmt.Sprintf("/Users/%s/Library/Application Support", os.Getenv("USER")),
			PreferencesPath:     fmt.Sprintf("/Users/%s/Library/Preferences", os.Getenv("USER")),
			CachesPath:          fmt.Sprintf("/Users/%s/Library/Caches", os.Getenv("USER")),
			ContainersPath:      fmt.Sprintf("/Users/%s/Library/Containers", os.Getenv("USER")),
			SavedStatePath:      fmt.Sprintf("/Users/%s/Library/Saved Application State", os.Getenv("USER")),
		},
		verbosity: verbose,
	}
	matches, err := finder.FindMatches(appName, bundleID)
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
		f.System.GlobalSupportFilesPath,
		//f.System.GlobalPreferencesFilesPath, //NOTE: Dir doesn't seem to hold user installed app data. Disabling for now
		f.UserPaths.AppSupportFilesPath,
		f.UserPaths.PreferencesPath,
		f.UserPaths.CachesPath,
		f.UserPaths.ContainersPath,
		f.UserPaths.SavedStatePath,
	}
}

// Walks the filepath for each path available and checks if each path contains a match
// to the bundleID or the appname.
//
// Internal WalkDir function passes matches to a channel which will be read from to
// build a string slice of matched paths that will be flagged for deletion
func (f *Finder) FindMatches(appname, bundleID string) ([]string, error) {
	var err error
	var matches []string
	matchesChan := make(chan string) // creates a string channel
	wg := sync.WaitGroup{}           // new waitgroup

	// Spin up a go routine for each path available to search in parallel
	// Uses closure pattern to enable internal function to read needed vars
	for _, rootPath := range f.AllSearchPaths() {
		wg.Add(1)
		// fmt.Println("Current path is: " + rootPath)
		// fmt.Println("Added to waitgroup")

		go func(rootPath string) {
			defer wg.Done()
			err := filepath.WalkDir(rootPath, func(subPath string, d fs.DirEntry, err error) error {
				switch rootPath {
				// Handle main root application directories
				case f.OSMain.RootApplicationsPath, f.OSMain.UserApplicationsPath:
					name := d.Name()
					relPath, _ := filepath.Rel(rootPath, subPath)
					pathSeg := strings.Split(relPath, string(os.PathSeparator))
					depth := len(pathSeg)

					// Only handle directory as .app files are a subtype of directory
					if d.Type().IsDir() && isMatch(name, appname, bundleID) {
						if f.verbosity {
							fmt.Printf("Match %s FOUND at: %s\n", pfmt.ApplyColor(name, 2), pfmt.ApplyColor(subPath, 3))
						}
						matchesChan <- subPath // send matched path to the channel and stop traversing
						return nil
					}
					if depth > STANDARD_DEPTH {
						return fs.SkipDir
					}

				// Handle all cases where STANDARD_DEPTH = 1
				case f.System.GlobalSupportFilesPath, f.UserPaths.AppSupportFilesPath, f.UserPaths.CachesPath, f.UserPaths.ContainersPath, f.UserPaths.SavedStatePath:
					err := f.handleScan(d, appname, bundleID, subPath, rootPath, matchesChan, STANDARD_DEPTH)
					if err != nil {
						return err
					}

				// Handle case where PREFERENCES_DEPTH = 2
				case f.UserPaths.PreferencesPath:
					err := f.handleScan(d, appname, bundleID, subPath, rootPath, matchesChan, PREFERENCES_DEPTH)
					if err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				fmt.Println(" Error on path:", rootPath, err)
			}
		}(rootPath) // call the function with the path value
	}

	// Go routine to close the channel
	go func() {
		wg.Wait()
		close(matchesChan)
	}()

	// Read from channel and add to matches
	for match := range matchesChan {
		matches = append(matches, match)
	}

	return matches, err
}

// Checks if the file/directory name contains the appName or bundleID
func isMatch(name, appName, bundleID string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, strings.ToLower(bundleID)) || strings.Contains(strings.ToLower(name), strings.ToLower(appName))
}

// Handles the files/directories if there is a match
//
// Sends all matches to a channel for shared goroutine communication
func (f *Finder) handleScan(d fs.DirEntry, appName, bundleID, subPath, rootPath string, matchesChan chan string, searchDepth int) error {
	// Handle file if regular
	if d.Type().IsRegular() {
		name := d.Name()
		//fmt.Println("File name is: ", name)
		if isMatch(name, appName, bundleID) {
			if f.verbosity {
				fmt.Printf("Match %s FOUND at: %s\n", pfmt.ApplyColor(name, 2), pfmt.ApplyColor(subPath, 3))
			}
			matchesChan <- subPath // send matched path to channel and stop traversing
			return nil
		}
	}

	// Handle and skip symlink
	if d.Type()&os.ModeSymlink != 0 {
		return fs.SkipDir
	}

	// Handle directory
	if d.Type().IsDir() {
		name := d.Name()
		relPath, _ := filepath.Rel(rootPath, subPath)
		pathSeg := strings.Split(relPath, string(os.PathSeparator))
		depth := len(pathSeg)

		// Check if direcotry matches
		if isMatch(name, appName, bundleID) {
			if f.verbosity {
				fmt.Printf("Match %s FOUND at: %s\n", pfmt.ApplyColor(name, 2), pfmt.ApplyColor(subPath, 3))
			}
			matchesChan <- subPath // send matched path to the channel and stop traversing
			return nil
		}

		// Check if the searchDepth is STANDARD_DEPTH = 1 or PREFERENCES_DEPTH = 2
		// If standard then skip all unneeded subdirs
		if searchDepth == 1 && (depth < searchDepth && (strings.Contains(name, ".") && !isMatch(name, appName, bundleID))) {
			//fmt.Println("Skipping . directory " + subPath + " with no match: " + name)
			return fs.SkipDir
		}

		if depth > searchDepth {
			//fmt.Println("Skipping directory due to depth", d.Name())
			return fs.SkipDir
		}
	}
	return nil
}
