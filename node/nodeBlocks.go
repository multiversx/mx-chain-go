package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/node/blockAPI"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
)

// GetBlockByHash return the block for a given hash
func (n *Node) GetBlockByHash(hash string, withTxs bool) (*api.Block, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := n.createAPIBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetBlockByHash(decodedHash, withTxs)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error) {
	apiBlockProcessor, err := n.createAPIBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetBlockByNonce(nonce, withTxs)
}

func (n *Node) createAPIBlockProcessor() (blockAPI.APIBlockHandler, error) {
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

	if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
		return blockAPI.NewShardApiBlockProcessor(blockApiArgs), nil
	}

	return blockAPI.NewMetaApiBlockProcessor(blockApiArgs), nil
}
