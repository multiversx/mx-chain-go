package mock

import (
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

// StatusHandlersFactoryMock -
type StatusHandlersFactoryMock struct {
}

// Create -
func (shfm *StatusHandlersFactoryMock) Create(_ marshal.Marshalizer, _ typeConverters.Uint64ByteSliceConverter) (factory.StatusHandlersUtils, error) {
	return &StatusHandlersUtilsMock{
		AppStatusHandler: &AppStatusHandlerMock{},
	}, nil
}
