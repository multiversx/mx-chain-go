package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/node/blockAPI"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
)

// TODO: comments update

// GetBlockByHash return the block for a given hash
func (n *Node) GetRawMetaBlockByHash(hash string, asJson bool) ([]byte, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawMetaBlockByHash(decodedHash, asJson)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetRawMetaBlockByNonce(nonce uint64, asJson bool) ([]byte, error) {
	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawMetaBlockByNonce(nonce, asJson)
}

func (n *Node) GetRawMetaBlockByRound(round uint64, asJson bool) ([]byte, error) {
	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawMetaBlockByRound(round, asJson)
}

// GetBlockByHash return the block for a given hash
func (n *Node) GetRawShardBlockByHash(hash string, asJson bool) ([]byte, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawShardBlockByHash(decodedHash, asJson)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetRawShardBlockByNonce(nonce uint64, asJson bool) ([]byte, error) {
	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawShardBlockByNonce(nonce, asJson)
}

func (n *Node) GetRawShardBlockByRound(round uint64, asJson bool) ([]byte, error) {
	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawShardBlockByRound(round, asJson)
}

// JSONNN

// GetBlockByHash return the block for a given hash
func (n *Node) GetInternalMetaBlockByHash(hash string, asJson bool) (*block.MetaBlock, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalMetaBlockByHash(decodedHash, asJson)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetInternalMetaBlockByNonce(nonce uint64, asJson bool) (*block.MetaBlock, error) {
	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalMetaBlockByNonce(nonce, asJson)
}

func (n *Node) GetInternalMetaBlockByRound(round uint64, asJson bool) (*block.MetaBlock, error) {
	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalMetaBlockByRound(round, asJson)
}

// GetBlockByHash return the block for a given hash
func (n *Node) GetInternalShardBlockByHash(hash string, asJson bool) (*block.Header, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalShardBlockByHash(decodedHash, asJson)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetInternalShardBlockByNonce(nonce uint64, asJson bool) (*block.Header, error) {
	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalShardBlockByNonce(nonce, asJson)
}

func (n *Node) GetInternalShardBlockByRound(round uint64, asJson bool) (*block.Header, error) {
	apiBlockProcessor, err := n.createRawBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetInternalShardBlockByRound(round, asJson)
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
