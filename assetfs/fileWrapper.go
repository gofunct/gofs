package assetfs

import (
	"bytes"
	"fmt"
	"os"

	"github.com/mgutz/ansi"
)

type FileWrapper struct {
	file      *os.File
	buf       *bytes.Buffer
	readLines string

	recorder *bytes.Buffer

	// Adds color to stdout & stderr if terminal supports it
	colorStart string
}

func NewFileWrapper(file *os.File, recorder *bytes.Buffer, color string) *FileWrapper {
	streamer := &FileWrapper{
		file:       file,
		buf:        bytes.NewBufferString(""),
		recorder:   recorder,
		colorStart: color,
	}

	return streamer
}

func (l *FileWrapper) Write(p []byte) (n int, err error) {
	if n, err = l.recorder.Write(p); err != nil {
		return
	}

	err = l.out(string(p))
	return
}

func (l *FileWrapper) WriteString(s string) (n int, err error) {
	if n, err = l.recorder.WriteString(s); err != nil {
		return
	}

	err = l.out(s)
	return
}

func (l *FileWrapper) Close() error {
	l.buf = bytes.NewBuffer([]byte(""))
	return nil
}

func (l *FileWrapper) out(str string) (err error) {

	if l.colorStart != "" {
		fmt.Fprint(l.file, l.colorStart)
		fmt.Fprint(l.file, str)
		fmt.Fprint(l.file, ansi.Reset)
	} else {
		fmt.Fprint(l.file, str)
	}

	return nil
}
