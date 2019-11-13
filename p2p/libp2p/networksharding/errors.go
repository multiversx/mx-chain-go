package networksharding

import (
	"errors"
)

// ErrAlreadySet a sharder is already set
var ErrAlreadySet = errors.New("already set")

//ErrBadParams bad parameters
var ErrBadParams = errors.New("bad parameters")

//ErrNilSharder the sharder is nil
var ErrNilSharder = errors.New("nil sharder")
