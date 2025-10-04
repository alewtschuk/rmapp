package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/alewtschuk/rmapp/deleter"
	"github.com/alewtschuk/rmapp/finder"
	"github.com/alewtschuk/rmapp/options"
)

// --- Test Helpers ---

// makeTestFiles creates n temporary files for testing.
func makeTestFiles(t *testing.T, n int, namePattern string) []string {
	t.Helper()
	var paths []string
	for i := 0; i < n; i++ {
		tmp, err := os.CreateTemp("", fmt.Sprintf("%s-%d-*", namePattern, i))
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		tmp.Close()
		paths = append(paths, tmp.Name())
	}
	return paths
}

// assertSlicesEqual checks if two string slices are equal, ignoring order.
func assertSlicesEqual(t *testing.T, expected, actual []string) {
	t.Helper()
	sort.Strings(expected)
	sort.Strings(actual)

	if len(expected) != len(actual) {
		t.Fatalf("Expected %d paths, but got %d.\nExpected: %v\nActual:   %v", len(expected), len(actual), expected, actual)
	}

	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf("Mismatched paths.\nExpected: %v\nActual:   %v", expected, actual)
		}
	}
}

// setupTestFileSystem creates a temporary directory structure to simulate a user's home directory.
// It returns the path to the fake home directory and a slice of all created file/dir paths that should be found.
func setupTestFileSystem(t *testing.T, appName, bundleID string) (string, []string) {
	t.Helper()
	homeDir := t.TempDir()

	// NOTE: We are not creating a fake .app bundle in /Applications because tests
	// should not write outside their temporary directory. This test assumes the
	// application bundle itself has been found and focuses on finding related files
	// within the user's home directory.
	pathsToCreate := map[string]string{
		"appSupport":  filepath.Join(homeDir, "Library", "Application Support", appName),
		"caches":      filepath.Join(homeDir, "Library", "Caches", bundleID),
		"logs":        filepath.Join(homeDir, "Library", "Logs", appName),
		"preferences": filepath.Join(homeDir, "Library", "Preferences", bundleID+".plist"),
		"randomFile":  filepath.Join(homeDir, "Documents", "a file with "+appName+" in its name.txt"),
	}

	var expectedPaths []string
	for key, path := range pathsToCreate {
		// Create parent dirs and the file/dir itself
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create parent dir for %s: %v", path, err)
		}

		if key == "preferences" || key == "randomFile" {
			f, err := os.Create(path)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", path, err)
			}
			f.Close()
		} else {
			if err := os.Mkdir(path, 0755); err != nil {
				t.Fatalf("Failed to create dir %s: %v", path, err)
			}
		}
		expectedPaths = append(expectedPaths, path)
	}

	return homeDir, expectedPaths
}

// --- Unit Tests ---

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
		hasError bool
	}{
		{`kMDItemCFBundleIdentifier = "com.apple.Safari"`, "com.apple.Safari", false},
		{`"just.quoted.string"`, "just.quoted.string", false},
		{`no quotes here`, "", true},
		{`"incomplete quote`, "", true},
		{`""`, "", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := extractQuotedSubstring(test.input)
			if (err != nil) != test.hasError {
				t.Errorf("extractQuotedSubstring(%q) returned error %v, want error=%v", test.input, err, test.hasError)
			}
			if result != test.expected {
				t.Errorf("extractQuotedSubstring(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

// --- Component Tests ---

func TestDeleterModes(t *testing.T) {
	tests := []struct {
		name     string
		isUnsafe bool
	}{
		{"SafeMode (Trash)", false},
		{"UnsafeMode (Delete)", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Note: The SafeMode test will move temporary files to the system's Trash.
			filesToDelete := makeTestFiles(t, 3, "test-deleter")
			opts := options.Options{Mode: tc.isUnsafe}
			d := deleter.NewDeleter(filesToDelete, opts)

			if err := d.Delete(); err != nil {
				t.Fatalf("Delete failed: %v", err)
			}

			// Check that files are gone from their original locations in both modes.
			for _, path := range filesToDelete {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("Expected file %s to be gone from original location, but it still exists", path)
				}
			}
		})
	}
}

// TestFinder_FindsHomeDirFiles tests the finder's ability to discover files in a controlled environment.
func TestFinder_FindsHomeDirFiles(t *testing.T) {
	// --- Test Setup ---
	appName := "MyTestApp"
	bundleID := "com.gemini.test"

	// 1. Create a fake file system layout in a temporary directory.
	fakeHome, expectedPaths := setupTestFileSystem(t, appName, bundleID)

	// 2. Point the HOME environment variable to our fake home directory.
	// This redirects file searches in `~/Library` to our safe temp directory.
	t.Setenv("HOME", fakeHome)

	// --- Test Execution ---
	// We test the finder directly to avoid calling the real getBundleID, which is not easily mockable.
	opts := options.Options{}
	finder := finder.NewFinder(appName, bundleID, opts)

	// --- Assertions ---
	// The main check: did we find all the files we created in our fake home directory?
	assertSlicesEqual(t, expectedPaths, finder.MatchedPaths)
}
