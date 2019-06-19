package external

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
func (er *ExternalResolver) RecentNotarizedBlocks(maxShardHeadersNum int) ([]*BlockHeader, error) {
	if er.coordinator.SelfId() != sharding.MetachainShardId {
		return nil, ErrOperationNotSupported
	}
	if maxShardHeadersNum < 1 {
		return nil, ErrInvalidValue
	}

	recentBlocks := make([]*BlockHeader, 0)

	currentHeader := er.chainHandler.GetCurrentBlockHeader()
	if currentHeader == nil {
		return recentBlocks, nil
	}

	currentMetablock, ok := currentHeader.(*block.MetaBlock)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	for currentMetablock.Nonce > 0 {
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

func (er *ExternalResolver) extractShardBlocksAsRecentBlocks(metablock *block.MetaBlock) ([]*BlockHeader, error) {
	recentBlocks := make([]*BlockHeader, len(metablock.ShardInfo))

	for idx, shardInfo := range metablock.ShardInfo {
		hash := shardInfo.HeaderHash
		shb, err := er.retrieveShardBlockHeader(hash)
		if err != nil {
			return nil, err
		}

		recentBlocks[idx] = shb
	}

	return recentBlocks, nil

}

func (er *ExternalResolver) retrieveShardBlockHeader(headerHash []byte) (*BlockHeader, error) {
	shardHeaderBytes, err := er.storage.Get(dataRetriever.BlockHeaderUnit, headerHash)
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

	shb := &BlockHeader{
		Nonce:          shardHeader.Nonce,
		Hash:           headerHash,
		PrevHash:       shardHeader.PrevHash,
		StateRootHash:  shardHeader.RootHash,
		TxCount:        shardHeader.TxCount,
		BlockSize:      int64(len(shardHeaderBytes)),
		ProposerPubKey: proposer,
		PubKeysBitmap:  shardHeader.PubKeysBitmap,
		ShardId:        shardHeader.ShardId,
		TimeStamp:      shardHeader.TimeStamp,
	}

	return shb, nil
}

// RetrieveShardBlock retrieves a shard block info containing header and transactions
func (er *ExternalResolver) RetrieveShardBlock(blockHash []byte) (*ShardBlockInfo, error) {
	if blockHash == nil {
		return nil, ErrNilHash
	}

	shb, err := er.retrieveShardBlockHeader(blockHash)
	if err != nil {
		return nil, err
	}

	sbi := &ShardBlockInfo{
		BlockHeader: *shb,
	}

	//TODO, implement transaction fetching
	return sbi, nil
}
