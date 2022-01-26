package mock

import "github.com/ElrondNetwork/elrond-go-core/data/api"

// BlockAPIHandler -
type BlockAPIHandler struct {
}

// GetBlockByNonce -
func (bah *BlockAPIHandler) GetBlockByNonce(_ uint64, _ bool) (*api.Block, error) {
	return nil, nil
}

// GetBlockByHash -
func (bah *BlockAPIHandler) GetBlockByHash(_ []byte, _ bool) (*api.Block, error) {
	return nil, nil
}

// GetBlockByRound -
func (bah *BlockAPIHandler) GetBlockByRound(_ uint64, _ bool) (*api.Block, error) {
	return nil, nil
}

// IsInterfaceNil -
func (bah *BlockAPIHandler) IsInterfaceNil() bool {
	return bah == nil
}
