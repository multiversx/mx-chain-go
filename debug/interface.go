package debug

import (
	"bytes"
	"time"
)

// QueryHandler defines the behavior of a queryable debug handler
type QueryHandler interface {
	Query(search string) []string
	Close() error
	IsInterfaceNil() bool
}

// GoRoutineHandlerMap represents an alias of a map of goroutineHandlers
type GoRoutineHandlerMap = map[string]GoRoutineHandler

// GoRoutinesProcessor is a component that can extract go routines from a buffer
type GoRoutinesProcessor interface {
	ProcessGoRoutineBuffer(previousData map[string]GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]GoRoutineHandlerMap
	IsInterfaceNil() bool
}

// GoRoutineHandler contains go routine information
type GoRoutineHandler interface {
	ID() string
	FirstOccurrence() time.Time
	StackTrace() string
}
