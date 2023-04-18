package scquery

import "errors"

var errInvalidIntervalAutoPrint = errors.New("invalid interval for auto print metrics")
var errNilLogger = errors.New("nil logger instance")
