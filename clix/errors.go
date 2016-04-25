package clix

import (
	"errors"
)

var (
	ErrorTypeNotSupported = errors.New("Type not supported")

	errorNotStarted = errors.New("Not started")
)
