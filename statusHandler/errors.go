package statusHandler

import "errors"

// ErrNilHandlersSlice will be returned when trying to create an AppStatusFacade with no handler
var ErrNilHandlersSlice = errors.New("no AppStatusHandler given")
