package mock

import (
	"bytes"

	"github.com/multiversx/mx-chain-go/debug"
)

// GoRoutineHandlerMap is an alias on the map of GoRoutineHandler
type GoRoutineHandlerMap = map[string]debug.GoRoutineHandler

// GoRoutineProcessorStub -
type GoRoutineProcessorStub struct {
	ProcessGoRoutineBufferCalled func(previousData map[string]GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]GoRoutineHandlerMap
}

// ProcessGoRoutineBuffer -
func (grps *GoRoutineProcessorStub) ProcessGoRoutineBuffer(previousData map[string]GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]GoRoutineHandlerMap {
	if grps.ProcessGoRoutineBufferCalled != nil {
		return grps.ProcessGoRoutineBufferCalled(previousData, buffer)
	}
	return make(map[string]GoRoutineHandlerMap)
}

// IsInterfaceNil -
func (grps *GoRoutineProcessorStub) IsInterfaceNil() bool {
	return grps == nil
}
