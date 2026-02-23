package profiling

import "errors"

var errEmptyFolderPath = errors.New("empty folder path")
var errInvalidRoundsPerFile = errors.New("rounds per file must be at least 1")
