package mock

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// MetaBlockInterceptorProcessorStub -
type MetaBlockInterceptorProcessorStub struct {
	GetEpochStartMetaBlockCalled func() (data.MetaHeaderHandler, error)
}

// Validate -
func (m *MetaBlockInterceptorProcessorStub) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save -
func (m *MetaBlockInterceptorProcessorStub) Save(_ process.InterceptedData, _ core.PeerID, _ string) error {
	return nil
}

// RegisterHandler -
func (m *MetaBlockInterceptorProcessorStub) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
}

// SignalEndOfProcessing -
func (m *MetaBlockInterceptorProcessorStub) SignalEndOfProcessing(_ []process.InterceptedData) {
}

// IsInterfaceNil -
func (m *MetaBlockInterceptorProcessorStub) IsInterfaceNil() bool {
	return m == nil
}

// GetEpochStartMetaBlock -
func (m *MetaBlockInterceptorProcessorStub) GetEpochStartMetaBlock(_ context.Context) (data.MetaHeaderHandler, error) {
	if m.GetEpochStartMetaBlockCalled != nil {
		return m.GetEpochStartMetaBlockCalled()
	}

	return &block.MetaBlock{}, nil
}
