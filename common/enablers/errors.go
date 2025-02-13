package enablers

import "errors"

var errMissingRoundActivation = errors.New("missing round activation definition")

// ErrNilEnableEpochsFactory signals that a nil enable epochs factory has been provided
var ErrNilEnableEpochsFactory = errors.New("nil enable epochs factory has been provided")
