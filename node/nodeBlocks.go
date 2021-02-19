package node

import (
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/blockAPI"
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
	statusComputer, err := transaction.NewStatusComputer(n.shardCoordinator.SelfId(), n.uint64ByteSliceConverter, n.store)
	if err != nil {
		return nil, errors.New("error creating transaction status computer " + err.Error())
	}

	if n.shardCoordinator.SelfId() != core.MetachainShardId {
		return blockAPI.NewShardApiBlockProcessor(
			&blockAPI.APIBlockProcessorArg{
				SelfShardID:              n.shardCoordinator.SelfId(),
				Store:                    n.store,
				Marshalizer:              n.internalMarshalizer,
				Uint64ByteSliceConverter: n.uint64ByteSliceConverter,
				HistoryRepo:              n.historyRepository,
				UnmarshalTx:              n.unmarshalTransaction,
				StatusComputer:           statusComputer,
			},
		), nil
	}

	return blockAPI.NewMetaApiBlockProcessor(
		&blockAPI.APIBlockProcessorArg{
			SelfShardID:              n.shardCoordinator.SelfId(),
			Store:                    n.store,
			Marshalizer:              n.internalMarshalizer,
			Uint64ByteSliceConverter: n.uint64ByteSliceConverter,
			HistoryRepo:              n.historyRepository,
			UnmarshalTx:              n.unmarshalTransaction,
			StatusComputer:           statusComputer,
		},
	), nil
}
