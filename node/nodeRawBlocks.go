package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/node/blockAPI"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
)

// TODO: comments update

// GetBlockByHash return the block for a given hash
func (n *Node) GetRawMetaBlockByHash(hash string) ([]byte, error) {
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

	return apiBlockProcessor.GetRawMetaBlockByHash(decodedHash)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetRawMetaBlockByNonce(nonce uint64) ([]byte, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawMetaBlockByNonce(nonce)
}

func (n *Node) GetRawMetaBlockByRound(round uint64) ([]byte, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawMetaBlockByRound(round)
}

// GetBlockByHash return the block for a given hash
func (n *Node) GetRawShardBlockByHash(hash string) ([]byte, error) {
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

	return apiBlockProcessor.GetRawShardBlockByHash(decodedHash)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetRawShardBlockByNonce(nonce uint64) ([]byte, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardchainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawShardBlockByNonce(nonce)
}

func (n *Node) GetRawShardBlockByRound(round uint64) ([]byte, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardchainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawShardBlockByRound(round)
}

// GetBlockByHash return the block for a given hash
func (n *Node) GetInternalMetaBlockByHash(hash string) (*block.MetaBlock, error) {
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

	return apiBlockProcessor.GetInternalMetaBlockByHash(decodedHash)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetInternalMetaBlockByNonce(nonce uint64) (*block.MetaBlock, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalMetaBlockByNonce(nonce)
}

func (n *Node) GetInternalMetaBlockByRound(round uint64) (*block.MetaBlock, error) {
	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalMetaBlockByRound(round)
}

// GetBlockByHash return the block for a given hash
func (n *Node) GetInternalShardBlockByHash(hash string) (*block.Header, error) {
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

	return apiBlockProcessor.GetInternalShardBlockByHash(decodedHash)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetInternalShardBlockByNonce(nonce uint64) (*block.Header, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardchainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalShardBlockByNonce(nonce)
}

func (n *Node) GetInternalShardBlockByRound(round uint64) (*block.Header, error) {
	if n.processComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		return nil, ErrShardchainOnlyEndpoint
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalShardBlockByRound(round)
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
