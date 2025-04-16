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
	opts := ResolverOptions{
		Verbosity: true,
		Mode:      false,
		Peek:      false,
	}
	timeout := time.After(300 * time.Second)
	done := make(chan bool)
	go func() {
		appname := "Wireshark"
		bundleID := "com.wireshark.Wireshark"
		finder := NewFinder(appname, bundleID, opts)
		matches, err := finder.FindMatches(appname, bundleID, opts)
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
	opts := ResolverOptions{
		Verbosity: true,
		Mode:      false,
		Peek:      false,
	}
	appname := "Wireshark"
	bundleID := "com.wireshark.Wireshark"
	finder := NewFinder(appname, bundleID, opts)

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

func TestTrash(t *testing.T) {
	path := "/Users/alexlewtschuk/Desktop/removeme"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("Test file does not exist: %s", path)
	}
	success := MoveFileToTrash(path)
	if !success {
		t.Fatalf("Expected to successfully trash file at %s, but it failed", path)
	}
}

func TestDeleteSAFE(t *testing.T) {
	opts := ResolverOptions{
		Verbosity: true,
		Mode:      false,
		Peek:      false,
	}
	matches := []string{"/Users/alexlewtschuk/Desktop/removeme", "/Users/alexlewtschuk/Desktop/rmshot.png", "/Users/alexlewtschuk/Desktop/rem"}
	d := NewDeleter(matches, opts)
	err := d.Delete()
	if err != nil {
		t.Fatalf("Expected to successfully delete file at %s, but it failed", matches)
	}
}

func TestDeleteUNSAFE(t *testing.T) {
	opts := ResolverOptions{
		Verbosity: true,
		Mode:      true,
		Peek:      false,
	}
	matches := []string{"/Users/alexlewtschuk/Desktop/removeme", "/Users/alexlewtschuk/Desktop/rmshot.png", "/Users/alexlewtschuk/Desktop/rem"}
	d := NewDeleter(matches, opts)
	err := d.Delete()
	if err != nil {
		t.Fatalf("Expected to successfully delete file at %s, but it failed", matches)
	}
}

func TestPeek(t *testing.T) {
	opts := ResolverOptions{
		Verbosity: true,
		Mode:      false,
		Peek:      true,
	}
	matches := []string{"/Users/alexlewtschuk/Desktop/removeme", "/Users/alexlewtschuk/Desktop/rmshot.png", "/Users/alexlewtschuk/Desktop/rem"}
	d := NewDeleter(matches, opts)
	err := d.Delete()
	if err != nil {
		t.Fatalf("Expected to successfully delete file at %s, but it failed", matches)
	}
}
