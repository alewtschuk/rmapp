package rmapp

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Whole Finder struct that holds everything related to finder
type Finder struct {
	OSMain       OSMainPaths
	System       SystemPaths
	UserPaths    UserPaths
	MatchedPaths []string
}

// The default os directories where the .app file should exist
type OSMainPaths struct {
	RootApplicationsPath string // default os applications path
	UserApplicationsPath string // default user applications path
}

// Directories where the system wide paths is stored
type SystemPaths struct {
	GlobalSupportFilesPath     string
	GlobalPreferencesFilesPath string
}

// Directories holding user specific paths
type UserPaths struct {
	AppSupportFilesPath string
	PreferencesPath     string
	CachesPath          string
	ContainersPath      string
	SavedStatePath      string
}

func NewFinder(appName string, bundleID string) Finder {
	return Finder{
		OSMain: OSMainPaths{
			RootApplicationsPath: "/Applications",
			UserApplicationsPath: fmt.Sprintf("/Users/%s/Applications", os.Getenv("USER")),
		},
		System: SystemPaths{
			GlobalSupportFilesPath:     "/Library/Application Support",
			GlobalPreferencesFilesPath: "/Library/Preferences",
		},
		UserPaths: UserPaths{
			AppSupportFilesPath: fmt.Sprintf("/Users/%s/Library/Application Support", os.Getenv("USER")),
			PreferencesPath:     fmt.Sprintf("/Users/%s/Library/Preferences", os.Getenv("USER")),
			CachesPath:          fmt.Sprintf("/Users/%s/Library/Caches", os.Getenv("USER")),
			ContainersPath:      fmt.Sprintf("/Users/%s/Library/Containers", os.Getenv("USER")),
			SavedStatePath:      fmt.Sprintf("/Users/%s/Library/Saved Application State", os.Getenv("USER")),
		},
	}
}

func (f Finder) AllSearchPaths() []string {
	return []string{
		f.OSMain.RootApplicationsPath,
		f.OSMain.UserApplicationsPath,
		f.System.GlobalSupportFilesPath,
		f.System.GlobalPreferencesFilesPath,
		f.UserPaths.AppSupportFilesPath,
		f.UserPaths.PreferencesPath,
		f.UserPaths.CachesPath,
		f.UserPaths.ContainersPath,
		f.UserPaths.SavedStatePath,
	}
}

func (f *Finder) FindMatches(appname, bundleID string) ([]string, error) {
	var err error
	var matches []string
	matchesChan := make(chan string) // creates a string channel
	wg := sync.WaitGroup{}           // new waitgroup

	for _, path := range f.AllSearchPaths() {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {

				// Handle directory or file types
				if d.Type()&os.ModeSymlink != 0 {
					return nil
				} else if d.Type().IsDir() {
					if strings.Contains(strings.ToLower(path), bundleID) || strings.Contains(strings.ToLower(path), appname) {
						fmt.Println("Path is: ", path)
						fmt.Println("File name is: ", d.Name())
						fmt.Println("File type: ", d.Type())
						matchesChan <- path
						return nil
					} else {
						return nil
					}
				} else if d.Type().IsRegular() {
					if strings.Contains(strings.ToLower(path), bundleID) || strings.Contains(strings.ToLower(path), appname) {
						fmt.Println("Path is: ", path)
						fmt.Println("File name is: ", d.Name())
						fmt.Println("File type: ", d.Type())
						matchesChan <- path
						return nil
					} else {
						return nil
					}
				}

				return nil
			})
		}(path)
	}

	// Go routine to close the channel
	go func() {
		wg.Wait()
		close(matchesChan)
	}()

	for match := range matchesChan {
		matches = append(matches, match)
	}
	fmt.Println(matches)

	return matches, err
}

func visit(matchesChan chan string, appname, bundleID, path string, di fs.DirEntry, err error) error {
	if strings.Contains(path, bundleID) || strings.Contains(path, appname) {
		match := path
		matchesChan <- match
		return nil
	} else {
		return err
	}
}
