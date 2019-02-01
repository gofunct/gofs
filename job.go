package gofs

import (
	"fmt"
	"github.com/gofunct/gofs/assetfs"
	"github.com/gofunct/gofs/print"
	"github.com/mgutz/minimist"
	"github.com/mgutz/str"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TaskFunction is the signature of the function used to define a type.
// type TaskFunc func(string, ...interface{}) *Task
// type UseFunc func(string, interface{})

// A Task is an operation performed on a user's project directory.
type Job struct {
	Name         string
	description  string
	Handler      Handler
	dependencies Series
	argm         minimist.ArgMap

	// Watches are watches files. On change the task is rerun. For example `**/*.less`
	// Usually Watches and Sources are the same.
	// WatchFiles   []*FileAsset
	// WatchGlobs   []string
	// WatchRegexps []*RegexpInfo

	// computed based on dependencies
	EffectiveWatchRegexps []*assetfs.RegexpInfo
	EffectiveWatchGlobs   []string

	// Complete indicates whether this task has already ran. This flag is
	// ignored in watch mode.
	Complete bool
	debounce time.Duration
	RunOnce  bool

	SrcFiles   []*assetfs.FileAsset
	SrcGlobs   []string
	SrcRegexps []*assetfs.RegexpInfo

	DestFiles   []*assetfs.FileAsset
	DestGlobs   []string
	DestRegexps []*assetfs.RegexpInfo

	// used when a file event is received between debounce intervals, the file event
	// will queue itself and set this flag and force debounce to run it
	// when time has elapsed
	sync.Mutex
	ignoreEvents bool
}

// NewTask creates a new Task.
func NewJob(name string, argm minimist.ArgMap) *Job {
	runOnce := false
	if strings.HasSuffix(name, "?") {
		runOnce = true
		name = str.ChompRight(name, "?")
	}
	return &Job{Name: name, RunOnce: runOnce, dependencies: Series{}, argm: argm}
}

// Expands assetfs patterns.
func (j *Job) expandGlobs() {

	// runs once lazily
	if len(j.SrcFiles) > 0 {
		return
	}

	files, regexps, err := assetfs.Glob(j.SrcGlobs)
	if err != nil {
		print.Error(j.Name, "%v", err)
		return
	}

	j.SrcRegexps = regexps
	j.SrcFiles = files

	if len(j.DestGlobs) > 0 {
		files, regexps, err := assetfs.Glob(j.DestGlobs)
		if err != nil {
			print.Error(j.Name, "%v", err)
			return
		}
		j.DestRegexps = regexps
		j.DestFiles = files
	}
}

// Run runs all the dependencies of this job and when they have completed,
// runs this job.
func (j *Job) Run() error {
	if !watching && j.Complete {
		print.Debug(j.Name, "Already ran\n")
		return nil
	}
	return j.RunWithEvent(j.Name, nil)
}

// isWatchedFile determines if a FileEvent's file is a watched file
func (j *Job) isWatchedFile(path string) bool {
	filename, err := filepath.Rel(wd, path)
	if err != nil {
		return false
	}

	filename = filepath.ToSlash(filename)
	//print.Debug("task", "checking for match %s\n", filename)

	matched := false
	for _, info := range j.EffectiveWatchRegexps {
		if info.Negate {
			if matched {
				matched = !info.MatchString(filename)
				//print.Debug("task", "negated match? %s %s\n", filename, matched)
				continue
			}
		} else if info.MatchString(filename) {
			matched = true
			//print.Debug("task", "matched %s %s\n", filename, matched)
			continue
		}
	}
	return matched
}

// RunWithEvent runs this task when triggered from a watch.
// *e* FileEvent contains information about the file/directory which changed
// in watch mode.
func (j *Job) RunWithEvent(logName string, e *assetfs.FileEvent) (err error) {
	if j.RunOnce && j.Complete {
		print.Debug(j.Name, "Already ran\n")
		return nil
	}

	j.expandGlobs()
	if !j.shouldRun(e) {
		print.Info(logName, "up-to-date 0ms\n")
		return nil
	}

	start := time.Now()
	if len(j.SrcGlobs) > 0 && len(j.SrcFiles) == 0 {
		print.Error("task", "\""+j.Name+"\" '%v' did not match any files\n", j.SrcGlobs)
	}

	// Run this task only if the file matches watch Regexps
	rebuilt := ""
	if e != nil {
		rebuilt = "rebuilt "
		if !j.isWatchedFile(e.Path) && len(j.SrcGlobs) > 0 {
			return nil
		}
		if verbose {
			print.Debug(logName, "%s\n", e.String())
		}
	}

	log := true
	if j.Handler != nil {
		context := Context{Job: j, Args: j.argm, FileEvent: e}
		defer func() {
			if p := recover(); p != nil {
				sp, ok := p.(*softPanic)
				if !ok {
					panic(p)
				}
				err = fmt.Errorf("%q: %s", logName, sp)
			}
		}()

		j.Handler.Handle(&context)
		if context.Error != nil {
			return fmt.Errorf("%q: %s", logName, context.Error.Error())
		}
	} else if len(j.dependencies) > 0 {
		// no need to log if just dependency
		log = false
	} else {
		print.Info(j.Name, "Ignored. Task does not have a handler or dependencies.\n")
		return nil
	}

	if log {
		if rebuilt != "" {
			print.InfoColorful(logName, "%s%vms\n", rebuilt, time.Since(start).Nanoseconds()/1e6)
		} else {
			print.Info(logName, "%s%vms\n", rebuilt, time.Since(start).Nanoseconds()/1e6)
		}
	}

	j.Complete = true

	return nil
}

// DependencyNames gets the flattened dependency names.
func (j *Job) DependencyNames() []string {
	if len(j.dependencies) == 0 {
		return nil
	}
	deps := []string{}
	for _, dep := range j.dependencies {
		switch d := dep.(type) {
		default:
			panic("dependencies can only be Serial or Parallel")
		case Series:
			deps = append(deps, d.names()...)
		case Parallel:
			deps = append(deps, d.names()...)
		case S:
			deps = append(deps, Series(d).names()...)
		case P:
			deps = append(deps, Parallel(d).names()...)
		}
	}
	return deps
}

func (j *Job) dump(buf io.Writer, indent string) {
	fmt.Fprintln(buf, indent, j.Name)
	fmt.Fprintln(buf, indent+indent, "EffectiveWatchGlobs", j.EffectiveWatchGlobs)
	fmt.Fprintln(buf, indent+indent, "SrcFiles", j.SrcFiles)
	fmt.Fprintln(buf, indent+indent, "SrcGlobs", j.SrcGlobs)

}

func (j *Job) shouldRun(e *assetfs.FileEvent) bool {
	if e == nil || len(j.SrcFiles) == 0 {
		return true
	} else if !j.isWatchedFile(e.Path) {
		// fmt.Printf("received a file so it should return immediately\n")
		return false
	}

	// lazily expand assetfss
	j.expandGlobs()

	if len(j.SrcFiles) == 0 || len(j.DestFiles) == 0 {
		// fmt.Printf("no source files %s %#v\n", j.Name, j.SrcFiles)
		// fmt.Printf("no source files %s %#v\n", j.Name, j.DestFiles)
		return true
	}

	// TODO figure out intelligent way to cache this instead of stating
	// each time
	for _, src := range j.SrcFiles {
		// refresh stat
		src.Stat()
		for _, dest := range j.DestFiles {
			// refresh stat
			dest.Stat()
			if filepath.Base(src.Path) == "foo.txt" {
				fmt.Printf("src %s %#v\n", src.Path, src.ModTime().UnixNano())
				fmt.Printf("dest %s %#v\n", dest.Path, dest.ModTime().UnixNano())
			}
			if src.ModTime().After(dest.ModTime()) {
				return true
			}
		}
	}

	fmt.Printf("FileEvent ignored %#v\n", e)

	return false
}

func (j *Job) debounceValue() time.Duration {
	if j.debounce == 0 {
		// use default Wait
		return Wait
	}
	return j.debounce
}
