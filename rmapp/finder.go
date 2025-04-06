package rmapp

import (
	"fmt"
	"os"
)

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

func NewFinder() Finder {
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
			CachesPath:          fmt.Sprintf("/User/%s/Caches", os.Getenv("USER")),
			ContainersPath:      fmt.Sprintf("/Users/%s/Library/Containers", os.Getenv("USER")),
			SavedStatePath:      fmt.Sprintf("/Users/%s/Library/Saved Application State", os.Getenv("USER")),
		},
	}
}
