package gofs

import (
	"bytes"
	"fmt"
	"github.com/gofunct/gofs/assetfs"
	"github.com/gofunct/gofs/print"
	"github.com/mgutz/ansi"
	"os"
	"os/exec"
	"strings"
)

// Processes are the processes spawned by Start()
var Processes = make(map[string]*os.Process)

const (
	// CaptureStdout is a bitmask to capture STDOUT
	CaptureStdout = 1
	// CaptureStderr is a bitmask to capture STDERR
	CaptureStderr = 2
	// CaptureBoth captures STDOUT and STDERR
	CaptureBoth = CaptureStdout + CaptureStderr
)

type command struct {
	// original command string
	commandstr string
	// parsed executable
	executable string
	// parsed argv
	argv []string
	// parsed env
	env []string
	// working directory
	wd string
	// bitmask to capture output
	capture int
	// the output buf
	buf bytes.Buffer
}

func (gcmd *command) toExecCmd() (cmd *exec.Cmd, err error) {
	cmd = exec.Command(gcmd.executable, gcmd.argv...)
	if gcmd.wd != "" {
		cmd.Dir = gcmd.wd
	}

	cmd.Env = EffectiveEnv(gcmd.env)
	cmd.Stdin = os.Stdin

	if gcmd.capture&CaptureStderr > 0 {
		cmd.Stderr = assetfs.NewFileWrapper(os.Stderr, &gcmd.buf, ansi.Red)
	} else {
		cmd.Stderr = os.Stderr
	}
	if gcmd.capture&CaptureStdout > 0 {
		cmd.Stdout = assetfs.NewFileWrapper(os.Stdout, &gcmd.buf, "")
	} else {
		cmd.Stdout = os.Stdout
	}

	if verbose {
		if Env != "" {
			print.Debug("#", "Env: %s\n", Env)
		}
		if gcmd.wd != "" {
			print.Debug("#", "Dir: %s\n", gcmd.wd)
		}
		print.Debug("#", "%s\n", gcmd.commandstr)
	}

	return cmd, nil
}

func (gcmd *command) run() (string, error) {
	var err error
	cmd, err := gcmd.toExecCmd()
	if err != nil {
		return "", err
	}

	err = cmd.Run()
	if gcmd.capture > 0 {
		return gcmd.buf.String(), err
	}
	return "", err

}

func (gcmd *command) runAsync() error {
	cmd, err := gcmd.toExecCmd()
	if err != nil {
		return err
	}

	id := gcmd.commandstr

	// kills previously spawned process (if exists)
	killSpawned(id)
	RunnerWaitGroup.Add(1)
	waitExit = true
	go func() {
		err = cmd.Start()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		Processes[id] = cmd.Process
		if verbose {
			print.Debug("#", "Processes[%q] added\n", id)
		}
		cmd.Wait()
		RunnerWaitGroup.Done()
	}()
	return nil
}

func killSpawned(command string) {
	process := Processes[command]
	if process == nil {
		return
	}

	err := process.Kill()
	//err := syscall.Kill(-process.Pid, syscall.SIGKILL)
	delete(Processes, command)
	if err != nil && !strings.Contains(err.Error(), "process already finished") {
		print.Error("Start", "Could not kill existing process %+v\n%s\n", process, err.Error())
		return
	}
	if verbose {
		print.Debug("#", "Processes[%q] killed\n", command)
	}
}
