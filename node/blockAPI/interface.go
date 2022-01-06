package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/api"
)

// APIBlockHandler defines the behavior of a component able to return api blocks
type APIBlockHandler interface {
	GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error)
	GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error)
	GetBlockByRound(round uint64, withTxs bool) (*api.Block, error)
}

type APIRawBlockHandler interface {
	GetRawBlockByNonce(nonce uint64, asJson bool) ([]byte, error)
	GetRawBlockByHash(hash []byte, asJson bool) ([]byte, error)
	GetRawBlockByRound(round uint64, asJson bool) ([]byte, error)
}
