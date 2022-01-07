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

type APIRawBlockHandler interface {
	GetRawShardBlockByNonce(nonce uint64) ([]byte, error)
	GetRawShardBlockByHash(hash []byte) ([]byte, error)
	GetRawShardBlockByRound(round uint64) ([]byte, error)
	GetRawMetaBlockByNonce(nonce uint64) ([]byte, error)
	GetRawMetaBlockByHash(hash []byte) ([]byte, error)
	GetRawMetaBlockByRound(round uint64) ([]byte, error)
	GetInternalShardBlockByNonce(nonce uint64) (*block.Header, error)
	GetInternalShardBlockByHash(hash []byte) (*block.Header, error)
	GetInternalShardBlockByRound(round uint64) (*block.Header, error)
	GetInternalMetaBlockByNonce(nonce uint64) (*block.MetaBlock, error)
	GetInternalMetaBlockByHash(hash []byte) (*block.MetaBlock, error)
	GetInternalMetaBlockByRound(round uint64) (*block.MetaBlock, error)
}
