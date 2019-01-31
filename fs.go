package gofs

import (
	"github.com/gofunct/common/pkg/print"
	"os"
	"path/filepath"
	"strings"
)

// FileExists determines if path exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// FindUp finds a path up the tree. On sucess, it returns found path, else "".
func FindUp(start, path string) string {
	absStart, err := filepath.Abs(start)
	if err != nil {
		return ""
	}

	filename := filepath.Join(absStart, path)
	if _, err := os.Stat(filename); err == nil {
		return filename
	}

	parent := filepath.Dir(absStart)
	if parent != absStart {
		return FindUp(parent, path)
	}
	return ""
}

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
		print.Error("common", "Outdated src error: %s", err.Error())
		return true
	}
	destFiles, _, err := Glob(destGlobs)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return true
		}
		print.Error("common", "Outdated dest error: %s", err.Error())
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
