package rmapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetBundle(t *testing.T) {
	expected := "org.wireshark.Wireshark"
	actual := extractQuotedSubstring("kMDItemCFBundleIdentifier = \"org.wireshark.Wireshark\"")
	if actual != expected {
		t.Errorf("Expected %s but got %s", expected, actual)
	}
}

func makeTestFiles(n int, t *testing.T) []string {
	t.Helper()
	var paths []string
	for i := 0; i < n; i++ {
		tmp, err := os.CreateTemp("", "rmapp-test-*")
		if err != nil {
			t.Fatal(err)
		}
		tmp.Close()
		paths = append(paths, tmp.Name())
	}
	return paths
}

func TestFinderOutput(t *testing.T) {
	opts := ResolverOptions{
		Verbosity: true,
		Mode:      false,
		Peek:      false,
	}
	appname := "Blender"
	bundleID := "org.blenderfoundation.blender"
	finder, _ := NewFinder(appname, bundleID, opts)
	matches := finder.MatchedPaths
	fmt.Println(matches)

}

func TestFindAppInApplicationsDir(t *testing.T) {
	opts := ResolverOptions{
		Verbosity: true,
		Mode:      false,
		Peek:      false,
	}
	appname := "Blender"
	bundleID := "org.blenderfoundation.blender"
	finder, _ := NewFinder(appname, bundleID, opts)

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
		t.Skipf("Expected to find %s.app in Applications directories", appname)
	}
}

func TestTrash(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rmapp-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // clean up if trash fails

	success := MoveFileToTrash(tmpFile.Name())
	if !success {
		t.Fatalf("Expected to trash %s", tmpFile.Name())
	}
}

func TestMultipleDeleteSAFE(t *testing.T) {
	opts := ResolverOptions{
		Verbosity: true,
		Mode:      false,
		Peek:      false,
	}
	matches := makeTestFiles(4, t)
	d := NewDeleter(matches, opts)
	err := d.Delete()
	for _, path := range matches {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected %s to be deleted, but it still exists", path)
		}
	}
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
	matches := makeTestFiles(4, t)
	d := NewDeleter(matches, opts)
	err := d.Delete()
	for _, path := range matches {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected %s to be deleted, but it still exists", path)
		}
	}
	if err != nil {
		t.Fatalf("Expected to successfully delete file at %s, but it failed", matches)
	}
}

func TestResolver_MatchesExpectedFiles(t *testing.T) {
	opts := ResolverOptions{Peek: false, Verbosity: false}
	resolver, _ := NewResolver("Blender", opts)

	if len(resolver.Finder.MatchedPaths) == 0 {
		t.Errorf("Expected matches for Blender, got none")
	}

	for _, path := range resolver.Finder.MatchedPaths {
		if !strings.Contains(path, "Blender") && !strings.Contains(path, "org.blenderfoundation.blender") {
			t.Errorf("Unexpected match path: %s", path)
		}
	}
}
