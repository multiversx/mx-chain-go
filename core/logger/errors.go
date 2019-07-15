package logger

import "errors"

// ErrNilFile signals that the provided file is nil
var ErrNilFile = errors.New("can not use nil file")
