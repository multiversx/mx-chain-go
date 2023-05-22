package cutoff

import "errors"

var errProcess = errors.New("block processing cutoff - intended processing error")

var errInvalidBlockProcessingCutOffMode = errors.New("invalid block processing cutoff mode")

var errInvalidBlockProcessingCutOffTrigger = errors.New("invalid block processing cutoff trigger")
