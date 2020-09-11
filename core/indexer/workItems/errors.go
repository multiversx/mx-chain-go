package workItems

import "errors"

// ErrBodyTypeAssertion signals that we could not create an elasticsearch index
var ErrBodyTypeAssertion = errors.New("elasticsearch - body type assertion failed")
