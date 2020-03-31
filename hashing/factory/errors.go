package factory

import "errors"

// ErrNoHasherInConfig signals that no hasher was provided in the config file
var ErrNoHasherInConfig = errors.New("no hasher provided in config file")
