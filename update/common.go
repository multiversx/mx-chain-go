package update

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("update")

// MbInfo defines the structure which hold the miniBlock info
type MbInfo struct {
	MbHash          []byte
	SenderShardID   uint32
	ReceiverShardID uint32
	Type            block.Type
	TxsInfo         []*TxInfo
}

// TxInfo defines the structure which hold the transaction info
type TxInfo struct {
	TxHash []byte
	Tx     data.TransactionHandler
}

// ArgsHardForkProcessor defines the arguments structure needed by hardfork processor methods
type ArgsHardForkProcessor struct {
	Hasher                    hashing.Hasher
	Marshalizer               marshal.Marshalizer
	ShardIDs                  []uint32
	MapBodies                 map[uint32]*block.Body
	MapHardForkBlockProcessor map[uint32]HardForkBlockProcessor
	PostMbs                   []*MbInfo
}

// GetPendingMiniBlocks get all the pending miniBlocks from epoch start metaBlock and unFinished metaBlocks
func GetPendingMiniBlocks(
	epochStartMetaBlock *block.MetaBlock,
	unFinishedMetaBlocksMap map[string]*block.MetaBlock,
) ([]block.MiniBlockHeader, error) {

	if epochStartMetaBlock == nil {
		return nil, ErrNilEpochStartMetaBlock
	}
	if unFinishedMetaBlocksMap == nil {
		return nil, ErrNilUnFinishedMetaBlocksMap
	}

	pendingMiniBlocks := make([]block.MiniBlockHeader, 0)
	nonceToHashMap := createNonceToHashMap(unFinishedMetaBlocksMap)

	for _, shardData := range epochStartMetaBlock.EpochStart.LastFinalizedHeaders {
		computedPendingMiniBlocks, err := computePendingMiniBlocksFromUnFinishedMetaBlocks(
			shardData,
			unFinishedMetaBlocksMap,
			nonceToHashMap,
			epochStartMetaBlock.GetNonce(),
		)
		if err != nil {
			return nil, err
		}

		pendingMiniBlocks = append(pendingMiniBlocks, computedPendingMiniBlocks...)
	}

	return pendingMiniBlocks, nil
}

// createNonceToHashMap creates a map of nonce to hash from all the given metaBlocks
func createNonceToHashMap(unFinishedMetaBlocks map[string]*block.MetaBlock) map[uint64]string {
	nonceToHashMap := make(map[uint64]string, len(unFinishedMetaBlocks))
	for metaBlockHash, metaBlock := range unFinishedMetaBlocks {
		nonceToHashMap[metaBlock.GetNonce()] = metaBlockHash
	}

	return nonceToHashMap
}

// computePendingMiniBlocksFromUnFinishedMetaBlocks computes all the pending miniBlocks from unFinished metaBlocks
func computePendingMiniBlocksFromUnFinishedMetaBlocks(
	epochStartShardData block.EpochStartShardData,
	unFinishedMetaBlocks map[string]*block.MetaBlock,
	nonceToHashMap map[uint64]string,
	epochStartMetaBlockNonce uint64,
) ([]block.MiniBlockHeader, error) {
	pendingMiniBlocks := make([]block.MiniBlockHeader, 0)
	pendingMiniBlocks = append(pendingMiniBlocks, epochStartShardData.PendingMiniBlockHeaders...)

	firstPendingMetaBlock, ok := unFinishedMetaBlocks[string(epochStartShardData.FirstPendingMetaBlock)]
	if !ok {
		return nil, ErrWrongUnFinishedMetaHdrsMap
	}

	firstUnFinishedMetaBlockNonce := firstPendingMetaBlock.GetNonce()
	for nonce := firstUnFinishedMetaBlockNonce + 1; nonce <= epochStartMetaBlockNonce; nonce++ {
		metaBlockHash, exists := nonceToHashMap[nonce]
		if !exists {
			return nil, ErrWrongUnFinishedMetaHdrsMap
		}

		metaBlock, exists := unFinishedMetaBlocks[metaBlockHash]
		if !exists {
			return nil, ErrWrongUnFinishedMetaHdrsMap
		}

		pendingMiniBlocksFromMetaBlock := getAllMiniBlocksWithDst(metaBlock, epochStartShardData.ShardID)
		pendingMiniBlocks = append(pendingMiniBlocks, pendingMiniBlocksFromMetaBlock...)
	}

	return pendingMiniBlocks, nil
}

// getAllMiniBlocksWithDst returns all miniBlock headers with the given destination from the given metaBlock
func getAllMiniBlocksWithDst(metaBlock *block.MetaBlock, destShardID uint32) []block.MiniBlockHeader {
	mbHdrs := make([]block.MiniBlockHeader, 0)
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		if metaBlock.ShardInfo[i].ShardID == destShardID {
			continue
		}

		for _, mbHdr := range metaBlock.ShardInfo[i].ShardMiniBlockHeaders {
			if mbHdr.ReceiverShardID == destShardID && mbHdr.SenderShardID != destShardID {
				mbHdrs = append(mbHdrs, mbHdr)
			}
		}
	}

	for _, mbHdr := range metaBlock.MiniBlockHeaders {
		if mbHdr.ReceiverShardID == destShardID && mbHdr.SenderShardID != destShardID {
			mbHdrs = append(mbHdrs, mbHdr)
		}
	}

	return mbHdrs
}

// CreateBody will create a block body after hardfork import
func CreateBody(args ArgsHardForkProcessor) ([]*MbInfo, error) {
	allPostMbs := make([]*MbInfo, 0)
	for _, shardID := range args.ShardIDs {
		hardForkBlockProcessor, ok := args.MapHardForkBlockProcessor[shardID]
		if !ok {
			return nil, ErrNilHardForkBlockProcessor
		}

		body, postMbs, err := hardForkBlockProcessor.CreateBody()
		if err != nil {
			return nil, err
		}

		log.Debug("CreateBody",
			"miniBlocks in body", len(body.MiniBlocks),
			"postMbs", len(postMbs),
		)

		allPostMbs = append(allPostMbs, postMbs...)
		args.MapBodies[shardID] = body
	}

	args.PostMbs = allPostMbs
	return CleanDuplicates(args)
}

// CreatePostMiniBlocks will create all the post miniBlocks after hardfork import
func CreatePostMiniBlocks(args ArgsHardForkProcessor) error {
	var err error
	numPostMbs := len(args.PostMbs)
	for numPostMbs > 0 {
		log.Debug("CreatePostBodies", "numPostMbs", numPostMbs)
		currentPostMbs := make([]*MbInfo, 0)
		for _, shardID := range args.ShardIDs {
			hardForkBlockProcessor, ok := args.MapHardForkBlockProcessor[shardID]
			if !ok {
				return ErrNilHardForkBlockProcessor
			}

			postBody, postMbs, errCreatePostMiniBlocks := hardForkBlockProcessor.CreatePostMiniBlocks(args.PostMbs)
			if errCreatePostMiniBlocks != nil {
				return errCreatePostMiniBlocks
			}

			currentBody, ok := args.MapBodies[shardID]
			if !ok {
				return ErrNilBlockBody
			}

			log.Debug("CreatePostBodies",
				"shard", shardID,
				"miniBlocks in current body", len(currentBody.MiniBlocks),
				"miniBlocks in post body", len(postBody.MiniBlocks),
				"postMbs", len(postMbs),
			)

			currentBody.MiniBlocks = append(currentBody.MiniBlocks, postBody.MiniBlocks...)
			currentPostMbs = append(currentPostMbs, postMbs...)
			args.MapBodies[shardID] = currentBody
		}

		args.PostMbs = currentPostMbs
		args.PostMbs, err = CleanDuplicates(args)
		if err != nil {
			return err
		}

		numPostMbs = len(args.PostMbs)
	}

	return nil
}

// CleanDuplicates cleans from the post miniBlocks map, the already existing miniBlocks in bodies map
func CleanDuplicates(args ArgsHardForkProcessor) ([]*MbInfo, error) {
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}

	mapMiniBlocksHashes := make(map[string]struct{})
	for _, shardID := range args.ShardIDs {
		currentBody, ok := args.MapBodies[shardID]
		if !ok {
			return nil, ErrNilBlockBody
		}

		for _, miniBlock := range currentBody.MiniBlocks {
			miniBlockHash, err := core.CalculateHash(args.Marshalizer, args.Hasher, miniBlock)
			if err != nil {
				return nil, err
			}

			mapMiniBlocksHashes[string(miniBlockHash)] = struct{}{}
		}
	}

	cleanedPostMbs := make([]*MbInfo, 0)
	for _, postMb := range args.PostMbs {
		_, ok := mapMiniBlocksHashes[string(postMb.MbHash)]
		if ok {
			log.Debug("CleanDuplicates: found duplicated miniBlock", "hash", postMb.MbHash)
			continue
		}

		cleanedPostMbs = append(cleanedPostMbs, postMb)
	}

	return cleanedPostMbs, nil
}
