package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alewtschuk/rmapp/darwin"
	"github.com/alewtschuk/rmapp/deleter"
	"github.com/alewtschuk/rmapp/finder"
	"github.com/alewtschuk/rmapp/options"
)

//NOTE: Blender is hardcoded for sake of simplicity with testing a large free app
// All Blender tests wil fail if Blender is not installed.

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
	opts := options.Options{
		Verbosity: true,
		Mode:      false,
		Peek:      false,
	}
	appname := "Blender"
	bundleID := "org.blenderfoundation.blender"
	finder, _ := finder.NewFinder(appname, bundleID, opts)
	matches := finder.MatchedPaths
	fmt.Println(matches)

}

func TestGetDotApp(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Blender", "Blender.app"},
		{"Blender.app", "Blender.app"},
		{"Some App Name", "Some App Name.app"},
		{"App.With.Dots", "App.With.Dots.app"},
		{"", ".app"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := getDotApp(test.input)
			if result != test.expected {
				t.Errorf("getDotApp(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestExtractQuotedSubstring(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`kMDItemCFBundleIdentifier = "com.apple.Safari"`, "com.apple.Safari"},
		{`kMDItemCFBundleIdentifier = "org.mozilla.firefox"`, "org.mozilla.firefox"},
		{`"just.quoted.string"`, "just.quoted.string"},
		{`no quotes here`, ""},
		{`"incomplete quote`, ""},
		{`""`, ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := extractQuotedSubstring(test.input)
			if result != test.expected {
				t.Errorf("extractQuotedSubstring(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestResolverWithPeekMode(t *testing.T) {
	// Skip if Blender app not likely to be present
	if _, err := os.Stat("/Applications/Blender.app"); os.IsNotExist(err) {
		t.Skip("Blender.app not found - skipping test")
	}

	opts := options.Options{
		Verbosity: false,
		Mode:      false,
		Peek:      true,
		Logical:   true,
	}

	resolver, peeked := NewResolver("Blender", opts)

	if !peeked {
		t.Error("Expected peeked=true with peek=true")
	}

	if len(resolver.Finder.MatchedPaths) == 0 {
		t.Error("Expected to find matches for Blender app")
	}

	if resolver.BundleID == "" {
		t.Error("Expected non-empty bundleID")
	}
}

func TestResolverWithDifferentAppNames(t *testing.T) {
	apps := []string{
		"Obsidian",
		"Ollama",
	}

	for _, app := range apps {
		t.Run(app, func(t *testing.T) {
			// Skip if app not found
			if _, err := os.Stat("/Applications/" + app + ".app"); os.IsNotExist(err) {
				t.Skipf("%s.app not found - skipping test", app)
				return
			}

			opts := options.Options{
				Verbosity: false,
				Mode:      false,
				Peek:      true,
			}

			resolver, _ := NewResolver(app, opts)

			if resolver.AppName != app+".app" {
				t.Errorf("Expected AppName=%s.app, got %s", app, resolver.AppName)
			}

			if resolver.BundleID == "" {
				t.Errorf("Expected non-empty bundleID for %s", app)
			}

			// Check if at least the main app was found
			foundMainApp := false
			for _, path := range resolver.Finder.MatchedPaths {
				if strings.HasSuffix(path, app+".app") {
					foundMainApp = true
					break
				}
			}

			if !foundMainApp {
				t.Errorf("Main app %s.app not found in matches", app)
			}
		})
	}
}

// Testing the actual FindMatchesApp functionality
func TestFinderFindMatchesApp(t *testing.T) {
	// Create a temporary directory to simulate Applications dir
	tmpDir, err := os.MkdirTemp("", "rmapp-test-apps")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create fake .app folders
	testApps := []string{
		"TestApp.app",
		"AnotherTestApp.app",
		"SomeOtherApp.app",
	}

	for _, app := range testApps {
		err := os.Mkdir(filepath.Join(tmpDir, app), 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test FindMatchesApp with our fake directory
	opts := options.Options{
		Verbosity: false,
		Mode:      false,
		Peek:      true,
	}

	f, _ := finder.NewFinder("TestApp", "com.test.app", opts)

	// Create a ScanContext
	matchesChan := make(chan string)
	ctx := finder.ScanContext{
		AppName:     "TestApp",
		BundleID:    "com.test.app",
		SearchDepth: finder.STANDARD_DEPTH,
		MatchesChan: matchesChan,
	}

	// Collect matches in a goroutine
	var matches []string
	done := make(chan bool)

	go func() {
		for match := range matchesChan {
			matches = append(matches, match)
		}
		done <- true
	}()

	// Test with our temp directory
	f.FindMatchesApp(tmpDir, ctx)

	close(matchesChan)
	<-done

	// Check if TestApp.app was found
	found := false
	for _, match := range matches {
		if match == filepath.Join(tmpDir, "TestApp.app") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find TestApp.app in results")
	}
}

// Test for resolver with app name that already has .app extension
func TestResolverWithDotApp(t *testing.T) {
	opts := options.Options{
		Verbosity: false,
		Mode:      false,
		Peek:      true,
	}

	// Test with an app name that already has .app extension
	resolver1, _ := NewResolver("Blender.app", opts)

	// Compare with normal usage
	resolver2, _ := NewResolver("Blender", opts)

	// Both should produce the same AppName
	if resolver1.AppName != resolver2.AppName {
		t.Errorf("Expected consistent AppName regardless of .app suffix, got %s and %s",
			resolver1.AppName, resolver2.AppName)
	}

	// BundleIDs should be the same too
	if resolver1.BundleID != resolver2.BundleID {
		t.Errorf("Expected consistent BundleID regardless of .app suffix, got %s and %s",
			resolver1.BundleID, resolver2.BundleID)
	}
}

func TestTrash(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rmapp-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // clean up if trash fails

	success := darwin.MoveFileToTrash(tmpFile.Name())
	if !success {
		t.Fatalf("Expected to trash %s", tmpFile.Name())
	}
}

func TestMultipleDeleteSAFE(t *testing.T) {
	opts := options.Options{
		Verbosity: true,
		Mode:      false,
		Peek:      false,
	}
	matches := makeTestFiles(4, t)
	d := deleter.NewDeleter(matches, opts)
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
	opts := options.Options{
		Verbosity: true,
		Mode:      true,
		Peek:      false,
	}
	matches := makeTestFiles(4, t)
	d := deleter.NewDeleter(matches, opts)
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
	opts := options.Options{Peek: false, Verbosity: false}
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
