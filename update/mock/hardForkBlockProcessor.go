package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/update"
)

// HardForkBlockProcessor -
type HardForkBlockProcessor struct {
	CreateBlockCalled func(
		body data.BodyHandler,
		chainID string,
		round uint64,
		nonce uint64,
		epoch uint32,
	) (data.HeaderHandler, error)
	CreateBodyCalled     func() (data.BodyHandler, []*update.MbInfo, error)
	CreatePostBodyCalled func(mbsInfo []*update.MbInfo) (data.BodyHandler, []*update.MbInfo, error)
}

// CreateBlock -
func (hfbp *HardForkBlockProcessor) CreateBlock(bodyHandler data.BodyHandler, chainID string, round uint64, nonce uint64, epoch uint32) (data.HeaderHandler, error) {
	if hfbp.CreateBlockCalled != nil {
		return hfbp.CreateBlockCalled(bodyHandler, chainID, round, nonce, epoch)
	}
	return nil, nil
}

// CreateBody -
func (hfbp *HardForkBlockProcessor) CreateBody() (data.BodyHandler, []*update.MbInfo, error) {
	if hfbp.CreateBodyCalled != nil {
		return hfbp.CreateBodyCalled()
	}
	return nil, nil, nil
}

// CreatePostBody -
func (hfbp *HardForkBlockProcessor) CreatePostBody(mbsInfo []*update.MbInfo) (data.BodyHandler, []*update.MbInfo, error) {
	if hfbp.CreatePostBodyCalled != nil {
		return hfbp.CreatePostBodyCalled(mbsInfo)
	}
	return nil, nil, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (hfbp *HardForkBlockProcessor) IsInterfaceNil() bool {
	return hfbp == nil
}
