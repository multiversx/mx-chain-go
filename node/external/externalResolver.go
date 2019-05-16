package external

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// ExternalResolver is a struct used to gather relevant info about the blockchain status
type ExternalResolver struct {
	coordinator      sharding.Coordinator
	chainHandler     data.ChainHandler
	storage          dataRetriever.StorageService
	marshalizer      marshal.Marshalizer
	proposerResolver ProposerResolver
}

// NewExternalResolver creates a new struct responsible with the gathering of relevant blockchain info
func NewExternalResolver(
	coordinator sharding.Coordinator,
	chainHandler data.ChainHandler,
	storage dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	proposerResolver ProposerResolver,
) (*ExternalResolver, error) {

	if coordinator == nil {
		return nil, ErrNilShardCoordinator
	}
	if chainHandler == nil {
		return nil, ErrNilBlockChain
	}
	if storage == nil {
		return nil, ErrNilStore
	}
	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	if proposerResolver == nil {
		return nil, ErrNilProposerResolver
	}

	er := &ExternalResolver{
		coordinator:      coordinator,
		chainHandler:     chainHandler,
		storage:          storage,
		marshalizer:      marshalizer,
		proposerResolver: proposerResolver,
	}

	return er, nil
}

// RecentNotarizedBlocks computes last notarized [maxShardHeadersNum] shard headers (by metachain node)
// it starts from the most recent notarized metablock, fetches shard blocks and process the data from shards blocks
// After it completes the current metablock it moves to the previous metablock, process the data and so on
// It stops if it reaches the genesis block or it gathers [maxShardHeadersNum] shard headers
func (er *ExternalResolver) RecentNotarizedBlocks(maxShardHeadersNum int) ([]RecentBlock, error) {
	if er.coordinator.SelfId() != sharding.MetachainShardId {
		return nil, ErrOperationNotSupported
	}
	if maxShardHeadersNum < 1 {
		return nil, ErrInvalidValue
	}

	recentBlocks := make([]RecentBlock, 0)

	currentHeader := er.chainHandler.GetCurrentBlockHeader()
	if currentHeader == nil {
		return recentBlocks, nil
	}

	currentMetablock, ok := currentHeader.(*block.MetaBlock)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	for currentMetablock.Nonce > 1 {
		fetchedRecentBlocks, err := er.extractShardBlocksAsRecentBlocks(currentMetablock)
		if err != nil {
			return nil, err
		}
		if len(fetchedRecentBlocks)+len(recentBlocks) < maxShardHeadersNum {
			//append all
			recentBlocks = append(recentBlocks, fetchedRecentBlocks...)
		} else {
			//append remaining and return
			recentBlocks = append(recentBlocks, fetchedRecentBlocks[0:maxShardHeadersNum-len(recentBlocks)]...)
			return recentBlocks, nil
		}

		prevHash := currentMetablock.PrevHash
		currentHeaderBytes, err := er.storage.Get(dataRetriever.MetaBlockUnit, prevHash)
		if err != nil {
			return nil, err
		}
		if currentHeaderBytes == nil {
			return nil, ErrMetablockNotFoundInLocalStorage
		}

		currentMetablock = &block.MetaBlock{}
		err = er.marshalizer.Unmarshal(currentMetablock, currentHeaderBytes)
		if err != nil {
			return nil, err
		}
	}

	return recentBlocks, nil
}

func (er *ExternalResolver) extractShardBlocksAsRecentBlocks(metablock *block.MetaBlock) ([]RecentBlock, error) {
	recentBlocks := make([]RecentBlock, len(metablock.ShardInfo))

	for idx, shardInfo := range metablock.ShardInfo {
		hash := shardInfo.HeaderHash
		shardHeaderBytes, err := er.storage.Get(dataRetriever.BlockHeaderUnit, hash)
		if err != nil {
			return nil, err
		}

		shardHeader := &block.Header{}
		err = er.marshalizer.Unmarshal(shardHeader, shardHeaderBytes)
		if err != nil {
			return nil, err
		}

		proposer, err := er.proposerResolver.ResolveProposer(shardHeader.ShardId, shardHeader.Round, shardHeader.PrevRandSeed)
		if err != nil {
			return nil, err
		}

		rb := RecentBlock{
			Nonce:          shardHeader.Nonce,
			Hash:           hash,
			PrevHash:       shardHeader.PrevHash,
			StateRootHash:  shardHeader.RootHash,
			TxCount:        shardHeader.TxCount,
			BlockSize:      int64(len(shardHeaderBytes)),
			ProposerPubKey: proposer,
			PubKeysBitmap:  shardHeader.PubKeysBitmap,
			ShardID:        shardHeader.ShardId,
			Timestamp:      shardHeader.TimeStamp,
		}

		recentBlocks[idx] = rb
	}

	return recentBlocks, nil

}
