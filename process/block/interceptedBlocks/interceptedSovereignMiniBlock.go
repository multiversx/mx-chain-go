package interceptedBlocks

import "github.com/multiversx/mx-chain-core-go/core"

type interceptedSovereignMiniBlock struct {
	*InterceptedMiniblock
}

// NewInterceptedSovereignMiniBlock creates a new instance of intercepted sovereign mini block
func NewInterceptedSovereignMiniBlock(arg *ArgInterceptedMiniblock) (*interceptedSovereignMiniBlock, error) {
	interceptedMbHandler, err := NewInterceptedMiniblock(arg)
	if err != nil {
		return nil, err
	}

	return &interceptedSovereignMiniBlock{
		interceptedMbHandler,
	}, nil
}

// CheckValidity checks if the received tx block body is valid (not nil fields)
func (ismb *interceptedSovereignMiniBlock) CheckValidity() error {
	return ismb.integrity(core.MainChainShardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ismb *interceptedSovereignMiniBlock) IsInterfaceNil() bool {
	return ismb == nil
}
