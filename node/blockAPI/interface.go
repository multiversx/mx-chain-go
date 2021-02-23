package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go/data/api"
)

// APIBlockHandler defines the behavior of a component able to return api blocks
type APIBlockHandler interface {
	GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error)
	GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error)
}
