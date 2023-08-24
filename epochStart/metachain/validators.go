package metachain

import (
	"bytes"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/compatibility"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
)

var _ process.EpochStartValidatorInfoCreator = (*validatorInfoCreator)(nil)

// ArgsNewValidatorInfoCreator defines the arguments structure needed to create a new validatorInfo creator
type ArgsNewValidatorInfoCreator struct {
	ShardCoordinator        sharding.Coordinator
	ValidatorInfoStorage    storage.Storer
	MiniBlockStorage        storage.Storer
	Hasher                  hashing.Hasher
	Marshalizer             marshal.Marshalizer
	DataPool                dataRetriever.PoolsHolder
	EnableEpochsHandler     common.EnableEpochsHandler
	EpochStartStaticStorage storage.Storer
}

type validatorInfoCreator struct {
	shardCoordinator        sharding.Coordinator
	validatorInfoStorage    storage.Storer
	miniBlockStorage        storage.Storer
	hasher                  hashing.Hasher
	marshalizer             marshal.Marshalizer
	dataPool                dataRetriever.PoolsHolder
	enableEpochsHandler     common.EnableEpochsHandler
	epochStartStaticStorage storage.Storer
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
	if check.IfNil(args.EpochStartStaticStorage) {
		return nil, epochStart.ErrNilStorage
	}

	vic := &validatorInfoCreator{
		shardCoordinator:        args.ShardCoordinator,
		hasher:                  args.Hasher,
		marshalizer:             args.Marshalizer,
		validatorInfoStorage:    args.ValidatorInfoStorage,
		miniBlockStorage:        args.MiniBlockStorage,
		dataPool:                args.DataPool,
		enableEpochsHandler:     args.EnableEpochsHandler,
		epochStartStaticStorage: args.EpochStartStaticStorage,
	}

	return vic, nil
}

// CreateValidatorInfoMiniBlocks creates the validatorInfo mini blocks according to the provided validatorInfo map
func (vic *validatorInfoCreator) CreateValidatorInfoMiniBlocks(validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if validatorsInfo == nil {
		return nil, epochStart.ErrNilValidatorInfo
	}

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

	validatorsCopy := make([]*state.ValidatorInfo, len(validatorsInfo))
	copy(validatorsCopy, validatorsInfo)

	vic.sortValidators(validatorsCopy)

	for index, validator := range validatorsCopy {
		shardValidatorInfo := createShardValidatorInfo(validator)

		shardValidatorInfoData, err := vic.getShardValidatorInfoData(shardValidatorInfo)
		if err != nil {
			return nil, err
		}

		miniBlock.TxHashes[index] = shardValidatorInfoData
	}

	return miniBlock, nil
}

func (vic *validatorInfoCreator) sortValidators(validators []*state.ValidatorInfo) {
	if vic.enableEpochsHandler.IsDeterministicSortOnValidatorsInfoFixEnabled() {
		vic.deterministicSortValidators(validators)
		return
	}

	vic.legacySortValidators(validators)
}

func (vic *validatorInfoCreator) deterministicSortValidators(validators []*state.ValidatorInfo) {
	sort.SliceStable(validators, func(a, b int) bool {
		result := bytes.Compare(validators[a].PublicKey, validators[b].PublicKey)
		if result != 0 {
			return result < 0
		}

		aValidatorString := validators[a].GoString()
		bValidatorString := validators[b].GoString()
		// possible issues as we have 2 entries with the same public key. Print & assure deterministic sorting
		log.Warn("found 2 entries in validatorInfoCreator.deterministicSortValidators with the same public key",
			"validator a", aValidatorString, "validator b", bValidatorString)

		// since the GoString will include all fields, we do not need to marshal the struct again. Strings comparison will
		// suffice in this case.
		return aValidatorString < bValidatorString
	})
}

func (vic *validatorInfoCreator) legacySortValidators(validators []*state.ValidatorInfo) {
	swap := func(a, b int) {
		validators[a], validators[b] = validators[b], validators[a]
	}
	less := func(a, b int) bool {
		return bytes.Compare(validators[a].PublicKey, validators[b].PublicKey) < 0
	}
	compatibility.SortSlice(swap, less, len(validators))
}

func (vic *validatorInfoCreator) getShardValidatorInfoData(shardValidatorInfo *state.ShardValidatorInfo) ([]byte, error) {
	if vic.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
		return vic.getShardValidatorInfoHash(shardValidatorInfo)
	}

	marshalledShardValidatorInfo, err := vic.marshalizer.Marshal(shardValidatorInfo)
	if err != nil {
		return nil, err
	}

	return marshalledShardValidatorInfo, nil
}

func (vic *validatorInfoCreator) getShardValidatorInfoHash(shardValidatorInfo *state.ShardValidatorInfo) ([]byte, error) {
	shardValidatorInfoHash, err := core.CalculateHash(vic.marshalizer, vic.hasher, shardValidatorInfo)
	if err != nil {
		return nil, err
	}

	validatorInfoCacher := vic.dataPool.CurrentEpochValidatorInfo()
	validatorInfoCacher.AddValidatorInfo(shardValidatorInfoHash, shardValidatorInfo)
	return shardValidatorInfoHash, nil
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
			vic.printAllMiniBlocks(createdMiniBlocks, miniBlocks)
			return epochStart.ErrValidatorMiniBlockHashDoesNotMatch
		}
	}

	if len(createdMiniBlocks) != numReceivedValidatorInfoMBs {
		return epochStart.ErrValidatorInfoMiniBlocksNumDoesNotMatch
	}

	return nil
}

func (vic *validatorInfoCreator) printAllMiniBlocks(created []*block.MiniBlock, received []*block.MiniBlock) {
	log.Debug("validatorInfoCreator.printAllMiniBlocks - printing created")
	for i, mb := range created {
		if mb == nil {
			log.Warn("nil miniblock found in printAllMiniBlocks, range on created", "index", i)
			continue
		}

		vic.printMiniBlock(mb, i)
	}

	log.Debug("validatorInfoCreator.printAllMiniBlocks - printing received")
	for i, mb := range received {
		if mb == nil {
			log.Warn("nil miniblock found in printAllMiniBlocks, range on received", "index", i)
			continue
		}

		vic.printMiniBlock(mb, i)
	}
}

func (vic *validatorInfoCreator) printMiniBlock(mb *block.MiniBlock, index int) {
	hashes := make([]string, 0, len(mb.TxHashes))
	for _, hash := range mb.TxHashes {
		hashes = append(hashes, hex.EncodeToString(hash))
	}

	mbHash, err := core.CalculateHash(vic.marshalizer, vic.hasher, mb)
	if err != nil {
		log.Warn("error computing hash in validatorInfoCreator.printMiniBlock",
			"index", index, "error", err)
		return
	}

	txHashSeparator := "\n   "
	log.Debug(" miniblock",
		"hash", mbHash,
		"type", mb.Type.String(),
		"sender shard ID", mb.SenderShardID,
		"receiver shard ID", mb.ReceiverShardID,
		"hashes", txHashSeparator+strings.Join(hashes, txHashSeparator))
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

	marshalledValidatorInfoTxs := make([][]byte, 0)
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}
		isCrossMiniBlockFromMe := miniBlock.SenderShardID == vic.shardCoordinator.SelfId() &&
			miniBlock.ReceiverShardID != vic.shardCoordinator.SelfId()
		if !isCrossMiniBlockFromMe {
			continue
		}

		marshalledValidatorInfoTxs = append(marshalledValidatorInfoTxs, vic.getMarshalledValidatorInfoTxs(miniBlock)...)
	}

	mapMarshalledValidatorInfoTxs := make(map[string][][]byte)
	if len(marshalledValidatorInfoTxs) > 0 {
		mapMarshalledValidatorInfoTxs[common.ValidatorInfoTopic] = marshalledValidatorInfoTxs
	}

	return mapMarshalledValidatorInfoTxs
}

func (vic *validatorInfoCreator) getMarshalledValidatorInfoTxs(miniBlock *block.MiniBlock) [][]byte {
	validatorInfoCacher := vic.dataPool.CurrentEpochValidatorInfo()

	marshalledValidatorInfoTxs := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		validatorInfoTx, err := validatorInfoCacher.GetValidatorInfo(txHash)
		if err != nil {
			log.Error("validatorInfoCreator.getMarshalledValidatorInfoTxs.GetValidatorInfo", "hash", txHash, "error", err)
			continue
		}

		marshalledData, err := vic.marshalizer.Marshal(validatorInfoTx)
		if err != nil {
			log.Error("validatorInfoCreator.getMarshalledValidatorInfoTxs.Marshal", "hash", txHash, "error", err)
			continue
		}

		marshalledValidatorInfoTxs = append(marshalledValidatorInfoTxs, marshalledData)
	}

	return marshalledValidatorInfoTxs
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
		err = vic.miniBlockStorage.Put(mbHash, marshalledData)
		if err != nil {
			log.Debug("validatorInfoCreator.SaveBlockDataToStorage.Put", "hash", mbHash, "error", err)
		}
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

		err = vic.validatorInfoStorage.Put(validatorInfoHash, marshalledData)
		if err != nil {
			log.Debug("validatorInfoCreator.saveValidatorInfo.Put", "hash", validatorInfoHash, "error", err)
			continue
		}

		err = vic.epochStartStaticStorage.Put(validatorInfoHash, marshalledData)
		if err != nil {
			log.Debug("validatorInfoCreator.epochStartStaticStorage.Put", "hash", validatorInfoHash, "error", err)
			continue
		}

		log.Trace("saveValidatorInfo: put validatorInfo into epochStartStaticStorage", "validatorInfo hash", validatorInfoHash)
	}
}

// DeleteBlockDataFromStorage deletes block data from storage
func (vic *validatorInfoCreator) DeleteBlockDataFromStorage(metaBlock data.HeaderHandler, body *block.Body) {
	if check.IfNil(metaBlock) || check.IfNil(body) {
		return
	}

	if vic.enableEpochsHandler.IsRefactorPeersMiniBlocksFlagEnabled() {
		vic.removeValidatorInfo(body)
	}

	for _, mbHeader := range metaBlock.GetMiniBlockHeaderHandlers() {
		if mbHeader.GetTypeInt32() == int32(block.PeerBlock) {
			err := vic.miniBlockStorage.Remove(mbHeader.GetHash())
			if err != nil {
				log.Debug("validatorInfoCreator.DeleteBlockDataFromStorage.Remove", "hash", mbHeader.GetHash(), "error", err)
			}
		}
	}
}

func (vic *validatorInfoCreator) removeValidatorInfo(body *block.Body) {
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		vic.removeValidatorInfoFromStorage(miniBlock)
	}
}

func (vic *validatorInfoCreator) removeValidatorInfoFromStorage(miniBlock *block.MiniBlock) {
	for _, txHash := range miniBlock.TxHashes {
		err := vic.validatorInfoStorage.Remove(txHash)
		if err != nil {
			log.Debug("validatorInfoCreator.removeValidatorInfoFromStorage.Remove", "hash", txHash, "error", err)
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
