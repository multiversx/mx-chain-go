package metachain

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.EpochStartRewardsCreator = (*rewardsCreator)(nil)

// ArgsNewRewardsCreator defines the arguments structure needed to create a new rewards creator
type ArgsNewRewardsCreator struct {
	ShardCoordinator sharding.Coordinator
	PubkeyConverter  state.PubkeyConverter
	RewardsStorage   storage.Storer
	MiniBlockStorage storage.Storer
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
	DataPool         dataRetriever.PoolsHolder
}

type rewardsCreator struct {
	currTxs          dataRetriever.TransactionCacher
	shardCoordinator sharding.Coordinator
	pubkeyConverter  state.PubkeyConverter
	rewardsStorage   storage.Storer
	miniBlockStorage storage.Storer

	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
	dataPool    dataRetriever.PoolsHolder
}

type rewardInfoData struct {
	LeaderSuccess              uint32
	LeaderFailure              uint32
	ValidatorSuccess           uint32
	ValidatorFailure           uint32
	NumSelectedInSuccessBlocks uint32
	AccumulatedFees            *big.Int
}

// NewEpochStartRewardsCreator creates a new rewards creator object
func NewEpochStartRewardsCreator(args ArgsNewRewardsCreator) (*rewardsCreator, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, epochStart.ErrNilShardCoordinator
	}
	if check.IfNil(args.PubkeyConverter) {
		return nil, epochStart.ErrNilPubkeyConverter
	}
	if check.IfNil(args.RewardsStorage) {
		return nil, epochStart.ErrNilStorage
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

	currTxsCache, err := dataPool.NewCurrentBlockPool()
	if err != nil {
		return nil, err
	}

	rc := &rewardsCreator{
		currTxs:          currTxsCache,
		shardCoordinator: args.ShardCoordinator,
		pubkeyConverter:  args.PubkeyConverter,
		rewardsStorage:   args.RewardsStorage,
		hasher:           args.Hasher,
		marshalizer:      args.Marshalizer,
		miniBlockStorage: args.MiniBlockStorage,
		dataPool:         args.DataPool,
	}

	return rc, nil
}

// CreateBlockStarted announces block creation started and cleans inside data
func (rc *rewardsCreator) clean() {
	rc.currTxs.Clean()
}

// CreateRewardsMiniBlocks creates the rewards miniblocks according to economics data and validator info
func (rc *rewardsCreator) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}

	rc.clean()

	miniBlocks := make(block.MiniBlockSlice, rc.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < rc.shardCoordinator.NumberOfShards(); i++ {
		miniBlocks[i] = &block.MiniBlock{}
		miniBlocks[i].SenderShardID = core.MetachainShardId
		miniBlocks[i].ReceiverShardID = i
		miniBlocks[i].Type = block.RewardsBlock
		miniBlocks[i].TxHashes = make([][]byte, 0)
	}

	err := rc.addCommunityRewardToMiniBlocks(miniBlocks, metaBlock)
	if err != nil {
		return nil, err
	}

	err = rc.addValidatorRewardsToMiniBlocks(validatorsInfo, metaBlock, miniBlocks)
	if err != nil {
		return nil, err
	}

	for shId := uint32(0); shId < rc.shardCoordinator.NumberOfShards(); shId++ {
		sort.Slice(miniBlocks[shId].TxHashes, func(i, j int) bool {
			return bytes.Compare(miniBlocks[shId].TxHashes[i], miniBlocks[shId].TxHashes[j]) < 0
		})
	}

	finalMiniBlocks := make(block.MiniBlockSlice, 0)
	for i := uint32(0); i < rc.shardCoordinator.NumberOfShards(); i++ {
		if len(miniBlocks[i].TxHashes) > 0 {
			finalMiniBlocks = append(finalMiniBlocks, miniBlocks[i])
		}
	}

	return finalMiniBlocks, nil
}

func (rc *rewardsCreator) addValidatorRewardsToMiniBlocks(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	metaBlock *block.MetaBlock,
	miniBlocks block.MiniBlockSlice,
) error {
	rwdAddrValidatorInfo := rc.computeValidatorInfoPerRewardAddress(validatorsInfo)
	for address, rwdInfo := range rwdAddrValidatorInfo {
		addrContainer, err := rc.pubkeyConverter.CreateAddressFromBytes([]byte(address))
		if err != nil {
			log.Warn("invalid reward address from validator info", "err", err, "provided address", address)
			continue
		}

		rwdTx, rwdTxHash, err := rc.createRewardFromRwdInfo([]byte(address), rwdInfo, &metaBlock.EpochStart.Economics, metaBlock)
		if err != nil {
			return err
		}

		rc.currTxs.AddTx(rwdTxHash, rwdTx)

		shardId := rc.shardCoordinator.ComputeId(addrContainer)
		miniBlocks[shardId].TxHashes = append(miniBlocks[shardId].TxHashes, rwdTxHash)
	}

	return nil
}

func (rc *rewardsCreator) addCommunityRewardToMiniBlocks(
	miniBlocks block.MiniBlockSlice,
	epochStartMetablock *block.MetaBlock,
) error {
	rwdTx, rwdTxHash, shardId, err := rc.createCommunityRewardTransaction(epochStartMetablock)
	if err != nil {
		return err
	}
	rc.currTxs.AddTx(rwdTxHash, rwdTx)
	miniBlocks[shardId].TxHashes = append(miniBlocks[shardId].TxHashes, rwdTxHash)

	return nil
}

func (rc *rewardsCreator) createCommunityRewardTransaction(
	metaBlock *block.MetaBlock,
) (*rewardTx.RewardTx, []byte, uint32, error) {

	communityAddress := metaBlock.EpochStart.Economics.CommunityAddress
	addressContainer, err := rc.pubkeyConverter.CreateAddressFromString(string(communityAddress))
	if err != nil {
		log.Warn("invalid community reward address", "err", err, "provided address", communityAddress)
		return nil, nil, 0, err
	}

	shardID := rc.shardCoordinator.ComputeId(addressContainer)

	communityRwdTx := &rewardTx.RewardTx{
		Round:   metaBlock.GetRound(),
		Value:   big.NewInt(0).Set(metaBlock.EpochStart.Economics.RewardsForCommunity),
		RcvAddr: addressContainer.Bytes(),
		Epoch:   metaBlock.Epoch,
	}

	txHash, err := core.CalculateHash(rc.marshalizer, rc.hasher, communityRwdTx)
	if err != nil {
		return nil, nil, 0, err
	}

	return communityRwdTx, txHash, shardID, nil
}

func (rc *rewardsCreator) computeValidatorInfoPerRewardAddress(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) map[string]*rewardInfoData {

	rwdAddrValidatorInfo := make(map[string]*rewardInfoData)

	for _, shardValidatorsInfo := range validatorsInfo {
		for _, validatorInfo := range shardValidatorsInfo {
			rwdInfo, ok := rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)]
			if !ok {
				rwdInfo = &rewardInfoData{
					AccumulatedFees: big.NewInt(0),
				}
				rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)] = rwdInfo
			}

			rwdInfo.LeaderSuccess += validatorInfo.LeaderSuccess
			rwdInfo.LeaderFailure += validatorInfo.LeaderFailure
			rwdInfo.ValidatorFailure += validatorInfo.ValidatorFailure
			rwdInfo.ValidatorSuccess += validatorInfo.ValidatorSuccess
			rwdInfo.NumSelectedInSuccessBlocks += validatorInfo.NumSelectedInSuccessBlocks

			rwdInfo.AccumulatedFees.Add(rwdInfo.AccumulatedFees, validatorInfo.AccumulatedFees)
		}
	}

	return rwdAddrValidatorInfo
}

func (rc *rewardsCreator) createRewardFromRwdInfo(
	address []byte,
	rwdInfo *rewardInfoData,
	economicsData *block.Economics,
	metaBlock *block.MetaBlock,
) (*rewardTx.RewardTx, []byte, error) {
	rwdTx := &rewardTx.RewardTx{
		Round:   metaBlock.GetRound(),
		Value:   big.NewInt(0).Set(rwdInfo.AccumulatedFees),
		RcvAddr: address,
		Epoch:   metaBlock.Epoch,
	}

	protocolRewardValue := big.NewInt(0).Mul(economicsData.RewardsPerBlockPerNode, big.NewInt(0).SetUint64(uint64(rwdInfo.NumSelectedInSuccessBlocks)))
	rwdTx.Value.Add(rwdTx.Value, protocolRewardValue)

	rwdTxHash, err := core.CalculateHash(rc.marshalizer, rc.hasher, rwdTx)
	if err != nil {
		return nil, nil, err
	}

	return rwdTx, rwdTxHash, nil
}

// VerifyRewardsMiniBlocks verifies if received rewards miniblocks are correct
func (rc *rewardsCreator) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error {
	if check.IfNil(metaBlock) {
		return epochStart.ErrNilHeaderHandler
	}

	createdMiniBlocks, err := rc.CreateRewardsMiniBlocks(metaBlock, validatorsInfo)
	if err != nil {
		return err
	}

	numReceivedRewardsMBs := 0
	for _, miniBlockHdr := range metaBlock.MiniBlockHeaders {
		if miniBlockHdr.Type != block.RewardsBlock {
			continue
		}

		numReceivedRewardsMBs++
		createdMiniBlock := createdMiniBlocks[miniBlockHdr.ReceiverShardID]
		createdMBHash, errComputeHash := core.CalculateHash(rc.marshalizer, rc.hasher, createdMiniBlock)
		if errComputeHash != nil {
			return errComputeHash
		}

		if !bytes.Equal(createdMBHash, miniBlockHdr.Hash) {
			// TODO: add display debug prints of miniblocks contents
			return epochStart.ErrRewardMiniBlockHashDoesNotMatch
		}
	}

	if len(createdMiniBlocks) != numReceivedRewardsMBs {
		return epochStart.ErrRewardMiniBlocksNumDoesNotMatch
	}

	return nil
}

// CreateMarshalizedData creates the marshalized data to be sent to shards
func (rc *rewardsCreator) CreateMarshalizedData(body *block.Body) map[string][][]byte {
	if check.IfNil(body) {
		return nil
	}

	txs := make(map[string][][]byte)

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		broadCastTopic := createBroadcastTopic(rc.shardCoordinator, miniBlock.ReceiverShardID)
		if _, ok := txs[broadCastTopic]; !ok {
			txs[broadCastTopic] = make([][]byte, 0, len(miniBlock.TxHashes))
		}

		for _, txHash := range miniBlock.TxHashes {
			rwdTx, err := rc.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			marshalizedData, err := rc.marshalizer.Marshal(rwdTx)
			if err != nil {
				continue
			}

			txs[broadCastTopic] = append(txs[broadCastTopic], marshalizedData)
		}
	}

	return txs
}

// GetRewardsTxs will return rewards txs MUST be called before SaveTxBlockToStorage
func (rc *rewardsCreator) GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler {
	rewardsTxs := make(map[string]data.TransactionHandler)
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			rwTx, err := rc.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			rewardsTxs[string(txHash)] = rwTx
		}
	}

	return rewardsTxs
}

// SaveTxBlockToStorage saves created data to storage
func (rc *rewardsCreator) SaveTxBlockToStorage(_ *block.MetaBlock, body *block.Body) {
	if check.IfNil(body) {
		return
	}

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			rwdTx, err := rc.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			marshalizedData, err := rc.marshalizer.Marshal(rwdTx)
			if err != nil {
				continue
			}

			_ = rc.rewardsStorage.Put(txHash, marshalizedData)
		}

		marshalizedData, err := rc.marshalizer.Marshal(miniBlock)
		if err != nil {
			continue
		}

		mbHash := rc.hasher.Compute(string(marshalizedData))
		_ = rc.miniBlockStorage.Put(mbHash, marshalizedData)
	}
	rc.clean()
}

// DeleteTxsFromStorage deletes data from storage
func (rc *rewardsCreator) DeleteTxsFromStorage(metaBlock *block.MetaBlock, body *block.Body) {
	if check.IfNil(metaBlock) || check.IfNil(body) {
		return
	}

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			_ = rc.rewardsStorage.Remove(txHash)
		}
	}

	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		if mbHeader.Type == block.RewardsBlock {
			_ = rc.miniBlockStorage.Remove(mbHeader.Hash)
		}
	}
}

// IsInterfaceNil return true if underlying object is nil
func (rc *rewardsCreator) IsInterfaceNil() bool {
	return rc == nil
}

func createBroadcastTopic(shardC sharding.Coordinator, destShId uint32) string {
	transactionTopic := factory.RewardsTransactionTopic +
		shardC.CommunicationIdentifier(destShId)
	return transactionTopic
}

// RemoveBlockDataFromPools removes block info from pools
func (rc *rewardsCreator) RemoveBlockDataFromPools(metaBlock *block.MetaBlock, body *block.Body) {
	if check.IfNil(metaBlock) || check.IfNil(body) {
		return
	}

	transactionsPool := rc.dataPool.Transactions()
	miniBlocksPool := rc.dataPool.MiniBlocks()

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		transactionsPool.RemoveSetOfDataFromPool(miniBlock.TxHashes, strCache)
	}

	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		if mbHeader.Type != block.RewardsBlock {
			continue
		}

		miniBlocksPool.Remove(mbHeader.Hash)

		log.Trace("RemoveBlockDataFromPools",
			"hash", mbHeader.Hash,
			"type", mbHeader.Type,
			"sender", mbHeader.SenderShardID,
			"receiver", mbHeader.ReceiverShardID,
			"num txs", mbHeader.TxCount)
	}
}
