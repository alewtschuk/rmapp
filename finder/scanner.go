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
	symlink := false

	// If type is a file
	if d.Type().IsRegular() && f.isMatch(name, ctx) {
		f.emitMatch(name, subPath, ctx.MatchesChan, opts, symlink)
		if !opts.Peek {
			fmt.Println()
		}

		return nil
	}

	// If type is a symlink check symlink bit and if symlink contains match hueristics emit match
	// Used to prevent dangling symlinks
	if d.Type()&os.ModeSymlink != 0 && f.isMatch(name, ctx) {
		symlink = true
		f.emitMatch(name, subPath, ctx.MatchesChan, opts, symlink)
		if !opts.Peek {
			fmt.Println()
		}

		return nil
	}

	// If type is a directory
	if d.Type().IsDir() {
		relPath, err := filepath.Rel(rootPath, subPath)
		if err != nil {
			return nil
		}
		pathSeg := strings.Split(relPath, string(os.PathSeparator))
		depth := len(pathSeg)

		if f.isMatch(name, ctx) {
			f.emitMatch(name, subPath, ctx.MatchesChan, opts, symlink)
			if !opts.Peek {
				fmt.Println()
			}
			return fs.SkipDir
		}

		if f.shouldSkipDir(name, depth, ctx) {
			return fs.SkipDir
		}
	}
	return nil
}

// Checks if the rootPath /Applications or ~/Applications contains the .app bundle.
//
// Uses directory scanning and handling as .app bundles are a
// specially defined directory type in MacOS, even though they contain
// a filetype identier
func (f *Finder) FindApp(rootPath string, ctx ScanContext) {
	applications, err := os.ReadDir(rootPath) // get all directories in the rootPath
	if err == nil {
		// Check each .app bundle, extract the name, check for match and send to channel
		for _, app := range applications {
			if !app.IsDir() {
				continue
			}
			name := app.Name()
			if f.isMatch(name, ctx) {
				ctx.MatchesChan <- filepath.Join(rootPath, name) // send full path for the channel
			}
		}
	}
}

// Walks the directory, ensures theres no error, passes to handle scan for further subpath walking
func (f *Finder) FindAppFiles(rootPath string, ctx ScanContext, opts options.Options) {
	err := filepath.WalkDir(rootPath,
		func(subPath string, d fs.DirEntry, err error) error {
			if err == nil {
				return f.handleScan(d, subPath, rootPath, ctx, opts)
			}

			// if os.IsNotExist(err) && f.verbosity {
			// 	fmt.Fprintf(os.Stderr, "Skipped nonexistent path: %s\n", pfmt.ApplyColor(subPath, 3))
			// }

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

// Gets the file size in bytes
func GetDiskSize(path string) int64 {
	return darwin.GetDiskUsageAtPath(path)
}

// Decide if a directory should be skipped based on context
func (f Finder) shouldSkipDir(name string, depth int, ctx ScanContext) bool {
	if depth > ctx.SearchDepth {
		return true
	}
	if (ctx.SearchDepth == STANDARD_DEPTH && depth < ctx.SearchDepth) && (strings.Contains(name, ctx.DomainHint) && !f.isMatch(name, ctx)) {
		return true
	}
	return false
}

// Helper function to print and send matches to channel
func (f *Finder) emitMatch(name, path string, matchesChan chan string, opts options.Options, symlink bool) {
	if opts.Verbosity && !opts.Peek && !symlink {
		fmt.Printf("Match %s FOUND at: %s", pfmt.ApplyColor(name, 2), pfmt.ApplyColor(path, 3))
	} else if opts.Verbosity && !opts.Peek && symlink {
		fmt.Printf("Symlink match %s FOUND at: %s", pfmt.ApplyColor(name, 2), pfmt.ApplyColor(path, 3))
	}

	matchesChan <- path
}
