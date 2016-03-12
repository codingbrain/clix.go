package args

import (
	"errors"
)

const (
	errMsgInvalidType  = "invalid type: "
	errMsgNameEmpty    = "name should not be empty"
	errMsgDupName      = "name/alias duplicated"
	errMsgNameTooShort = "name should be long name, short name comes in alias"
)

var (
	ErrArgsTooFew = errors.New("too few args")
)
