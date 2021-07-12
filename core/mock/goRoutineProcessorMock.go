package mock

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
)

type GoRoutineHandlerMap = map[string]core.GoRoutineHandler

type GoRoutineProcessorMock struct {
	ProcessGoRoutineBufferCalled func(previousData map[string]GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]GoRoutineHandlerMap
}

func (grpm *GoRoutineProcessorMock) ProcessGoRoutineBuffer(previousData map[string]GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]GoRoutineHandlerMap {
	if grpm.ProcessGoRoutineBufferCalled != nil {
		return grpm.ProcessGoRoutineBufferCalled(previousData, buffer)
	}
	return make(map[string]GoRoutineHandlerMap)
}

func (grpm *GoRoutineProcessorMock) IsInterfaceNil() bool {
	return grpm == nil
}
