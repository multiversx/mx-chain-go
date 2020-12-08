package mock

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// EpochStartMetaBlockProcessorStub -
type EpochStartMetaBlockProcessorStub struct {
	ValidateCalled               func(data process.InterceptedData, fromConnectedPeer core.PeerID) error
	SaveCalled                   func(data process.InterceptedData, fromConnectedPeer core.PeerID, topic string) error
	RegisterHandlerCalled        func(handler func(topic string, hash []byte, data interface{}))
	GetEpochStartMetaBlockCalled func(ctx context.Context) (*block.MetaBlock, error)
}

// Validate -
func (esmbps *EpochStartMetaBlockProcessorStub) Validate(data process.InterceptedData, fromConnectedPeer core.PeerID) error {
	if esmbps.ValidateCalled != nil {
		return esmbps.ValidateCalled(data, fromConnectedPeer)
	}

	return nil
}

// Save -
func (esmbps *EpochStartMetaBlockProcessorStub) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, topic string) error {
	if esmbps.SaveCalled != nil {
		return esmbps.SaveCalled(data, fromConnectedPeer, topic)
	}

	return nil
}

// RegisterHandler -
func (esmbps *EpochStartMetaBlockProcessorStub) RegisterHandler(handler func(topic string, hash []byte, data interface{})) {
	if esmbps.RegisterHandlerCalled != nil {
		esmbps.RegisterHandlerCalled(handler)
	}
}

// IsInterfaceNil -
func (esmbps *EpochStartMetaBlockProcessorStub) IsInterfaceNil() bool {
	return esmbps == nil
}

// GetEpochStartMetaBlock -
func (esmbps *EpochStartMetaBlockProcessorStub) GetEpochStartMetaBlock(ctx context.Context) (*block.MetaBlock, error) {
	if esmbps.GetEpochStartMetaBlockCalled != nil {
		return esmbps.GetEpochStartMetaBlockCalled(ctx)
	}

	return nil, nil
}
