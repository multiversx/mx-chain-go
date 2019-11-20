package logs

import "errors"

// ErrWriterBusy signals that the data queue inside the writer is full
var ErrWriterBusy = errors.New("writer is busy")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilLogger signals that a nil logger has been provided
var ErrNilLogger = errors.New("nil logger")

// ErrNilWsConn signals that a nil web socket connection has been provided
var ErrNilWsConn = errors.New("nil web socket connection")

// ErrWriterClosed signals that the current writer is closed and no longer accepts writes
var ErrWriterClosed = errors.New("writer is closed")
