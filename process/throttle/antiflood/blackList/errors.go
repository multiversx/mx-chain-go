package blackList

import "errors"

// ErrEmptyConfigVarNameForNumFloodingRounds signals that an empty config variable name for num flooding rounds has been provided
var ErrEmptyConfigVarNameForNumFloodingRounds = errors.New("empty config variable name for num flooding rounds")
