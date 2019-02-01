package gofs

import (
	"context"
	"encoding/json"
	"github.com/gofunct/common/pkg/exec"
	"io"
)

type Script struct {
	Context context.Context
	Name    string
	Args    []string
}

type GoFs struct {
	exec.Interface
	Scripts []*Script
	bits    []byte
}

func (s *GoFs) AddScript(script *Script) {
	if s.Interface == nil {
		s.Interface = exec.New()
	}
	s.Scripts = append(s.Scripts, script)
}

func (s *GoFs) AddBits(bits []byte) {
	if s.Interface == nil {
		s.Interface = exec.New()
	}
	s.bits = append(s.bits, bits...)
}

func (s *GoFs) Run() error {
	if s.Interface == nil {
		s.Interface = exec.New()
	}
	for _, v := range s.Scripts {
		cmd := s.CommandContext(v.Context, v.Name, v.Args...)
		out, err := cmd.Output()
		if err != nil {
			return err
		}
		s.bits = append(s.bits, out...)
	}
	return nil
}
func (s *GoFs) GetBits() []byte {
	err := s.jsonify()
	if err != nil {
		panic(err)
	}
	return s.bits
}

func (s *GoFs) WriteTo(w io.Writer) error {
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
func (s *GoFs) jsonify() error {
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
