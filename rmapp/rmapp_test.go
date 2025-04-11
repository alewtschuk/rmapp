package rmapp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetBundle(t *testing.T) {
	expected := "org.wireshark.Wireshark"
	actual := extractQuotedSubstring("kMDItemCFBundleIdentifier = \"org.wireshark.Wireshark\"")
	if actual != expected {
		t.Errorf("Expected %s but got %s", expected, actual)
	}
}

// func TestFindMatches(t *testing.T) {

// }

func TestWithTimeOut(t *testing.T) {
	timeout := time.After(300 * time.Second)
	done := make(chan bool)
	go func() {
		appname := "Wireshark"
		bundleID := "com.wireshark.Wireshark"
		finder := NewFinder(appname, bundleID, false)
		matches, err := finder.FindMatches(appname, bundleID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(matches)
		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}
}

func TestFindAppInApplicationsDir(t *testing.T) {
	appname := "Wireshark"
	bundleID := "com.wireshark.Wireshark"
	finder := NewFinder(appname, bundleID, false)

	// Only search in Applications directories
	appPaths := []string{
		finder.OSMain.RootApplicationsPath,
		finder.OSMain.UserApplicationsPath,
	}

	var matches []string
	for _, path := range appPaths {
		fmt.Println("Searching in:", path)
		// Directly check if the app exists
		appPath := filepath.Join(path, appname+".app")
		if _, err := os.Stat(appPath); err == nil {
			fmt.Println("Found app at:", appPath)
			matches = append(matches, appPath)
		}
	}

	if len(matches) == 0 {
		t.Errorf("Expected to find %s.app in Applications directories", appname)
	}
}
