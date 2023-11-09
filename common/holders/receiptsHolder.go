package holders

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type receiptsHolder struct {
	miniblocks []*block.MiniBlock
}

// NewReceiptsHolder creates a receiptsHolder
func NewReceiptsHolder(miniblocks []*block.MiniBlock) *receiptsHolder {
	return &receiptsHolder{miniblocks: miniblocks}
}

// GetMiniblocks returns the contained miniblocks
func (holder *receiptsHolder) GetMiniblocks() []*block.MiniBlock {
	return holder.miniblocks
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *receiptsHolder) IsInterfaceNil() bool {
	return holder == nil
}
