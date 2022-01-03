package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go/node/blockAPI"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
)

// GetBlockByHash return the block for a given hash
func (n *Node) GetRawBlockByHash(hash string, withTxs bool) ([]byte, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor, err := n.createAPIMetaBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawBlockByHash(decodedHash, withTxs)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetRawBlockByNonce(nonce uint64, withTxs bool) ([]byte, error) {
	apiBlockProcessor, err := n.createAPIMetaBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawBlockByNonce(nonce, withTxs)
}

func (n *Node) GetRawBlockByRound(round uint64, withTxs bool) ([]byte, error) {
	apiBlockProcessor, err := n.createAPIMetaBlockProcessor()
	if err != nil {
		return nil, err
	}

	return apiBlockProcessor.GetRawBlockByRound(round, withTxs)
}

func (n *Node) createAPIMetaBlockProcessor() (blockAPI.APIRawMetaBlockHandler, error) {
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

	// if n.processComponents.ShardCoordinator().SelfId() != core.MetachainShardId {
	// 	return blockAPI.NewShardApiBlockProcessor(blockApiArgs), nil
	// }

	return blockAPI.NewMetaApiBlockProcessor(blockApiArgs), nil
}
