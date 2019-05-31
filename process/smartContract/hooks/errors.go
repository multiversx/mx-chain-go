package hooks

import "errors"

// ErrNotImplemented signals that a functionality can not be used as it is not implemented
var ErrNotImplemented = errors.New("not implemented")

// ErrEmptyCode signals that an account does not contain code
var ErrEmptyCode = errors.New("empty code in provided smart contract holding account")
