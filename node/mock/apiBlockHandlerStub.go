package mock

import "github.com/ElrondNetwork/elrond-go-core/data/api"

// BlockAPIHandlerStub -
type BlockAPIHandlerStub struct {
	GetBlockByNonceCalled func(nonce uint64, withTxs bool) (*api.Block, error)
	GetBlockByHashCalled  func(hash []byte, withTxs bool) (*api.Block, error)
	GetBlockByRoundCalled func(round uint64, withTxs bool) (*api.Block, error)
}

// GetBlockByNonce -
func (bah *BlockAPIHandlerStub) GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error) {
	if bah.GetBlockByNonceCalled != nil {
		return bah.GetBlockByNonceCalled(nonce, withTxs)
	}

	return nil, nil
}

// GetBlockByHash -
func (bah *BlockAPIHandlerStub) GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error) {
	if bah.GetBlockByHashCalled != nil {
		return bah.GetBlockByHashCalled(hash, withTxs)
	}

	return nil, nil
}

// GetBlockByRound -
func (bah *BlockAPIHandlerStub) GetBlockByRound(round uint64, withTxs bool) (*api.Block, error) {
	if bah.GetBlockByRoundCalled != nil {
		return bah.GetBlockByRoundCalled(round, withTxs)
	}

	return nil, nil
}

// IsInterfaceNil -
func (bah *BlockAPIHandlerStub) IsInterfaceNil() bool {
	return bah == nil
}
