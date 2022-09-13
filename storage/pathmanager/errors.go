package pathmanager

import (
	"errors"
	"strings"
)

// ErrEmptyPruningPathTemplate signals that an empty path template for pruning storers has been provided
var ErrEmptyPruningPathTemplate = errors.New("empty path template for pruning storers")

// ErrEmptyStaticPathTemplate signals that an empty path template for static storers has been provided
var ErrEmptyStaticPathTemplate = errors.New("empty path template for static storers")

// ErrInvalidPruningPathTemplate signals that an invalid path template for pruning storers has been provided
var ErrInvalidPruningPathTemplate = errors.New("invalid path template for pruning storers")

// ErrInvalidStaticPathTemplate signals that an invalid path template for static storers has been provided
var ErrInvalidStaticPathTemplate = errors.New("invalid path template for static storers")

// ErrInvalidDatabasePath signals that an invalid database path has been provided
var ErrInvalidDatabasePath = errors.New("invalid database path")

// IsNotFoundInStorageErr returns whether an error is a "not found in storage" error.
// Currently, "item not found" storage errors are untyped (thus not distinguishable from others). E.g. see "pruningStorer.go".
// As a workaround, we test the error message for a match.
func IsNotFoundInStorageErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "not found")
}
