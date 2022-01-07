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
	GetRawShardBlockByNonce(nonce uint64, asJson bool) ([]byte, error)
	GetRawShardBlockByHash(hash []byte, asJson bool) ([]byte, error)
	GetRawShardBlockByRound(round uint64, asJson bool) ([]byte, error)
	GetRawMetaBlockByNonce(nonce uint64, asJson bool) ([]byte, error)
	GetRawMetaBlockByHash(hash []byte, asJson bool) ([]byte, error)
	GetRawMetaBlockByRound(round uint64, asJson bool) ([]byte, error)
	GetInternalShardBlockByNonce(nonce uint64, asJson bool) (*block.Header, error)
	GetInternalShardBlockByHash(hash []byte, asJson bool) (*block.Header, error)
	GetInternalShardBlockByRound(round uint64, asJson bool) (*block.Header, error)
	GetInternalMetaBlockByNonce(nonce uint64, asJson bool) (*block.MetaBlock, error)
	GetInternalMetaBlockByHash(hash []byte, asJson bool) (*block.MetaBlock, error)
	GetInternalMetaBlockByRound(round uint64, asJson bool) (*block.MetaBlock, error)
}
