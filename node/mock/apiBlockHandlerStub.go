package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
)

// BlockAPIHandlerStub -
type BlockAPIHandlerStub struct {
	GetBlockByNonceCalled            func(nonce uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByHashCalled             func(hash []byte, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByRoundCalled            func(round uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetAlteredAccountsForBlockCalled func(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error)
}

// GetBlockByNonce -
func (bah *BlockAPIHandlerStub) GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
	if bah.GetBlockByNonceCalled != nil {
		return bah.GetBlockByNonceCalled(nonce, options)
	}

	return nil, nil
}

// GetBlockByHash -
func (bah *BlockAPIHandlerStub) GetBlockByHash(hash []byte, options api.BlockQueryOptions) (*api.Block, error) {
	if bah.GetBlockByHashCalled != nil {
		return bah.GetBlockByHashCalled(hash, options)
	}

	return nil, nil
}

// GetBlockByRound -
func (bah *BlockAPIHandlerStub) GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
	if bah.GetBlockByRoundCalled != nil {
		return bah.GetBlockByRoundCalled(round, options)
	}

	return nil, nil
}

// GetAlteredAccountsForBlock -
func (bah *BlockAPIHandlerStub) GetAlteredAccountsForBlock(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
	if bah.GetAlteredAccountsForBlockCalled != nil {
		return bah.GetAlteredAccountsForBlockCalled(options)
	}

	return nil, nil
}

// IsInterfaceNil -
func (bah *BlockAPIHandlerStub) IsInterfaceNil() bool {
	return bah == nil
}
