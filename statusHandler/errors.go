package statusHandler

import "errors"

// ErrHandlersSliceIsNil will be returned when trying to create an AppStatusFacade with no handler
var ErrHandlersSliceIsNil = errors.New("no AppStatusHandler provided")

// ErrNilHandlerInSlice will be returned when one of the handlers passed to the Facade is nil
var ErrNilHandlerInSlice = errors.New("nil AppStatusHandler")

// ErrNilPresenterInterface will be returned when a nil PresenterInterface is passed as parameter
var ErrNilPresenterInterface = errors.New("nil presenter interface")

// ErrNilGrid will be returned when a nil grid is returned
var ErrNilGrid = errors.New("nil grid")
