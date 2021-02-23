package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/node/blockAPI"
)

// GetBlockByHash return the block for a given hash
func (n *Node) GetBlockByHash(hash string, withTxs bool) (*api.Block, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	apiBlockProcessor := n.createAPIBlockProcessor()
	return apiBlockProcessor.GetBlockByHash(decodedHash, withTxs)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error) {
	apiBlockProcessor := n.createAPIBlockProcessor()

	return apiBlockProcessor.GetBlockByNonce(nonce, withTxs)
}

func (n *Node) createAPIBlockProcessor() blockAPI.APIBlockHandler {
	if n.shardCoordinator.SelfId() != core.MetachainShardId {
		return blockAPI.NewShardApiBlockProcessor(
			&blockAPI.APIBlockProcessorArg{
				SelfShardID:              n.shardCoordinator.SelfId(),
				Store:                    n.store,
				Marshalizer:              n.internalMarshalizer,
				Uint64ByteSliceConverter: n.uint64ByteSliceConverter,
				HistoryRepo:              n.historyRepository,
				UnmarshalTx:              n.unmarshalTransaction,
			},
		)
	}

	return blockAPI.NewMetaApiBlockProcessor(
		&blockAPI.APIBlockProcessorArg{
			SelfShardID:              n.shardCoordinator.SelfId(),
			Store:                    n.store,
			Marshalizer:              n.internalMarshalizer,
			Uint64ByteSliceConverter: n.uint64ByteSliceConverter,
			HistoryRepo:              n.historyRepository,
			UnmarshalTx:              n.unmarshalTransaction,
		},
	)
}
