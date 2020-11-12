package mock

import (
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

type StatusHandlersFactoryMock struct {
}

func (shfm *StatusHandlersFactoryMock) Create(marshalizer marshal.Marshalizer, converter typeConverters.Uint64ByteSliceConverter) (factory.StatusHandlersUtils, error) {
	return &StatusHandlersUtilsMock{
		AppStatusHandler: &AppStatusHandlerMock{},
	}, nil
}
