package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/update"
)

// HardForkBlockProcessor -
type HardForkBlockProcessor struct {
	CreateBlockCalled func(
		body *block.Body,
		chainID string,
		round uint64,
		nonce uint64,
		epoch uint32,
	) (data.HeaderHandler, error)
	CreateBodyCalled           func() (*block.Body, []*update.MbInfo, error)
	CreatePostMiniBlocksCalled func(mbsInfo []*update.MbInfo) (*block.Body, []*update.MbInfo, error)
}

// CreateBlock -
func (hfbp *HardForkBlockProcessor) CreateBlock(body *block.Body, chainID string, round uint64, nonce uint64, epoch uint32) (data.HeaderHandler, error) {
	if hfbp.CreateBlockCalled != nil {
		return hfbp.CreateBlockCalled(body, chainID, round, nonce, epoch)
	}
	return nil, nil
}

// CreateBody -
func (hfbp *HardForkBlockProcessor) CreateBody() (*block.Body, []*update.MbInfo, error) {
	if hfbp.CreateBodyCalled != nil {
		return hfbp.CreateBodyCalled()
	}
	return nil, nil, nil
}

// CreatePostMiniBlocks -
func (hfbp *HardForkBlockProcessor) CreatePostMiniBlocks(mbsInfo []*update.MbInfo) (*block.Body, []*update.MbInfo, error) {
	if hfbp.CreatePostMiniBlocksCalled != nil {
		return hfbp.CreatePostMiniBlocksCalled(mbsInfo)
	}
	return nil, nil, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (hfbp *HardForkBlockProcessor) IsInterfaceNil() bool {
	return hfbp == nil
}
