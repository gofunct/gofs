package zap

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(L)
}

var L, _ = zap.NewDevelopment()

// LogF logs fatal "msg: err" in case of error
func LogF(msg string, err error) {
	if err != nil {
		L.Fatal(msg, zap.Error(err))
	}
}

// LogE logs error "msg: err" in case of error
func LogE(msg string, err error) {
	if err != nil {
		L.Info(msg, zap.Error(err))
	}
}

// LogE logs error "msg: err" in case of error
func LogW(msg string, err error) {
	if err != nil {
		L.Warn(msg, zap.Error(err))
	}
}

// LogE logs error "msg: err" in case of error
func With(k, v string) {
	L.With(zap.String(k, v))
}

// LogE logs error "msg: err" in case of error
func Debug(msg string, k, v string) {
	L.Debug(msg, zap.String(k, v))
}

func Close() func() error {
	return func() error {
		return L.Sync()
	}
}

func StackF(msg string, err error) {
	if err != nil {
		L.Fatal(msg, zap.Error(errors.WithStack(err)))
	}
}
