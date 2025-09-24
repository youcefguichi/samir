package errs

import (
	"errors"
)

var (
	ErrFdNotFound = errors.New("fd not found")
)
