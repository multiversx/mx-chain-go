package metachain

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
)

// BaseRewardsCreatorArgs defines the arguments structure needed to create a base rewards creator
type BaseRewardsCreatorArgs struct {
	ShardCoordinator              sharding.Coordinator
	PubkeyConverter               core.PubkeyConverter
	RewardsStorage                storage.Storer
	MiniBlockStorage              storage.Storer
	Hasher                        hashing.Hasher
	Marshalizer                   marshal.Marshalizer
	DataPool                      dataRetriever.PoolsHolder
	ProtocolSustainabilityAddress string
	NodesConfigProvider           epochStart.NodesConfigProvider
	UserAccountsDB                state.AccountsAdapter
	EnableEpochsHandler           common.EnableEpochsHandler
	ExecutionOrderHandler         common.TxExecutionOrderHandler
}

type baseRewardsCreator struct {
	currTxs                            dataRetriever.TransactionCacher
	shardCoordinator                   sharding.Coordinator
	pubkeyConverter                    core.PubkeyConverter
	rewardsStorage                     storage.Storer
	miniBlockStorage                   storage.Storer
	protocolSustainabilityAddress      []byte
	nodesConfigProvider                epochStart.NodesConfigProvider
	hasher                             hashing.Hasher
	marshalizer                        marshal.Marshalizer
	dataPool                           dataRetriever.PoolsHolder
	mapBaseRewardsPerBlockPerValidator map[uint32]*big.Int
	accumulatedRewards                 *big.Int
	protocolSustainabilityValue        *big.Int
	flagDelegationSystemSCEnabled      atomic.Flag //nolint
	userAccountsDB                     state.AccountsAdapter
	enableEpochsHandler                common.EnableEpochsHandler
	mutRewardsData                     sync.RWMutex
	executionOrderHandler              common.TxExecutionOrderHandler
}

// NewBaseRewardsCreator will create a new base rewards creator instance
func NewBaseRewardsCreator(args BaseRewardsCreatorArgs) (*baseRewardsCreator, error) {
	err := checkBaseArgs(args)
	if err != nil {
		return nil, err
	}

	address, err := args.PubkeyConverter.Decode(args.ProtocolSustainabilityAddress)
	if err != nil {
		log.Warn("invalid protocol sustainability reward address", "err", err, "provided address", args.ProtocolSustainabilityAddress)
		return nil, err
	}

	protocolSustainabilityShardID := args.ShardCoordinator.ComputeId(address)
	if protocolSustainabilityShardID == core.MetachainShardId {
		return nil, epochStart.ErrProtocolSustainabilityAddressInMetachain
	}

	currTxsCache := dataPool.NewCurrentBlockTransactionsPool()
	brc := &baseRewardsCreator{
		currTxs:                            currTxsCache,
		shardCoordinator:                   args.ShardCoordinator,
		pubkeyConverter:                    args.PubkeyConverter,
		rewardsStorage:                     args.RewardsStorage,
		hasher:                             args.Hasher,
		marshalizer:                        args.Marshalizer,
		miniBlockStorage:                   args.MiniBlockStorage,
		dataPool:                           args.DataPool,
		protocolSustainabilityAddress:      address,
		nodesConfigProvider:                args.NodesConfigProvider,
		accumulatedRewards:                 big.NewInt(0),
		protocolSustainabilityValue:        big.NewInt(0),
		userAccountsDB:                     args.UserAccountsDB,
		mapBaseRewardsPerBlockPerValidator: make(map[uint32]*big.Int),
		enableEpochsHandler:                args.EnableEpochsHandler,
		executionOrderHandler:              args.ExecutionOrderHandler,
	}

	return brc, nil
}

// GetProtocolSustainabilityRewards returns the sum of all rewards
func (brc *baseRewardsCreator) GetProtocolSustainabilityRewards() *big.Int {
	brc.mutRewardsData.RLock()
	defer brc.mutRewardsData.RUnlock()

	return brc.protocolSustainabilityValue
}

// GetLocalTxCache returns the local tx cache which holds all the rewards
func (brc *baseRewardsCreator) GetLocalTxCache() epochStart.TransactionCacher {
	return brc.currTxs
}

// CreateMarshalledData creates the marshalled data to be sent to shards
func (brc *baseRewardsCreator) CreateMarshalledData(body *block.Body) map[string][][]byte {
	if check.IfNil(body) {
		return nil
	}

	marshalledRewardsTxs := make(map[string][][]byte)

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}
		if miniBlock.SenderShardID != brc.shardCoordinator.SelfId() ||
			miniBlock.ReceiverShardID == brc.shardCoordinator.SelfId() {
			continue
		}

		broadcastTopic := createBroadcastTopic(brc.shardCoordinator, miniBlock.ReceiverShardID)
		if _, ok := marshalledRewardsTxs[broadcastTopic]; !ok {
			marshalledRewardsTxs[broadcastTopic] = make([][]byte, 0, len(miniBlock.TxHashes))
		}

		for _, txHash := range miniBlock.TxHashes {
			rwdTx, err := brc.currTxs.GetTx(txHash)
			if err != nil {
				log.Error("rewardsCreator.CreateMarshalledData.GetTx", "hash", txHash, "error", err)
				continue
			}

			marshalledData, err := brc.marshalizer.Marshal(rwdTx)
			if err != nil {
				log.Error("rewardsCreator.CreateMarshalledData.Marshal", "hash", txHash, "error", err)
				continue
			}

			marshalledRewardsTxs[broadcastTopic] = append(marshalledRewardsTxs[broadcastTopic], marshalledData)
		}

		if len(marshalledRewardsTxs[broadcastTopic]) == 0 {
			delete(marshalledRewardsTxs, broadcastTopic)
		}
	}

	return marshalledRewardsTxs
}

// GetRewardsTxs returns rewards txs for the current block
func (brc *baseRewardsCreator) GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler {
	rewardsTxs := make(map[string]data.TransactionHandler)
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			rwTx, err := brc.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			rewardsTxs[string(txHash)] = rwTx
		}
	}

	return rewardsTxs
}

// SaveBlockDataToStorage saves block data to storage
func (brc *baseRewardsCreator) SaveBlockDataToStorage(_ data.MetaHeaderHandler, body *block.Body) {
	if check.IfNil(body) {
		return
	}

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			rwdTx, err := brc.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			marshalizedData, err := brc.marshalizer.Marshal(rwdTx)
			if err != nil {
				continue
			}

			_ = brc.rewardsStorage.Put(txHash, marshalizedData)
		}

		marshalizedData, err := brc.marshalizer.Marshal(miniBlock)
		if err != nil {
			continue
		}

		mbHash := brc.hasher.Compute(string(marshalizedData))
		_ = brc.miniBlockStorage.Put(mbHash, marshalizedData)
	}
}

// DeleteBlockDataFromStorage deletes block data from storage
func (brc *baseRewardsCreator) DeleteBlockDataFromStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if check.IfNil(metaBlock) || check.IfNil(body) {
		return
	}

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			_ = brc.rewardsStorage.Remove(txHash)
		}
	}

	for _, mbHeader := range metaBlock.GetMiniBlockHeaderHandlers() {
		if mbHeader.GetTypeInt32() == int32(block.RewardsBlock) {
			_ = brc.miniBlockStorage.Remove(mbHeader.GetHash())
		}
	}
}

// RemoveBlockDataFromPools removes block data from pools
func (brc *baseRewardsCreator) RemoveBlockDataFromPools(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if check.IfNil(metaBlock) || check.IfNil(body) {
		return
	}

	transactionsPool := brc.dataPool.Transactions()
	miniBlocksPool := brc.dataPool.MiniBlocks()

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		transactionsPool.RemoveSetOfDataFromPool(miniBlock.TxHashes, strCache)
	}

	for _, mbHeader := range metaBlock.GetMiniBlockHeaderHandlers() {
		if mbHeader.GetTypeInt32() != int32(block.RewardsBlock) {
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

func checkBaseArgs(args BaseRewardsCreatorArgs) error {
	if check.IfNil(args.ShardCoordinator) {
		return epochStart.ErrNilShardCoordinator
	}
	if check.IfNil(args.PubkeyConverter) {
		return epochStart.ErrNilPubkeyConverter
	}
	if check.IfNil(args.RewardsStorage) {
		return epochStart.ErrNilStorage
	}
	if check.IfNil(args.Marshalizer) {
		return epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return epochStart.ErrNilHasher
	}
	if check.IfNil(args.MiniBlockStorage) {
		return epochStart.ErrNilStorage
	}
	if check.IfNil(args.DataPool) {
		return epochStart.ErrNilDataPoolsHolder
	}
	if len(args.ProtocolSustainabilityAddress) == 0 {
		return epochStart.ErrNilProtocolSustainabilityAddress
	}
	if check.IfNil(args.NodesConfigProvider) {
		return epochStart.ErrNilNodesConfigProvider
	}
	if check.IfNil(args.UserAccountsDB) {
		return epochStart.ErrNilAccountsDB
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return epochStart.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.ExecutionOrderHandler) {
		return epochStart.ErrNilExecutionOrderHandler
	}

	return nil
}

// CreateBlockStarted announces block creation started and cleans inside data
func (brc *baseRewardsCreator) clean() {
	brc.mapBaseRewardsPerBlockPerValidator = make(map[uint32]*big.Int)
	brc.currTxs.Clean()
	brc.accumulatedRewards = big.NewInt(0)
	brc.protocolSustainabilityValue = big.NewInt(0)
}

func (brc *baseRewardsCreator) isSystemDelegationSC(address []byte) bool {
	acc, errExist := brc.userAccountsDB.GetExistingAccount(address)
	if errExist != nil {
		return false
	}

	userAcc, ok := acc.(state.UserAccountHandler)
	if !ok {
		return false
	}

	val, _, err := userAcc.RetrieveValue([]byte(core.DelegationSystemSCKey))
	if err != nil {
		return false
	}

	return len(val) > 0
}

func (brc *baseRewardsCreator) createProtocolSustainabilityRewardTransaction(
	metaBlock data.HeaderHandler,
	computedEconomics *block.Economics,
) (*rewardTx.RewardTx, uint32, error) {

	shardID := brc.shardCoordinator.ComputeId(brc.protocolSustainabilityAddress)
	protocolSustainabilityRwdTx := &rewardTx.RewardTx{
		Round:   metaBlock.GetRound(),
		Value:   big.NewInt(0).Set(computedEconomics.RewardsForProtocolSustainability),
		RcvAddr: brc.protocolSustainabilityAddress,
		Epoch:   metaBlock.GetEpoch(),
	}

	brc.accumulatedRewards.Add(brc.accumulatedRewards, protocolSustainabilityRwdTx.Value)
	return protocolSustainabilityRwdTx, shardID, nil
}

func (brc *baseRewardsCreator) createRewardFromRwdInfo(
	rwdInfo *rewardInfoData,
	metaBlock data.HeaderHandler,
) (*rewardTx.RewardTx, []byte, error) {
	rwdTx := &rewardTx.RewardTx{
		Round:   metaBlock.GetRound(),
		Value:   big.NewInt(0).Add(rwdInfo.accumulatedFees, rwdInfo.rewardsFromProtocol),
		RcvAddr: []byte(rwdInfo.address),
		Epoch:   metaBlock.GetEpoch(),
	}

	rwdTxHash, err := core.CalculateHash(brc.marshalizer, brc.hasher, rwdTx)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("rewardTx",
		"address", []byte(rwdInfo.address),
		"value", rwdTx.Value.String(),
		"hash", rwdTxHash,
		"accumulatedFees", rwdInfo.accumulatedFees,
		"rewardsFromProtocol", rwdInfo.rewardsFromProtocol,
	)

	return rwdTx, rwdTxHash, nil
}

func (brc *baseRewardsCreator) initializeRewardsMiniBlocks() block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, brc.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i <= brc.shardCoordinator.NumberOfShards(); i++ {
		miniBlocks[i] = &block.MiniBlock{
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: i,
			Type:            block.RewardsBlock,
			TxHashes:        make([][]byte, 0),
		}
	}
	miniBlocks[brc.shardCoordinator.NumberOfShards()].ReceiverShardID = core.MetachainShardId
	return miniBlocks
}

func (brc *baseRewardsCreator) addProtocolRewardToMiniBlocks(
	protocolSustainabilityRwdTx *rewardTx.RewardTx,
	miniBlocks block.MiniBlockSlice,
	protocolSustainabilityShardId uint32,
) error {
	protocolSustainabilityRwdHash, errHash := core.CalculateHash(brc.marshalizer, brc.hasher, protocolSustainabilityRwdTx)
	if errHash != nil {
		return errHash
	}

	brc.currTxs.AddTx(protocolSustainabilityRwdHash, protocolSustainabilityRwdTx)
	miniBlocks[protocolSustainabilityShardId].TxHashes = append(miniBlocks[protocolSustainabilityShardId].TxHashes, protocolSustainabilityRwdHash)
	brc.protocolSustainabilityValue.Set(protocolSustainabilityRwdTx.Value)

	return nil
}

func (brc *baseRewardsCreator) finalizeMiniBlocks(miniBlocks block.MiniBlockSlice) block.MiniBlockSlice {
	for shId := uint32(0); shId <= brc.shardCoordinator.NumberOfShards(); shId++ {
		sort.Slice(miniBlocks[shId].TxHashes, func(i, j int) bool {
			return bytes.Compare(miniBlocks[shId].TxHashes[i], miniBlocks[shId].TxHashes[j]) < 0
		})
	}

	finalMiniBlocks := make(block.MiniBlockSlice, 0)
	for i := uint32(0); i <= brc.shardCoordinator.NumberOfShards(); i++ {
		if len(miniBlocks[i].TxHashes) > 0 {
			finalMiniBlocks = append(finalMiniBlocks, miniBlocks[i])
			brc.addExecutionOrdering(miniBlocks[i].TxHashes)
		}
	}
	return finalMiniBlocks
}

func (brc *baseRewardsCreator) addExecutionOrdering(txHashes [][]byte) {
	for _, txHash := range txHashes {
		brc.executionOrderHandler.Add(txHash)
	}
}

func (brc *baseRewardsCreator) fillBaseRewardsPerBlockPerNode(baseRewardsPerNode *big.Int) {
	brc.mapBaseRewardsPerBlockPerValidator = make(map[uint32]*big.Int)
	for i := uint32(0); i < brc.shardCoordinator.NumberOfShards(); i++ {
		consensusSize := big.NewInt(int64(brc.nodesConfigProvider.ConsensusGroupSize(i)))
		brc.mapBaseRewardsPerBlockPerValidator[i] = big.NewInt(0).Div(baseRewardsPerNode, consensusSize)
		log.Debug("baseRewardsPerBlockPerValidator", "shardID", i, "value", brc.mapBaseRewardsPerBlockPerValidator[i].String())
	}

	consensusSize := big.NewInt(int64(brc.nodesConfigProvider.ConsensusGroupSize(core.MetachainShardId)))
	brc.mapBaseRewardsPerBlockPerValidator[core.MetachainShardId] = big.NewInt(0).Div(baseRewardsPerNode, consensusSize)
	log.Debug("baseRewardsPerBlockPerValidator", "shardID", core.MetachainShardId, "value", brc.mapBaseRewardsPerBlockPerValidator[core.MetachainShardId].String())
}

func (brc *baseRewardsCreator) verifyCreatedRewardMiniBlocksWithMetaBlock(metaBlock data.HeaderHandler, createdMiniBlocks block.MiniBlockSlice) error {
	numReceivedRewardsMBs := 0
	for _, miniBlockHdr := range metaBlock.GetMiniBlockHeaderHandlers() {
		if miniBlockHdr.GetTypeInt32() != int32(block.RewardsBlock) {
			continue
		}

		numReceivedRewardsMBs++
		createdMiniBlock := getMiniBlockWithReceiverShardID(miniBlockHdr.GetReceiverShardID(), createdMiniBlocks)
		if createdMiniBlock == nil {
			return epochStart.ErrRewardMiniBlockHashDoesNotMatch
		}

		createdMBHash, errComputeHash := core.CalculateHash(brc.marshalizer, brc.hasher, createdMiniBlock)
		if errComputeHash != nil {
			return errComputeHash
		}

		if !bytes.Equal(createdMBHash, miniBlockHdr.GetHash()) {
			generatedTxHashes := make([]string, 0, len(createdMiniBlock.TxHashes))
			for _, hash := range createdMiniBlock.TxHashes {
				generatedTxHashes = append(generatedTxHashes, hex.EncodeToString(hash))
			}

			log.Debug("rewardsCreator.VerifyRewardsMiniBlocks, generated reward tx hashes:\n" +
				strings.Join(generatedTxHashes, "\n"))
			log.Debug("rewardsCreator.VerifyRewardsMiniBlocks",
				"received mb hash", miniBlockHdr.GetHash(),
				"computed mb hash", createdMBHash,
			)

			return epochStart.ErrRewardMiniBlockHashDoesNotMatch
		}
	}

	if len(createdMiniBlocks) != numReceivedRewardsMBs {
		return epochStart.ErrRewardMiniBlocksNumDoesNotMatch
	}

	return nil
}

func getMiniBlockWithReceiverShardID(shardId uint32, miniBlocks block.MiniBlockSlice) *block.MiniBlock {
	for _, miniBlock := range miniBlocks {
		if miniBlock.ReceiverShardID == shardId {
			return miniBlock
		}
	}
	return nil
}

func createBroadcastTopic(shardC sharding.Coordinator, destShId uint32) string {
	transactionTopic := factory.RewardsTransactionTopic + shardC.CommunicationIdentifier(destShId)
	return transactionTopic
}

// IsInterfaceNil returns true if the underlying object is nil
func (brc *baseRewardsCreator) IsInterfaceNil() bool {
	return brc == nil
}
