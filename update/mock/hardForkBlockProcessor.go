package mock

import "github.com/ElrondNetwork/elrond-go/data"

// HardForkBlockProcessor --
type HardForkBlockProcessor struct {
	CreateNewBlockCalled func(
		chainID string,
		round uint64,
		nonce uint64,
		epoch uint32,
	) (data.HeaderHandler, data.BodyHandler, error)
}

// CreateNewBlock --
func (hfbp *HardForkBlockProcessor) CreateNewBlock(chainID string, round uint64, nonce uint64, epoch uint32) (data.HeaderHandler, data.BodyHandler, error) {
	if hfbp.CreateNewBlockCalled != nil {
		return hfbp.CreateNewBlockCalled(chainID, round, nonce, epoch)
	}
	return nil, nil, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (hfbp *HardForkBlockProcessor) IsInterfaceNil() bool {
	return hfbp == nil
}
