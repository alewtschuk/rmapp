package deleter

import (
	"errors"
	"fmt"
	"log"
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
	matches []string
	opts    options.Options
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

	var totalSize int64
	var isSudo bool
	sudoUser := os.Getenv("SUDO_USER") //check what user called sudo if any

	// Set if user is sudo or not
	if sudoUser == "" {
		isSudo = false
	} else {
		isSudo = true
	}

	for _, match := range d.matches {
		size := finder.GetDiskSize(match)
		totalSize += size
	}

	switch d.opts.Mode {
	case false: // default trashing behavior
		for _, match := range d.matches {
			wg.Add(1)
			go func() error {
				defer wg.Done()
				err := exists(match)
				if err != nil {
					return err
				}

				if isSudo {
					appleScript := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, match) //setup applescript string
					cmd := exec.Command("sudo", "-u", sudoUser, "osascript", "-e", appleScript)
					err := cmd.Run()
					if err != nil {
						fmt.Println(pfmt.ApplyColor(fmt.Sprintf("[rmapp] ERROR: privileged trash for %s failed: %v", match, err), 9))
						// Even if this fails, we should continue with other files.
						return nil
					}

					log.Printf("Successfully moved %s to Trash 🗑️\n", pfmt.ApplyColor(match, 3))

				} else {
					success := darwin.MoveFileToTrash(match)
					if !success {
						fmt.Println(pfmt.ApplyColor("WARN: file "+match+" requires elevated permissions to remove", 3))
						fmt.Println("Attempting trashing via osascript...")
						cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, match))
						err = cmd.Run()
						if err != nil {
							fmt.Println(pfmt.ApplyColor("[rmapp] ERROR: file "+match+" unable to be moved to Trash", 9))
						}

						log.Printf("Successfully moved %s to Trash 🗑️\n", pfmt.ApplyColor(match, 3))

						return err
					}

					log.Printf("Successfully moved %s to Trash 🗑️\n", pfmt.ApplyColor(match, 3))

				}

				return nil
			}()
		}
		wg.Wait()

	case true: // -f or --force full removal enabled
		var protectedPaths []string
		mu := sync.Mutex{}

		for _, match := range d.matches {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()

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
				}

				log.Printf("Successfully deleted %s 💥\n", pfmt.ApplyColor(path, 3))

			}(match)
		}
		wg.Wait() // block till all routines have returned

		if len(protectedPaths) > 0 {
			if err := RunPrivilegedDelete(protectedPaths, d.opts.Verbosity); err != nil {
				return err
			}
		}
	}

	fmt.Printf("Total: %s has been freed\n\n", finder.FormatSize(totalSize))

	return nil
}

// Checks if file/directory exists
func exists(match string) error {
	_, err := os.Stat(match) // explicitly used for error only
	if err == nil {
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
// This is used as a fallback when permissions prevent os.RemoveAll.
func RunPrivilegedDelete(paths []string, verbose bool) error {
	if len(paths) == 0 {
		return nil
	}

	fmt.Println(pfmt.ApplyColor("WARN: Some files are permission protected. Escalating with osascript…", 3))

	var quoted []string
	for _, path := range paths {
		quoted = append(quoted, fmt.Sprintf("'%s'", path))
	}
	joined := strings.Join(quoted, " ")

	cmd := exec.Command("osascript", "-e",
		fmt.Sprintf(`do shell script "rm -rf %s" with administrator privileges`, joined))

	if err := cmd.Run(); err != nil {
		fmt.Println(pfmt.ApplyColor("[rmapp] ERROR: privileged delete failed", 9))
		return err
	}

	if verbose {
		for _, path := range paths {
			fmt.Printf("Successfully deleted %s 💥\n", pfmt.ApplyColor(path, 3))
		}
	}

	return nil
}
