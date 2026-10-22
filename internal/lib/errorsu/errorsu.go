package errorsu

import (
	"errors"
	"log/slog"
)

type ErrorFunc = func() error

func Log(msg string, efs ...ErrorFunc) {
	var err error
	for _, ef := range efs {
		err = errors.Join(ef())
	}
	if err == nil {
		return
	}
	slog.Error(msg, "error", err)
}
