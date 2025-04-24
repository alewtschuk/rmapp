package finder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alewtschuk/pfmt"
	"github.com/alewtschuk/rmapp/options"
)

// Holds the match metadata
type MatchMeta struct {
	Path      string
	SizeStr   string
	PrintLine string
	Size      int64
}

func GeneratePeekReport(matches []string, appName string, opts options.Options) {
	if len(matches) == 0 {
		fmt.Printf("Found 0 files for %s\n", appName)
		return
	}

	var (
		size         int64
		totalSize    int64
		numFiles     int
		maxLineWidth int
		metas        []MatchMeta
	)

	for _, match := range matches {
		if opts.Logical {
			size = getLogicalSize(match)
		} else {
			size = GetDiskSize(match)
		}
		totalSize += size
		numFiles++

		sizeStr := FormatSize(size)
		appColored := pfmt.ApplyColor(appName, 2)
		pathColored := pfmt.ApplyColor(match, 3)

		printLine := fmt.Sprintf("• Match %s FOUND at: %s", appColored, pathColored)
		printLineStripped := fmt.Sprintf("• Match %s FOUND at: %s", appName, match)

		if len(printLineStripped) > maxLineWidth {
			maxLineWidth = len(printLineStripped)
		}

		metas = append(metas, MatchMeta{
			Path:      match,
			SizeStr:   sizeStr,
			PrintLine: printLine,
			Size:      size,
		})
	}

	// Sort the metas by size in descending order
	sort.SliceStable(metas, func(i, j int) bool {
		return metas[i].Size > metas[j].Size
	})

	fmt.Printf("\nFound %s files for %s\n", pfmt.ApplyColor(fmt.Sprintf("%d", numFiles), 3), appName)

	// Print all formatted
	for _, meta := range metas {
		lineStripped := StripColor(meta.PrintLine)
		padding := maxLineWidth - len(lineStripped)
		fmt.Printf("%s%s %s\n", meta.PrintLine, strings.Repeat(" ", padding), meta.SizeStr)
	}

	fmt.Printf("→ Total: %s would be freed\n\n", FormatSize(totalSize))
	fmt.Println("Run again without -p '--peek' to Trash files or with -f '--force' to delete files")
}
