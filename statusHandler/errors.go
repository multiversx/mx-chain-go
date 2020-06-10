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

// ErrNilAppStatusHandler signals that a nil status handler has been provided
var ErrNilAppStatusHandler = errors.New("appStatusHandler is nil")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNilUint64Converter signals that uint64converter is nil
var ErrNilUint64Converter = errors.New("unit64converter is nil")

// ErrNilStorage signals that a nil storage has been provided
var ErrNilStorage = errors.New("nil storage")

// ErrNilTermUIStartChannel signals that a nil TermUI start channel has been provided
var ErrNilTermUIStartChannel = errors.New("nil TermUI start channel")
