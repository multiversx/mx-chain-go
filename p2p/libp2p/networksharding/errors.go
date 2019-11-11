package networksharding

import (
	"errors"
)

var (
	ErrAlreadySet = errors.New("already set")
	ErrBadParams  = errors.New("bad Pareams")
	ErrNilSharder = errors.New("nil sharde")
)
