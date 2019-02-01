package gofs

import (
	"github.com/gofunct/gofs/print"
	"github.com/mgutz/str"
	"time"
)

// Dependency marks an interface as a dependency.
type Dependency interface {
	markAsDependency()
}

// Series are dependent jobs which must run in series.
type Series []interface{}

func (s Series) names() []string {
	names := []string{}
	for _, step := range s {
		switch t := step.(type) {
		case string:
			if str.SliceIndexOf(names, t) < 0 {
				names = append(names, t)
			}
		case Series:
			names = append(names, t.names()...)
		case Parallel:
			names = append(names, t.names()...)
		}

	}
	return names
}

func (s Series) markAsDependency() {}

// Parallel runs jobs in parallel
type Parallel []interface{}

func (p Parallel) names() []string {
	names := []string{}
	for _, step := range p {
		switch t := step.(type) {
		case string:
			if str.SliceIndexOf(names, t) < 0 {
				names = append(names, t)
			}
		case Series:
			names = append(names, t.names()...)
		case Parallel:
			names = append(names, t.names()...)
		}

	}
	return names
}

func (p Parallel) markAsDependency() {}

// S is alias for Series
type S []interface{}

func (s S) markAsDependency() {}

// P is alias for Parallel
type P []interface{}

func (p P) markAsDependency() {}

// Wait is minimum milliseconds before task can run again
func (j *Job) Wait(duration time.Duration) *Job {
	if duration > 0 {
		j.debounce = duration
	}
	return j
}

// Deps are task dependencies and must specify how to run jobs in series or in parallel.
func (j *Job) Deps(names ...interface{}) {
	for _, name := range names {
		switch dep := name.(type) {
		default:
			print.Error(j.Name, "Dependency types must be (string | P | Parallel | S | Series)")
		case string:
			j.dependencies = append(j.dependencies, dep)
		case P:
			j.dependencies = append(j.dependencies, Parallel(dep))
		case Parallel:
			j.dependencies = append(j.dependencies, dep)
		case S:
			j.dependencies = append(j.dependencies, Series(dep))
		case Series:
			j.dependencies = append(j.dependencies, dep)
		}
	}
}

// Description sets the description for the task.
func (j *Job) Description(desc string) *Job {
	if desc != "" {
		j.description = desc
	}
	return j
}

// Desc is alias for Description.
func (j *Job) Desc(desc string) *Job {
	return j.Description(desc)
}

// Dest adds target globs which are used to calculated outdated files.
// The jobs is not run unless ANY file Src are newer than ANY
// in DestN.
func (j *Job) Dest(globs ...string) *Job {
	if len(globs) > 0 {
		j.DestGlobs = globs
	}
	return j
}

// Src adds a source globs to this task. The task is
// not run unless files are outdated between Src and Dest globs.
func (j *Job) Src(globs ...string) *Job {
	if len(globs) > 0 {
		j.SrcGlobs = globs
	}
	return j
}
