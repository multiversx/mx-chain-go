package mock

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// MetaBlockInterceptorProcessorStub -
type MetaBlockInterceptorProcessorStub struct {
	GetEpochStartMetaBlockCalled func() (*block.MetaBlock, error)
}

// Validate -
func (m *MetaBlockInterceptorProcessorStub) Validate(data process.InterceptedData, _ p2p.PeerID) error {
	return nil
}

// Save -
func (m *MetaBlockInterceptorProcessorStub) Save(_ process.InterceptedData, _ p2p.PeerID) error {
	return nil
}

// RegisterHandler -
func (m *MetaBlockInterceptorProcessorStub) RegisterHandler(_ func(toShard uint32, data []byte)) {
}

// SignalEndOfProcessing -
func (m *MetaBlockInterceptorProcessorStub) SignalEndOfProcessing(_ []process.InterceptedData) {
}

// IsInterfaceNil -
func (m *MetaBlockInterceptorProcessorStub) IsInterfaceNil() bool {
	return m == nil
}

// GetEpochStartMetaBlock -
func (m *MetaBlockInterceptorProcessorStub) GetEpochStartMetaBlock(_ context.Context) (*block.MetaBlock, error) {
	if m.GetEpochStartMetaBlockCalled != nil {
		return m.GetEpochStartMetaBlockCalled()
	}

	return &block.MetaBlock{}, nil
}
