package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
)

// GetInternalMetaBlockByHash wil return a meta block by hash
func (n *Node) GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return n.internalBlockProcessor.GetInternalMetaBlockByHash(format, decodedHash)
}

// GetInternalMetaBlockByNonce wil return a meta block by nonce
func (n *Node) GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	return n.internalBlockProcessor.GetInternalMetaBlockByNonce(format, nonce)
}

// GetInternalMetaBlockByRound wil return a meta block by round
func (n *Node) GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	return n.internalBlockProcessor.GetInternalMetaBlockByRound(format, round)
}

// GetInternalShardBlockByHash wil return a shard block by hash
func (n *Node) GetInternalShardBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardOnlyEndpoint
	}

	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return n.internalBlockProcessor.GetInternalShardBlockByHash(format, decodedHash)
}

// GetInternalShardBlockByNonce wil return a shard block by nonce
func (n *Node) GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardOnlyEndpoint
	}

	return n.internalBlockProcessor.GetInternalShardBlockByNonce(format, nonce)
}

// GetInternalShardBlockByRound wil return a shard block by round
func (n *Node) GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardOnlyEndpoint
	}

	return n.internalBlockProcessor.GetInternalShardBlockByRound(format, round)
}
