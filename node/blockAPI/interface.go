package blockAPI

import apiBlock "github.com/ElrondNetwork/elrond-go/api/block"

// APIBlockHandler defines the behavior of a component able to return api blocks
type APIBlockHandler interface {
	GetBlockByNonce(nonce uint64, withTxs bool) (*apiBlock.APIBlock, error)
	GetBlockByHash(hash []byte, withTxs bool) (*apiBlock.APIBlock, error)
}
