package consensus

import "errors"

// ErrWrongTypeAssertion signals that an type assertion failed
var ErrWrongTypeAssertion = errors.New("wrong type assertion")
