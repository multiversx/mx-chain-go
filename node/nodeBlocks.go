package node

import (
	"encoding/hex"

	apiBlock "github.com/ElrondNetwork/elrond-go/api/block"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// GetBlockByHash return the block for a given hash
func (n *Node) GetBlockByHash(hash string) (*apiBlock.APIBlock, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	if n.historyProcessor.IsEnabled() {
		return n.getBlockFromEpoch(decodedHash)
	}

	return n.getBlock(decodedHash)
}

// GetBlockByNonce returns the block for a given nonce
func (n *Node) GetBlockByNonce(nonce uint64) (*apiBlock.APIBlock, error) {
	if n.historyProcessor.IsEnabled() {
		return n.getBlockByNonceFromHistoryNonce(nonce)
	}

	return n.getBlockByNonce(nonce)
}

func (n *Node) getBlockByNonceFromHistoryNonce(nonce uint64) (*apiBlock.APIBlock, error) {
	var storerUnit dataRetriever.UnitType
	if n.shardCoordinator.SelfId() != core.MetachainShardId {
		storerUnit = dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(n.shardCoordinator.SelfId())
	} else {
		storerUnit = dataRetriever.MetaHdrNonceHashDataUnit
	}

	nonceToByteSlice := n.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := n.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	return n.getBlockFromEpoch(headerHash)
}

func (n *Node) getBlockByNonce(nonce uint64) (*apiBlock.APIBlock, error) {
	var storerUnit dataRetriever.UnitType
	if n.shardCoordinator.SelfId() != core.MetachainShardId {
		storerUnit = dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(n.shardCoordinator.SelfId())
	} else {
		storerUnit = dataRetriever.MetaHdrNonceHashDataUnit
	}

	nonceToByteSlice := n.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := n.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	return n.getBlock(headerHash)
}

func (n *Node) getBlockFromEpoch(hash []byte) (*apiBlock.APIBlock, error) {
	blockEpoch, err := n.historyProcessor.GetEpochForHash(hash)
	if err != nil {
		return nil, err
	}

	if n.shardCoordinator.SelfId() != core.MetachainShardId {
		blockStorer := n.store.GetStorer(dataRetriever.BlockHeaderUnit)

		blockBytes, err := blockStorer.GetFromEpoch(hash, blockEpoch)
		if err != nil {
			return nil, err
		}

		return n.convertShardBlockBytesToAPIBlock(hash, blockBytes)

	}

	blockStorer := n.store.GetStorer(dataRetriever.MetaBlockUnit)
	blockBytes, err := blockStorer.GetFromEpoch(hash, blockEpoch)
	if err != nil {
		return nil, err
	}

	return n.convertMetaBlockBytesToAPIBlock(hash, blockBytes)
}

func (n *Node) getBlock(hash []byte) (*apiBlock.APIBlock, error) {
	if n.shardCoordinator.SelfId() != core.MetachainShardId {
		blockStorer := n.store.GetStorer(dataRetriever.BlockHeaderUnit)

		blockBytes, err := blockStorer.Get(hash)
		if err != nil {
			return nil, err
		}

		return n.convertShardBlockBytesToAPIBlock(hash, blockBytes)

	}

	blockStorer := n.store.GetStorer(dataRetriever.MetaBlockUnit)
	blockBytes, err := blockStorer.Get(hash)
	if err != nil {
		return nil, err
	}

	return n.convertMetaBlockBytesToAPIBlock(hash, blockBytes)
}

func (n *Node) convertMetaBlockBytesToAPIBlock(hash []byte, blockBytes []byte) (*apiBlock.APIBlock, error) {
	metaBlock := &block.MetaBlock{}
	err := n.internalMarshalizer.Unmarshal(metaBlock, blockBytes)
	if err != nil {
		return nil, err
	}

	return &apiBlock.APIBlock{
		Nonce:   metaBlock.Nonce,
		Round:   metaBlock.Round,
		Epoch:   metaBlock.Epoch,
		ShardID: core.MetachainShardId,
		Hash:    hex.EncodeToString(hash),
		NumTxs:  metaBlock.TxCount,
	}, nil
}

func (n *Node) convertShardBlockBytesToAPIBlock(hash []byte, blockBytes []byte) (*apiBlock.APIBlock, error) {
	blockHeader := &block.Header{}
	err := n.internalMarshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	numOfTxs := uint32(0)
	miniBlocksHashes := make([]string, len(blockHeader.MiniBlockHeaders))
	for idx, mb := range blockHeader.MiniBlockHeaders {
		if mb.Type == block.PeerBlock {
			continue
		}
		numOfTxs += mb.TxCount
		miniBlocksHashes[idx] = hex.EncodeToString(mb.Hash)
	}

	return &apiBlock.APIBlock{
		Nonce:      blockHeader.Nonce,
		Round:      blockHeader.Round,
		Epoch:      blockHeader.Epoch,
		ShardID:    blockHeader.ShardID,
		Hash:       hex.EncodeToString(hash),
		NumTxs:     numOfTxs,
		MiniBlocks: miniBlocksHashes,
	}, nil
}
