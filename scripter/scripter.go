package exec

import (
	"context"
	"encoding/json"
	"github.com/gofunct/common/pkg/exec"
	"io"
)

type ExecName int

const (
	Bash       ExecName = 0
	Docker     ExecName = 1
	Kubectl    ExecName = 3
	Terraform  ExecName = 4
	Protoc     ExecName = 5
	Stencil    ExecName = 6
	StencilBin ExecName = 7
	Gcloud     ExecName = 8
)

func (e ExecName) String() string {
	names := [...]string{
		"Bash",
		"Docker",
		"Kubectl",
		"Terraform",
		"Protoc",
		"Stencil",
		"StencilBin",
		"Gcloud",
	}

	return names[e]
}

func (e ExecName) Command() string {
	names := [...]string{
		"Bash",
		"Docker",
		"Kubectl",
		"Terraform",
		"Protoc",
		"Stencil",
		"StencilBin",
		"Gcloud",
	}

	return names[e]
}

type Script struct {
	Context context.Context
	Name    ExecName
	Args    []string
}

type Scripter struct {
	exec.Interface
	Scripts []*Script
	bits    []byte
}

func (s *Scripter) AddScript(script *Script) {
	if s.Interface == nil {
		s.Interface = exec.New()
	}
	s.Scripts = append(s.Scripts, script)
}

func (s *Scripter) AddBits(bits []byte) {
	if s.Interface == nil {
		s.Interface = exec.New()
	}
	s.bits = append(s.bits, bits...)
}

func (s *Scripter) Run() error {
	if s.Interface == nil {
		s.Interface = exec.New()
	}
	for _, v := range s.Scripts {
		cmd := s.CommandContext(v.Context, v.Name.Command(), v.Args...)
		out, err := cmd.Output()
		if err != nil {
			return err
		}
		s.bits = append(s.bits, out...)
	}
	return nil
}
func (s *Scripter) GetBits() []byte {
	err := s.jsonify()
	if err != nil {
		panic(err)
	}
	return s.bits
}

func (s *Scripter) WriteTo(w io.Writer) error {
	if s.Interface == nil {
		s.Interface = exec.New()
	}
	if err := s.jsonify(); err != nil {
		return err
	}
	_, err := w.Write(s.bits)
	if err != nil {
		return err
	}
	return nil
}
func (s *Scripter) jsonify() error {
	if s.Interface == nil {
		s.Interface = exec.New()
	}
	var err error
	s.bits, err = json.MarshalIndent(s.bits, "", "  ")
	if err != nil {
		return err
	}
	return nil
}
