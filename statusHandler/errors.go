package statusHandler

import "errors"

// ErrHandlersSliceIsNil will be returned when trying to create an AppStatusFacade with no handler
var ErrHandlersSliceIsNil = errors.New("no AppStatusHandler provided")

// ErrNilHandlerInSlice will be returned when one of the handlers passed to the Facade is nil
var ErrNilHandlerInSlice = errors.New("nil AppStatusHandler")
