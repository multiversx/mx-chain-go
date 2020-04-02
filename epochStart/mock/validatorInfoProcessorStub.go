package mock

import (
	data2 "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// ValidatorInfoProcessorStub -
type ValidatorInfoProcessorStub struct {
}

// SyncPeerMiniBlocks -
func (vip *ValidatorInfoProcessorStub) SyncPeerMiniBlocks(*block.MetaBlock) ([][]byte, data2.BodyHandler, error) {
	return nil, nil, nil
}

// IsInterfaceNil
func (vip *ValidatorInfoProcessorStub) IsInterfaceNil() bool {
	return vip == nil
}
