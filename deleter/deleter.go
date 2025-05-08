package deleter

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/alewtschuk/pfmt"
	"github.com/alewtschuk/rmapp/darwin"
	"github.com/alewtschuk/rmapp/finder"
	"github.com/alewtschuk/rmapp/options"
)

// Define the Deleter and its fields
type Deleter struct {
	matches      []string
	opts         options.Options
	data         []data
	totalDeleted int64 // actual amount of data removed
	totalSize    int64 // total size that should be removed
}

// Define data struct
type data struct {
	match   string // path to file
	size    int64  // file size
	deleted bool   // has it been deleted
}

// Creates and returns the Deleter
func NewDeleter(matches []string, opts options.Options) Deleter {
	return Deleter{
		matches: matches,
		opts:    opts,
	}
}

// Handles deletion logic based on execution mode
//
// Creates go routine for each individual match.
func (d *Deleter) Delete() error {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, match := range d.matches {
		size := finder.GetDiskSize(match)
		d.totalSize += size
		data := data{
			match:   match,
			size:    size,
			deleted: false,
		}
		d.data = append(d.data, data)
	}

	switch d.opts.Mode {
	case false: // default trashing behavior
		for idx, match := range d.matches {
			wg.Add(1)
			go func(idx int, match string) error {
				defer wg.Done()
				err := exists(match)
				if err != nil {
					return err
				}

				// Call native interop to move to trash and check success
				success := darwin.MoveFileToTrash(match)
				matchSize := d.data[idx].size
				if !success {
					fmt.Println(pfmt.ApplyColor("WARN: file "+match+" is sandboxed and SIP protected", 3))
					fmt.Println("Attempting trashing via osascript...")
					cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, match))
					err = cmd.Run() // ensure command runs without err

					// If command ran without issue and exists is not nil (file doesnt exist)
					mu.Lock()
					if err == nil {
						// Check if the file still exists
						if exists(match) != nil {
							// File was successfully deleted
							d.data[idx].deleted = true
							d.totalDeleted += matchSize
							fmt.Printf("Successfully moved %s to Trash 🗑️\n", pfmt.ApplyColor(match, 3))
						} else {
							// File still exists
							d.data[idx].deleted = false
							fmt.Println(pfmt.ApplyColor("[rmapp] ERROR: file "+match+" unable to be moved to Trash", 9))
						}
					} else {
						// osascript failed
						d.data[idx].deleted = false
						fmt.Println(pfmt.ApplyColor("[rmapp] ERROR: file "+match+" unable to be moved to Trash", 9))
					}
					mu.Unlock()

					//Ensure verbose output only prints if file exists
					if d.opts.Verbosity && d.data[idx].deleted {
						fmt.Printf("Successfully moved %s to Trash 🗑️\n", pfmt.ApplyColor(match, 3))
					}

					return err
				} else {
					mu.Lock()
					d.data[idx].deleted = true
					d.totalDeleted += matchSize
					mu.Unlock()
				}

				if d.opts.Verbosity {
					mu.Lock()
					if d.data[idx].deleted {
						fmt.Printf("Successfully moved %s to Trash 🗑️\n", pfmt.ApplyColor(match, 3))
					}
					mu.Unlock()
				}

				return nil
			}(idx, match)
		}
		wg.Wait()

	case true: // -f or --force full removal enabled
		var protectedPaths []string
		mu := sync.Mutex{}
		for idx, match := range d.matches {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				matchSize := d.data[idx].size

				if err := exists(path); err != nil {
					return
				}

				if rmErr := os.RemoveAll(path); rmErr != nil {
					if errors.Is(rmErr, os.ErrPermission) {
						mu.Lock()
						protectedPaths = append(protectedPaths, path)
						mu.Unlock()
						return
					}
					fmt.Println(pfmt.ApplyColor("[rmapp] ERROR: "+path+" could not be deleted", 9))
				} else if d.opts.Verbosity {
					mu.Lock()
					d.data[idx].deleted = true
					if d.data[idx].deleted {
						d.totalDeleted += matchSize
					}
					mu.Unlock()
					fmt.Printf("Successfully deleted %s 💥\n", pfmt.ApplyColor(path, 3))
				} else {
					mu.Lock()
					d.data[idx].deleted = true
					if d.data[idx].deleted {
						d.totalDeleted += matchSize
					}
					mu.Unlock()
				}
			}(match)
		}
		wg.Wait() // block till all routines have returned

		if len(protectedPaths) > 0 {
			if err := d.RunPrivilegedDelete(protectedPaths, d.opts.Verbosity); err != nil {
				return err
			}
		}
	}

	if d.totalSize != d.totalDeleted {
		fmt.Printf("Total: %s has been freed out of %s\n\n", finder.FormatSize(d.totalDeleted), finder.FormatSize(d.totalSize))
	} else {
		fmt.Printf("Total: %s has been freed\n\n", finder.FormatSize(d.totalDeleted))
	}

	return nil
}

// Checks if file/directory exists.
// Returns nil if file exists
func exists(match string) error {
	_, err := os.Stat(match)
	if err == nil {
		// File exists
		return nil
	}

	if os.IsNotExist(err) {
		fmt.Printf("File %s does not exist. Skipping...\n", pfmt.ApplyColor(match, 3))
		return err
	} else {
		fmt.Println("[rmapp] Error:", err)
		return err
	}
}

// RunPrivilegedDelete deletes a list of files/directories using AppleScript with elevated privileges.
// This is used as a fallback when SIP or permissions prevent os.RemoveAll.
func (d *Deleter) RunPrivilegedDelete(paths []string, verbose bool) error {
	mu := sync.Mutex{}
	if len(paths) == 0 {
		return nil
	}

	var pdata []data

	fmt.Println(pfmt.ApplyColor("WARN: Some files are SIP-protected. Escalating with osascript…", 3))

	var quoted []string
	for _, path := range paths {
		quoted = append(quoted, fmt.Sprintf("'%s'", path))
		pdata = append(pdata, data{match: path, size: finder.GetDiskSize(path), deleted: false})
	}
	joined := strings.Join(quoted, " ")

	//NOTE: Could add xargs call to the osascript to avoid input line length limit (extreme edge case, most likely will not occur)
	cmd := exec.Command("osascript", "-e",
		fmt.Sprintf(`do shell script "rm -rf %s" with administrator privileges`, joined))

	if err := cmd.Run(); err == nil {
		mu.Lock()
		for idx := range pdata {
			pdata[idx].deleted = true
			if pdata[idx].deleted {
				d.totalDeleted += pdata[idx].size
			}
		}
		mu.Unlock()
	} else {
		fmt.Println(pfmt.ApplyColor("[rmapp] ERROR: privileged delete failed", 9))
		if d.totalSize != d.totalDeleted {
			fmt.Printf("Total: %s has been freed out of %s\n\n", finder.FormatSize(d.totalDeleted), finder.FormatSize(d.totalSize))
		} else {
			fmt.Printf("Total: %s has been freed\n\n", finder.FormatSize(d.totalDeleted))
		}
		return err
	}

	if verbose {
		for _, path := range paths {
			fmt.Printf("Successfully deleted %s 💥\n", pfmt.ApplyColor(path, 3))
		}
	}

	return nil
}
