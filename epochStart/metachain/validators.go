package metachain

import (
	"bytes"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.EpochStartValidatorInfoCreator = (*validatorInfoCreator)(nil)

// ArgsNewValidatorInfoCreator defines the arguments structure needed to create a new validatorInfo creator
type ArgsNewValidatorInfoCreator struct {
	ShardCoordinator     sharding.Coordinator
	ValidatorInfoStorage storage.Storer
	MiniBlockStorage     storage.Storer
	Hasher               hashing.Hasher
	Marshalizer          marshal.Marshalizer
	DataPool             dataRetriever.PoolsHolder
	EnableEpochsHandler  common.EnableEpochsHandler
}

type validatorInfoCreator struct {
	shardCoordinator     sharding.Coordinator
	validatorInfoStorage storage.Storer
	miniBlockStorage     storage.Storer
	hasher               hashing.Hasher
	marshalizer          marshal.Marshalizer
	dataPool             dataRetriever.PoolsHolder
	mutValidatorInfo     sync.Mutex
	enableEpochsHandler  common.EnableEpochsHandler
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
	if check.IfNil(args.ValidatorInfoStorage) {
		return nil, epochStart.ErrNilValidatorInfoStorage
	}
	if check.IfNil(args.MiniBlockStorage) {
		return nil, epochStart.ErrNilStorage
	}
	if check.IfNil(args.DataPool) {
		return nil, epochStart.ErrNilDataPoolsHolder
	}
	if check.IfNil(args.DataPool.CurrentEpochValidatorInfo()) {
		return nil, epochStart.ErrNilCurrentEpochValidatorsInfoPool
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, epochStart.ErrNilEnableEpochsHandler
	}

	vic := &validatorInfoCreator{
		shardCoordinator:     args.ShardCoordinator,
		hasher:               args.Hasher,
		marshalizer:          args.Marshalizer,
		validatorInfoStorage: args.ValidatorInfoStorage,
		miniBlockStorage:     args.MiniBlockStorage,
		dataPool:             args.DataPool,
		enableEpochsHandler:  args.EnableEpochsHandler,
	}

	return vic, nil
}

// CreateValidatorInfoMiniBlocks creates the validatorInfo mini blocks according to the provided validatorInfo map
func (vic *validatorInfoCreator) CreateValidatorInfoMiniBlocks(validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if validatorsInfo == nil {
		return nil, epochStart.ErrNilValidatorInfo
	}

	vic.mutValidatorInfo.Lock()
	defer vic.mutValidatorInfo.Unlock()

	vic.clean()

	miniBlocks := make([]*block.MiniBlock, 0)

	for shardId := uint32(0); shardId < vic.shardCoordinator.NumberOfShards(); shardId++ {
		validators := validatorsInfo[shardId]
		if len(validators) == 0 {
			continue
		}

		miniBlock, err := vic.createMiniBlock(validators)
		if err != nil {
			return nil, err
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	validators := validatorsInfo[core.MetachainShardId]
	if len(validators) == 0 {
		return miniBlocks, nil
	}

	miniBlock, err := vic.createMiniBlock(validators)
	if err != nil {
		return nil, err
	}

	miniBlocks = append(miniBlocks, miniBlock)

	return miniBlocks, nil
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

		shardValidatorInfoData, err := vic.getShardValidatorInfoData(shardValidatorInfo)
		if err != nil {
			return nil, err
		}

		miniBlock.TxHashes[index] = shardValidatorInfoData
	}

	return miniBlock, nil
}

func (vic *validatorInfoCreator) getShardValidatorInfoData(shardValidatorInfo *state.ShardValidatorInfo) ([]byte, error) {
	if vic.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
		shardValidatorInfoHash, err := core.CalculateHash(vic.marshalizer, vic.hasher, shardValidatorInfo)
		if err != nil {
			return nil, err
		}

		validatorInfoCacher := vic.dataPool.CurrentEpochValidatorInfo()
		validatorInfoCacher.AddValidatorInfo(shardValidatorInfoHash, shardValidatorInfo)
		return shardValidatorInfoHash, nil
	}

	marshalledShardValidatorInfo, err := vic.marshalizer.Marshal(shardValidatorInfo)
	if err != nil {
		return nil, err
	}

	return marshalledShardValidatorInfo, nil
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

// VerifyValidatorInfoMiniBlocks verifies if received validator info mini blocks are correct
func (vic *validatorInfoCreator) VerifyValidatorInfoMiniBlocks(miniBlocks []*block.MiniBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error {
	if len(miniBlocks) == 0 {
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
	for _, receivedMb := range miniBlocks {
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
			// TODO: add display debug prints of mini blocks contents
			return epochStart.ErrValidatorMiniBlockHashDoesNotMatch
		}
	}

	if len(createdMiniBlocks) != numReceivedValidatorInfoMBs {
		return epochStart.ErrValidatorInfoMiniBlocksNumDoesNotMatch
	}

	return nil
}

// GetLocalValidatorInfoCache returns the local validator info cache which holds all the validator info for the current epoch
func (vic *validatorInfoCreator) GetLocalValidatorInfoCache() epochStart.ValidatorInfoCacher {
	return vic.dataPool.CurrentEpochValidatorInfo()
}

// CreateMarshalledData creates the marshalled data to be sent to shards
func (vic *validatorInfoCreator) CreateMarshalledData(body *block.Body) map[string][][]byte {
	if !vic.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
		return nil
	}

	if check.IfNil(body) {
		return nil
	}

	marshalledValidatorInfoTxs := make(map[string][][]byte)

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}
		if miniBlock.SenderShardID != vic.shardCoordinator.SelfId() ||
			miniBlock.ReceiverShardID == vic.shardCoordinator.SelfId() {
			continue
		}

		broadcastTopic := common.ValidatorInfoTopic
		if _, ok := marshalledValidatorInfoTxs[broadcastTopic]; !ok {
			marshalledValidatorInfoTxs[broadcastTopic] = make([][]byte, 0, len(miniBlock.TxHashes))
		}

		vic.setMarshalledValidatorInfoTxs(miniBlock, marshalledValidatorInfoTxs, broadcastTopic)

		if len(marshalledValidatorInfoTxs[broadcastTopic]) == 0 {
			delete(marshalledValidatorInfoTxs, broadcastTopic)
		}
	}

	return marshalledValidatorInfoTxs
}

func (vic *validatorInfoCreator) setMarshalledValidatorInfoTxs(miniBlock *block.MiniBlock, marshalledValidatorInfoTxs map[string][][]byte, broadcastTopic string) {
	validatorInfoCacher := vic.dataPool.CurrentEpochValidatorInfo()

	for _, txHash := range miniBlock.TxHashes {
		validatorInfoTx, err := validatorInfoCacher.GetValidatorInfo(txHash)
		if err != nil {
			log.Error("validatorInfoCreator.setMarshalledValidatorInfoTxs.GetValidatorInfo", "hash", txHash, "error", err)
			continue
		}

		marshalledData, err := vic.marshalizer.Marshal(validatorInfoTx)
		if err != nil {
			log.Error("validatorInfoCreator.setMarshalledValidatorInfoTxs.Marshal", "hash", txHash, "error", err)
			continue
		}

		marshalledValidatorInfoTxs[broadcastTopic] = append(marshalledValidatorInfoTxs[broadcastTopic], marshalledData)
	}
}

// GetValidatorInfoTxs returns validator info txs for the current epoch
func (vic *validatorInfoCreator) GetValidatorInfoTxs(body *block.Body) map[string]*state.ShardValidatorInfo {
	mapShardValidatorInfo := make(map[string]*state.ShardValidatorInfo)

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		vic.setMapShardValidatorInfo(miniBlock, mapShardValidatorInfo)
	}

	return mapShardValidatorInfo
}

func (vic *validatorInfoCreator) setMapShardValidatorInfo(miniBlock *block.MiniBlock, mapShardValidatorInfo map[string]*state.ShardValidatorInfo) {
	for _, txHash := range miniBlock.TxHashes {
		shardValidatorInfo, err := vic.getShardValidatorInfo(txHash)
		if err != nil {
			log.Error("validatorInfoCreator.setMapShardValidatorInfo", "hash", txHash, "error", err)
			continue
		}

		mapShardValidatorInfo[string(txHash)] = shardValidatorInfo
	}
}

func (vic *validatorInfoCreator) getShardValidatorInfo(txHash []byte) (*state.ShardValidatorInfo, error) {
	if vic.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
		validatorInfoCacher := vic.dataPool.CurrentEpochValidatorInfo()
		shardValidatorInfo, err := validatorInfoCacher.GetValidatorInfo(txHash)
		if err != nil {
			return nil, err
		}

		return shardValidatorInfo, nil
	}

	shardValidatorInfo := &state.ShardValidatorInfo{}
	err := vic.marshalizer.Unmarshal(shardValidatorInfo, txHash)
	if err != nil {
		return nil, err
	}

	return shardValidatorInfo, nil
}

// SaveBlockDataToStorage saves block data to storage
func (vic *validatorInfoCreator) SaveBlockDataToStorage(_ data.HeaderHandler, body *block.Body) {
	if check.IfNil(body) {
		return
	}

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		if vic.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
			vic.saveValidatorInfo(miniBlock)
		}

		marshalledData, err := vic.marshalizer.Marshal(miniBlock)
		if err != nil {
			log.Error("validatorInfoCreator.SaveBlockDataToStorage.Marshal", "error", err)
			continue
		}

		mbHash := vic.hasher.Compute(string(marshalledData))
		_ = vic.miniBlockStorage.Put(mbHash, marshalledData)
	}
}

func (vic *validatorInfoCreator) saveValidatorInfo(miniBlock *block.MiniBlock) {
	validatorInfoCacher := vic.dataPool.CurrentEpochValidatorInfo()

	for _, validatorInfoHash := range miniBlock.TxHashes {
		validatorInfo, err := validatorInfoCacher.GetValidatorInfo(validatorInfoHash)
		if err != nil {
			log.Error("validatorInfoCreator.saveValidatorInfo.GetValidatorInfo", "hash", validatorInfoHash, "error", err)
			continue
		}

		marshalledData, err := vic.marshalizer.Marshal(validatorInfo)
		if err != nil {
			log.Error("validatorInfoCreator.saveValidatorInfo.Marshal", "hash", validatorInfoHash, "error", err)
			continue
		}

		_ = vic.validatorInfoStorage.Put(validatorInfoHash, marshalledData)
	}
}

// DeleteBlockDataFromStorage deletes block data from storage
func (vic *validatorInfoCreator) DeleteBlockDataFromStorage(metaBlock data.HeaderHandler, body *block.Body) {
	if check.IfNil(metaBlock) || check.IfNil(body) {
		return
	}

	if vic.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
		vic.removeValidatorInfoFromStorage(body)
	}

	for _, mbHeader := range metaBlock.GetMiniBlockHeaderHandlers() {
		if mbHeader.GetTypeInt32() == int32(block.PeerBlock) {
			_ = vic.miniBlockStorage.Remove(mbHeader.GetHash())
		}
	}
}

func (vic *validatorInfoCreator) removeValidatorInfoFromStorage(body *block.Body) {
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			_ = vic.validatorInfoStorage.Remove(txHash)
		}
	}
}

// RemoveBlockDataFromPools removes block data from pools
func (vic *validatorInfoCreator) RemoveBlockDataFromPools(metaBlock data.HeaderHandler, body *block.Body) {
	if check.IfNil(metaBlock) {
		return
	}

	if vic.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
		vic.removeValidatorInfoFromPool(body)
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

func (vic *validatorInfoCreator) removeValidatorInfoFromPool(body *block.Body) {
	validatorInfoPool := vic.dataPool.ValidatorsInfo()

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			validatorInfoPool.RemoveDataFromAllShards(txHash)
		}
	}
}

func (vic *validatorInfoCreator) clean() {
	currentEpochValidatorInfo := vic.dataPool.CurrentEpochValidatorInfo()
	currentEpochValidatorInfo.Clean()
}

// IsInterfaceNil returns true if underlying object is nil
func (vic *validatorInfoCreator) IsInterfaceNil() bool {
	return vic == nil
}
