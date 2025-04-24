package finder

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alewtschuk/pfmt"
	"github.com/alewtschuk/rmapp/darwin"
	"github.com/alewtschuk/rmapp/options"
)

// Handles the files/directories if there is a match
//
// Sends all matches to a channel for shared goroutine communication
func (f *Finder) handleScan(d fs.DirEntry, subPath, rootPath string, ctx ScanContext, opts options.Options) error {
	name := d.Name()

	// If type is a file
	if d.Type().IsRegular() && isMatch(name, ctx.AppName, ctx.BundleID) {
		emitMatch(name, subPath, ctx.MatchesChan, opts)
		if !opts.Peek {
			fmt.Println()
		}

		return nil
	}

	// If type is a symlink
	if d.Type()&os.ModeSymlink != 0 {
		return fs.SkipDir
	}

	// If type is a directory
	if d.Type().IsDir() {
		relPath, err := filepath.Rel(rootPath, subPath)
		if err != nil {
			return nil
		}
		pathSeg := strings.Split(relPath, string(os.PathSeparator))
		depth := len(pathSeg)

		if isMatch(name, ctx.AppName, ctx.BundleID) {
			emitMatch(name, subPath, ctx.MatchesChan, opts)
			if !opts.Peek {
				fmt.Println()
			}
			return nil
		}

		if shouldSkipDir(name, depth, ctx) {
			return fs.SkipDir
		}
	}
	return nil
}

func (f *Finder) FindMatchesApp(rootPath string, ctx ScanContext) {
	entries, err := os.ReadDir(rootPath)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if isMatch(name, ctx.AppName, ctx.BundleID) {
				ctx.MatchesChan <- filepath.Join(rootPath, name)
			}
		}
	}
}

func (f *Finder) FindMatchesWalk(rootPath string, ctx ScanContext, opts options.Options) {
	err := filepath.WalkDir(rootPath,
		func(subPath string, d fs.DirEntry, err error) error {
			if err == nil {
				return f.handleScan(d, subPath, rootPath, ctx, opts)
			}

			if os.IsNotExist(err) && f.verbosity {
				fmt.Fprintf(os.Stderr, "Skipped nonexistent path: %s\n", pfmt.ApplyColor(subPath, 3))
			}
			return nil
		})

	if err != nil {
		fmt.Println("[rmapp] Error on path:", rootPath, err)
	}
}

// Recursively walks filepath and sums up full logical sizes of the files in the dir
func getLogicalSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info fs.FileInfo, err error) error {
		if err == nil {
			size += info.Size()
			return nil
		}
		return nil
	})
	return size
}

func GetDiskSize(path string) int64 {
	return darwin.GetDiskUsageAtPath(path)
}
