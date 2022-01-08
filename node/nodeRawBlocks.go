package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/node/blockAPI"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
)

// GetInternalMetaBlockByHash wil return a meta block by hash
func (n *Node) GetInternalMetaBlockByHash(format common.OutportFormat, hash string) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalMetaBlockByHash(format, decodedHash)
}

// GetInternalMetaBlockByNonce wil return a meta block by nonce
func (n *Node) GetInternalMetaBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalMetaBlockByNonce(format, nonce)
}

// GetInternalMetaBlockByRound wil return a meta block by round
func (n *Node) GetInternalMetaBlockByRound(format common.OutportFormat, round uint64) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalMetaBlockByRound(format, round)
}

// GetInternalShardBlockByHash wil return a shard block by hash
func (n *Node) GetInternalShardBlockByHash(format common.OutportFormat, hash string) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardchainOnlyEndpoint
	}

	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalShardBlockByHash(format, decodedHash)
}

// GetInternalShardBlockByNonce wil return a shard block by nonce
func (n *Node) GetInternalShardBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardchainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalShardBlockByNonce(format, nonce)
}

// GetInternalShardBlockByRound wil return a shard block by round
func (n *Node) GetInternalShardBlockByRound(format common.OutportFormat, round uint64) (interface{}, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardchainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalShardBlockByRound(format, round)
}

func (n *Node) createRawBlockProcessor() (blockAPI.APIRawBlockHandler, error) {
	statusComputer, err := txstatus.NewStatusComputer(n.processComponents.ShardCoordinator().SelfId(), n.coreComponents.Uint64ByteSliceConverter(), n.dataComponents.StorageService())
	if err != nil {
		return nil, errors.New("error creating transaction status computer " + err.Error())
	}

	blockApiArgs := &blockAPI.APIBlockProcessorArg{
		SelfShardID:              n.processComponents.ShardCoordinator().SelfId(),
		Store:                    n.dataComponents.StorageService(),
		Marshalizer:              n.coreComponents.InternalMarshalizer(),
		Uint64ByteSliceConverter: n.coreComponents.Uint64ByteSliceConverter(),
		HistoryRepo:              n.processComponents.HistoryRepository(),
		UnmarshalTx:              n.unmarshalTransaction,
		StatusComputer:           statusComputer,
	}

	return blockAPI.NewRawBlockProcessor(blockApiArgs), nil
}
