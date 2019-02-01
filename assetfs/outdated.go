package assetfs

import (
	"strings"
	"github.com/gofunct/gofs/print"
)

// Outdated determines if ANY src has been modified after ANY dest.
//
// For example: *.go.html -> *.go
//
// If any go.html has changed then generate go files.
func Outdated(srcGlobs, destGlobs []string) bool {
	srcFiles, _, err := Glob(srcGlobs)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return true
		}
		print.Error("outdated", "Outdated src error: %s", err.Error())
		return true
	}
	destFiles, _, err := Glob(destGlobs)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return true
		}
			print.Error("outdated", "Outdated dest error: %s", err.Error())
		return true
	}

	for _, src := range srcFiles {
		for _, dest := range destFiles {
			if src.ModTime().After(dest.ModTime()) {
				return true
			}
		}
	}
	return false
}

// TODO outdated 1-1 mapping
