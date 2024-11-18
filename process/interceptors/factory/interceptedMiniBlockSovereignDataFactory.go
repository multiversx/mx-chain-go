package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
)

type interceptedSovereignMiniBlockDataFactory struct {
	*interceptedMiniblockDataFactory
}

// NewInterceptedSovereignMiniBlockDataFactory creates an instance of intercepted sovereign mini block data factory
func NewInterceptedSovereignMiniBlockDataFactory(argument *ArgInterceptedDataFactory) (*interceptedSovereignMiniBlockDataFactory, error) {
	interceptedMbFactory, err := NewInterceptedMiniblockDataFactory(argument)
	if err != nil {
		return nil, err
	}

	return &interceptedSovereignMiniBlockDataFactory{
		interceptedMbFactory,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (imfd *interceptedSovereignMiniBlockDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := imfd.createArgsInterceptedMiniBlock(buff)
	return interceptedBlocks.NewInterceptedSovereignMiniBlock(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ismb *interceptedSovereignMiniBlockDataFactory) IsInterfaceNil() bool {
	return ismb == nil
}
