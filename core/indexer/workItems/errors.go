package workItems

import "errors"

// ErrBodyTypeAssertion signals that body type assertion failed
var ErrBodyTypeAssertion = errors.New("elasticsearch - body type assertion failed")
