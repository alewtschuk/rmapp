package finder

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/alewtschuk/rmapp/options"
)

// Benchmarks how quickly FindMatches gathers an app's data when it is scattered
// across the search roots and surrounded by unrelated data
//
// Run with: go test ./finder -run '^$' -bench FindMatches -benchmem

const (
	benchAppName     = "Acme Studio"
	benchBundleID    = "com.acme.studio"
	noiseFilesPerDir = 4
)

func BenchmarkFindMatches(b *testing.B) {
	for _, noise := range []int{100, 1000} { // unrelated entries per search root
		for _, scatter := range []int{8, 64} { // related artifacts spread across the roots
			b.Run(fmt.Sprintf("noise=%d/scatter=%d", noise, scatter), func(b *testing.B) {
				f, expected, entries := buildFixture(b, noise, scatter)
				opts := options.Options{}

				// Ensure the finder is actually finding the scattered data before timing it
				matches, err := f.FindMatches(benchAppName, benchBundleID, opts)
				if err != nil {
					b.Fatal(err)
				}
				slices.Sort(matches)
				if !slices.Equal(matches, expected) {
					b.Fatalf("found %d paths, expected %d\nfound:    %v\nexpected: %v", len(matches), len(expected), matches, expected)
				}

				for b.Loop() {
					f.FindMatches(benchAppName, benchBundleID, opts)
				}

				b.ReportMetric(float64(entries), "entries")
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(entries), "ns/entry")
			})
		}
	}
}

// Sweeps the worker pool size over one large fixture to find where added
// concurrency stops paying off. 29 workers is one per root, the old behavior
func BenchmarkFindMatchesWorkers(b *testing.B) {
	f, expected, entries := buildFixture(b, 1000, 64)
	opts := options.Options{}

	for _, workers := range []int{1, 2, 4, 8, 29} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			f.Workers = workers

			matches, err := f.FindMatches(benchAppName, benchBundleID, opts)
			if err != nil {
				b.Fatal(err)
			}
			slices.Sort(matches)
			if !slices.Equal(matches, expected) {
				b.Fatalf("found %d paths, expected %d", len(matches), len(expected))
			}

			for b.Loop() {
				f.FindMatches(benchAppName, benchBundleID, opts)
			}

			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(entries), "ns/entry")
		})
	}
}

// Builds a fake filesystem under a temp dir mirroring the real search roots
//
// Returns the finder pointed at it, the sorted paths it should match
// and the total number of entries created
func buildFixture(b *testing.B, noise, scatter int) (*Finder, []string, int) {
	b.Helper()
	start := time.Now()
	base := b.TempDir()
	f := searchLayout(base, filepath.Join(base, "Users", "bench"))
	entries := 0

	mkdir := func(path string) {
		if err := os.MkdirAll(path, 0755); err != nil {
			b.Fatal(err)
		}
		entries++
	}
	touch := func(path string) {
		if err := os.WriteFile(path, nil, 0644); err != nil {
			b.Fatal(err)
		}
		entries++
	}

	// App bundles, only the bundle directory itself is inspected so its contents are irrelevant
	for _, root := range []string{f.OSMain.RootApplicationsPath, f.OSMain.UserApplicationsPath} {
		mkdir(root)
		for k := range noise {
			mkdir(filepath.Join(root, fmt.Sprintf("Vendor%d Product%d.app", k, k)))
		}
	}
	expected := []string{filepath.Join(f.OSMain.RootApplicationsPath, benchAppName+".app")}
	mkdir(expected[0])

	// CrashReporter is nested inside Application Support and walked by both roots, so it only gets noise
	dataRoots := slices.DeleteFunc(f.AllSearchPaths()[2:], func(root string) bool {
		return root == f.System.SystemCrashReports
	})

	// Unrelated vendor data, every tenth entry shares the vendor or a similar name to exercise the matcher
	for _, root := range dataRoots {
		mkdir(root)
		for k := range noise {
			var name string
			switch {
			case k%20 == 0:
				name = fmt.Sprintf("com.acme.editor%d", k)
			case k%20 == 10:
				name = fmt.Sprintf("Acme Studios Archive %d", k)
			case k%2 == 0:
				name = fmt.Sprintf("com.vendor%d.product%d", k, k)
			default:
				name = fmt.Sprintf("Vendor%d Product%d", k, k)
			}
			dir := filepath.Join(root, name)
			mkdir(dir)
			for n := range noiseFilesPerDir {
				touch(filepath.Join(dir, fmt.Sprintf("data%d.db", n)))
			}
			touch(filepath.Join(root, fmt.Sprintf("com.vendor%d.product%d.plist", k, k)))
		}
	}

	// Related data spread round robin across the roots using the naming styles apps use in practice
	for i := range scatter {
		root := dataRoots[i%len(dataRoots)]
		var path string
		switch i % 7 {
		case 0: // e.g. ~/Library/Caches/com.acme.studio
			path = filepath.Join(root, fmt.Sprintf("%s.%d", benchBundleID, i))
			mkdir(path)
			touch(filepath.Join(path, "Cache.db"))
		case 1: // e.g. ~/Library/Preferences/com.acme.studio.plist
			path = filepath.Join(root, fmt.Sprintf("%s.%d.plist", benchBundleID, i))
			touch(path)
		case 2: // e.g. ~/Library/Group Containers/TEAMID.com.acme.studio
			path = filepath.Join(root, fmt.Sprintf("ABCDE12345.%s.group%d", benchBundleID, i))
			mkdir(path)
		case 3: // e.g. ~/Library/Application Support/Acme Studio
			path = filepath.Join(root, fmt.Sprintf("%s %d", benchAppName, i))
			mkdir(path)
			touch(filepath.Join(path, "settings.json"))
		case 4: // e.g. ~/Library/Logs/acme-studio-helper.log
			path = filepath.Join(root, fmt.Sprintf("acme-studio-helper-%d.log", i))
			touch(path)
		case 5: // e.g. ~/Library/Preferences/ByHost/com.acme.studio.UUID.plist, nested one level in
			path = filepath.Join(root, "Vendor1 Product1", fmt.Sprintf("%s.plugin%d", benchBundleID, i))
			touch(path)
		case 6: // e.g. /usr/local/bin/acme-studio, left dangling as the bundle it points into is gone
			path = filepath.Join(root, fmt.Sprintf("acme-studio-%d", i))
			if err := os.Symlink(filepath.Join(base, "missing", "acme-studio"), path); err != nil {
				b.Fatal(err)
			}
			entries++
		}
		expected = append(expected, path)
	}

	slices.Sort(expected)
	b.Logf("fixture: %d entries, %d related paths (built in %s)", entries, len(expected), time.Since(start))
	return &f, expected, entries
}
