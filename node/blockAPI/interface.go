package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
)

// APIBlockHandler defines the behavior of a component able to return api blocks
type APIBlockHandler interface {
	GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error)
	GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error)
	GetBlockByRound(round uint64, withTxs bool) (*api.Block, error)
}

// APIInternalBlockHandler defines the behaviour of a component able to return internal blocks
type APIInternalBlockHandler interface {
	GetInternalShardBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHash(format common.OutportFormat, hash []byte) (interface{}, error)
	GetInternalShardBlockByRound(format common.OutportFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHash(format common.OutportFormat, hash []byte) (interface{}, error)
	GetInternalMetaBlockByRound(format common.OutportFormat, round uint64) (interface{}, error)
}
