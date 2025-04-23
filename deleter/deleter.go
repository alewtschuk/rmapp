package deleter

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/alewtschuk/pfmt"
	"github.com/alewtschuk/rmapp/darwin"
	"github.com/alewtschuk/rmapp/options"
)

// Define the Deleter and its fields
type Deleter struct {
	matches []string
	opts    options.Options
}

// Creates and returns the Deleter
//
// # MODE definition:
//
// - Mode FALSE is default safe trashing
//
// - Mode TRUE is unsafe force removal
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

				success := darwin.MoveFileToTrash(match)
				if !success {
					fmt.Println(pfmt.ApplyColor("WARN: file "+match+" is sandboxed", 3))
					fmt.Println("Attempting trashing via osascript...")
					cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, match))
					err = cmd.Run()
					fmt.Println(err)
					fmt.Println(pfmt.ApplyColor("ERROR: file "+match+" unable to be moved to Trash", 9))
					err = errors.New("File trashing error")
					return err
				}

				if d.opts.Verbosity {
					fmt.Printf("Successfully moved %s to Trash 🗑️\n", pfmt.ApplyColor(match, 3))
				}

				return nil
			}()
			wg.Wait()
		}
	case true: // -f or --force full removal enabled
		for _, match := range d.matches {
			wg.Add(1)
			go func() error {
				defer wg.Done()
				err := exists(match)
				if err != nil {
					return err
				}

				err = os.RemoveAll(match)
				if err != nil {
					fmt.Println(pfmt.ApplyColor("ERROR: file "+match+" unable to be deleted", 9))
					return err
				}

				if d.opts.Verbosity {
					fmt.Printf("Successfully deleted %s 💥\n", pfmt.ApplyColor(match, 3))
				}

				return nil
			}()
		}
		wg.Wait() // block till all routines have returned
	}

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
		fmt.Println("Error:", err)
		return err
	}
}
