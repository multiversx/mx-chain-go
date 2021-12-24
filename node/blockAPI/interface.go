package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// APIBlockHandler defines the behavior of a component able to return api blocks
type APIBlockHandler interface {
	GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error)
	GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error)
	GetBlockByRound(round uint64, withTxs bool) (*api.Block, error)
}

type APIRawMetaBlockHandler interface {
	GetRawBlockByNonce(nonce uint64, withTxs bool) (*block.MetaBlock, error)
	GetRawBlockByHash(hash []byte, withTxs bool) (*block.MetaBlock, error)
	GetRawBlockByRound(round uint64, withTxs bool) (*block.MetaBlock, error)
}

type APIRawHeaderHandler interface {
	GetRawBlockByNonce(nonce uint64, withTxs bool) (*block.Header, error)
	GetRawBlockByHash(hash []byte, withTxs bool) (*block.Header, error)
	GetRawBlockByRound(round uint64, withTxs bool) (*block.Header, error)
}
