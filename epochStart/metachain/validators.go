package metachain

import (
	"bytes"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.EpochStartValidatorInfoCreator = (*validatorInfoCreator)(nil)

// ArgsNewValidatorInfoCreator defines the arguments structure needed to create a new validatorInfo creator
type ArgsNewValidatorInfoCreator struct {
	ShardCoordinator sharding.Coordinator
	MiniBlockStorage storage.Storer
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
	DataPool         dataRetriever.PoolsHolder
}

type validatorInfoCreator struct {
	shardCoordinator sharding.Coordinator
	miniBlockStorage storage.Storer
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	dataPool         dataRetriever.PoolsHolder
}

// NewValidatorInfoCreator creates a new validatorInfo creator object
func NewValidatorInfoCreator(args ArgsNewValidatorInfoCreator) (*validatorInfoCreator, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, epochStart.ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, epochStart.ErrNilHasher
	}
	if check.IfNil(args.MiniBlockStorage) {
		return nil, epochStart.ErrNilStorage
	}
	if check.IfNil(args.DataPool) {
		return nil, epochStart.ErrNilDataPoolsHolder
	}

	vic := &validatorInfoCreator{
		shardCoordinator: args.ShardCoordinator,
		hasher:           args.Hasher,
		marshalizer:      args.Marshalizer,
		miniBlockStorage: args.MiniBlockStorage,
		dataPool:         args.DataPool,
	}

	return vic, nil
}

// CreateValidatorInfoMiniBlocks creates the validatorInfo miniblocks according to the provided validatorInfo map
func (vic *validatorInfoCreator) CreateValidatorInfoMiniBlocks(validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if validatorsInfo == nil {
		return nil, epochStart.ErrNilValidatorInfo
	}

	miniblocks := make([]*block.MiniBlock, 0)

	for shardId := uint32(0); shardId < vic.shardCoordinator.NumberOfShards(); shardId++ {
		validators := validatorsInfo[shardId]
		if len(validators) == 0 {
			continue
		}

		miniBlock, err := vic.createMiniBlock(validators)
		if err != nil {
			return nil, err
		}

		miniblocks = append(miniblocks, miniBlock)
	}

	validators := validatorsInfo[core.MetachainShardId]
	if len(validators) == 0 {
		return miniblocks, nil
	}

	miniBlock, err := vic.createMiniBlock(validators)
	if err != nil {
		return nil, err
	}

	miniblocks = append(miniblocks, miniBlock)

	return miniblocks, nil
}

func (vic *validatorInfoCreator) createMiniBlock(validatorsInfo []*state.ValidatorInfo) (*block.MiniBlock, error) {
	miniBlock := &block.MiniBlock{}
	miniBlock.SenderShardID = vic.shardCoordinator.SelfId()
	miniBlock.ReceiverShardID = core.AllShardId
	miniBlock.TxHashes = make([][]byte, len(validatorsInfo))
	miniBlock.Type = block.PeerBlock

	validatorCopy := make([]*state.ValidatorInfo, len(validatorsInfo))
	copy(validatorCopy, validatorsInfo)
	sort.Slice(validatorCopy, func(a, b int) bool {
		return bytes.Compare(validatorCopy[a].PublicKey, validatorCopy[b].PublicKey) < 0
	})

	for index, validator := range validatorCopy {
		shardValidatorInfo := createShardValidatorInfo(validator)
		marshalizedShardValidatorInfo, err := vic.marshalizer.Marshal(shardValidatorInfo)
		if err != nil {
			return nil, err
		}

		miniBlock.TxHashes[index] = marshalizedShardValidatorInfo
	}

	return miniBlock, nil
}

func createShardValidatorInfo(validator *state.ValidatorInfo) *state.ShardValidatorInfo {
	return &state.ShardValidatorInfo{
		PublicKey:  validator.PublicKey,
		ShardId:    validator.ShardId,
		List:       validator.List,
		Index:      validator.Index,
		TempRating: validator.TempRating,
	}
}

// VerifyValidatorInfoMiniBlocks verifies if received validatorinfo miniblocks are correct
func (vic *validatorInfoCreator) VerifyValidatorInfoMiniBlocks(
	miniblocks []*block.MiniBlock,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) error {
	if len(miniblocks) == 0 {
		return epochStart.ErrNilMiniblocks
	}

	createdMiniBlocks, err := vic.CreateValidatorInfoMiniBlocks(validatorsInfo)
	if err != nil {
		return err
	}

	hashesToMiniBlocks := make(map[string]*block.MiniBlock)
	for _, mb := range createdMiniBlocks {
		hash, hashError := core.CalculateHash(vic.marshalizer, vic.hasher, mb)
		if hashError != nil {
			return hashError
		}

		hashesToMiniBlocks[string(hash)] = mb
	}

	numReceivedValidatorInfoMBs := 0
	var receivedMbHash []byte
	for _, receivedMb := range miniblocks {
		if receivedMb == nil {
			return epochStart.ErrNilMiniblock
		}

		if receivedMb.Type != block.PeerBlock {
			continue
		}

		numReceivedValidatorInfoMBs++
		receivedMbHash, err = core.CalculateHash(vic.marshalizer, vic.hasher, receivedMb)
		if err != nil {
			return err
		}

		_, ok := hashesToMiniBlocks[string(receivedMbHash)]
		if !ok {
			// TODO: add display debug prints of miniblocks contents
			return epochStart.ErrValidatorMiniBlockHashDoesNotMatch
		}
	}

	if len(createdMiniBlocks) != numReceivedValidatorInfoMBs {
		return epochStart.ErrValidatorInfoMiniBlocksNumDoesNotMatch
	}

	return nil
}

// SaveValidatorInfoBlocksToStorage saves created data to storage
func (vic *validatorInfoCreator) SaveValidatorInfoBlocksToStorage(_ data.HeaderHandler, body *block.Body) {
	if check.IfNil(body) {
		return
	}

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		marshalizedData, err := vic.marshalizer.Marshal(miniBlock)
		if err != nil {
			continue
		}

		mbHash := vic.hasher.Compute(string(marshalizedData))
		_ = vic.miniBlockStorage.Put(mbHash, marshalizedData)
	}
}

// DeleteValidatorInfoBlocksFromStorage deletes data from storage
func (vic *validatorInfoCreator) DeleteValidatorInfoBlocksFromStorage(metaBlock data.HeaderHandler) {
	if check.IfNil(metaBlock) {
		return
	}

	for _, mbHeader := range metaBlock.GetMiniBlockHeaderHandlers() {
		if mbHeader.GetTypeInt32() == int32(block.PeerBlock) {
			_ = vic.miniBlockStorage.Remove(mbHeader.GetHash())
		}
	}
}

// IsInterfaceNil return true if underlying object is nil
func (vic *validatorInfoCreator) IsInterfaceNil() bool {
	return vic == nil
}

// RemoveBlockDataFromPools removes block info from pools
func (vic *validatorInfoCreator) RemoveBlockDataFromPools(metaBlock data.HeaderHandler, _ *block.Body) {
	if check.IfNil(metaBlock) {
		return
	}

	miniBlocksPool := vic.dataPool.MiniBlocks()

	for _, mbHeader := range metaBlock.GetMiniBlockHeaderHandlers() {
		if mbHeader.GetTypeInt32() != int32(block.PeerBlock) {
			continue
		}

		miniBlocksPool.Remove(mbHeader.GetHash())

		log.Trace("RemoveBlockDataFromPools",
			"hash", mbHeader.GetHash(),
			"type", mbHeader.GetTypeInt32(),
			"sender", mbHeader.GetSenderShardID(),
			"receiver", mbHeader.GetReceiverShardID(),
			"num txs", mbHeader.GetTxCount())
	}
}
