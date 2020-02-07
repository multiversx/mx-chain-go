package preprocess

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type rewardsHandler struct {
	address          process.SpecialAddressHandler
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	adrConv          state.AddressConverter
	store            dataRetriever.StorageService
	rewardTxPool     dataRetriever.ShardedDataCacherNotifier

	mutGenRewardTxs     sync.RWMutex
	protocolRewards     []data.TransactionHandler
	protocolRewardsMeta []data.TransactionHandler
	feeRewards          []data.TransactionHandler

	mut                 sync.Mutex
	accumulatedFees     *big.Int
	rewardTxsForBlock   map[string]*rewardTx.RewardTx
	economicsRewards    process.RewardsHandler
	intraShardMiniBlock *block.MiniBlock
}

// NewRewardTxHandler constructor for the reward transaction handler
func NewRewardTxHandler(
	address process.SpecialAddressHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	adrConv state.AddressConverter,
	store dataRetriever.StorageService,
	rewardTxPool dataRetriever.ShardedDataCacherNotifier,
	economicsRewards process.RewardsHandler,
) (*rewardsHandler, error) {
	if address == nil || address.IsInterfaceNil() {
		return nil, process.ErrNilSpecialAddressHandler
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if adrConv == nil || adrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if store == nil || store.IsInterfaceNil() {
		return nil, process.ErrNilStorage
	}
	if rewardTxPool == nil || rewardTxPool.IsInterfaceNil() {
		return nil, process.ErrNilRewardTxDataPool
	}
	if economicsRewards == nil || economicsRewards.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsRewardsHandler
	}

	rtxh := &rewardsHandler{
		address:          address,
		shardCoordinator: shardCoordinator,
		adrConv:          adrConv,
		hasher:           hasher,
		marshalizer:      marshalizer,
		store:            store,
		rewardTxPool:     rewardTxPool,
		economicsRewards: economicsRewards,
	}

	rtxh.accumulatedFees = big.NewInt(0)
	rtxh.rewardTxsForBlock = make(map[string]*rewardTx.RewardTx)

	return rtxh, nil
}

// SaveCurrentIntermediateTxToStorage saves current cached data into storage - already saved for txs
func (rtxh *rewardsHandler) SaveCurrentIntermediateTxToStorage() error {
	rtxh.mut.Lock()
	defer rtxh.mut.Unlock()

	for _, rTx := range rtxh.rewardTxsForBlock {
		buff, err := rtxh.marshalizer.Marshal(rTx)
		if err != nil {
			return err
		}

		errNotCritical := rtxh.store.Put(dataRetriever.RewardTransactionUnit, rtxh.hasher.Compute(string(buff)), buff)
		if errNotCritical != nil {
			log.Debug("RewardTransactionUnit.Put", "error", errNotCritical.Error())
		}
	}

	return nil
}

// AddIntermediateTransactions adds intermediate transactions to local cache
func (rtxh *rewardsHandler) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	rtxh.mut.Lock()
	defer rtxh.mut.Unlock()

	for i := 0; i < len(txs); i++ {
		addedRewardTx, ok := txs[i].(*rewardTx.RewardTx)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		if addedRewardTx.ShardId != rtxh.shardCoordinator.SelfId() {
			continue
		}

		rewardTxHash, err := core.CalculateHash(rtxh.marshalizer, rtxh.hasher, txs[i])
		if err != nil {
			return err
		}

		rtxh.rewardTxsForBlock[string(rewardTxHash)] = addedRewardTx
	}

	return nil
}

// CreateAllInterMiniBlocks creates miniblocks from process transactions
func (rtxh *rewardsHandler) CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock {
	rtxh.mutGenRewardTxs.Lock()

	log.Trace("total accumulated fees", "value", rtxh.accumulatedFees)

	rtxh.feeRewards = rtxh.createRewardFromFees()
	rtxh.addTransactionsToPool(rtxh.feeRewards)

	rtxh.protocolRewards = rtxh.createProtocolRewards()
	rtxh.addTransactionsToPool(rtxh.protocolRewards)

	rtxh.protocolRewardsMeta = rtxh.createProtocolRewardsForMeta()
	rtxh.addTransactionsToPool(rtxh.protocolRewardsMeta)

	calculatedRewardTxsLen := len(rtxh.protocolRewards) + len(rtxh.protocolRewardsMeta) + len(rtxh.feeRewards)
	calculatedRewardTxs := make([]data.TransactionHandler, 0, calculatedRewardTxsLen)
	calculatedRewardTxs = append(calculatedRewardTxs, rtxh.protocolRewards...)
	calculatedRewardTxs = append(calculatedRewardTxs, rtxh.protocolRewardsMeta...)
	calculatedRewardTxs = append(calculatedRewardTxs, rtxh.feeRewards...)

	rtxh.displayCalculatedRewardTxs()

	rtxh.mutGenRewardTxs.Unlock()

	miniBlocks := rtxh.miniblocksFromRewardTxs(calculatedRewardTxs)

	if _, ok := miniBlocks[rtxh.shardCoordinator.SelfId()]; ok {
		rtxh.intraShardMiniBlock = miniBlocks[rtxh.shardCoordinator.SelfId()].Clone()
	}

	return miniBlocks
}

func (rtxh *rewardsHandler) addTransactionsToPool(rewardTxs []data.TransactionHandler) {
	for _, rTx := range rewardTxs {
		dstShId, err := rtxh.address.ShardIdForAddress(rTx.GetRecvAddress())
		if err != nil {
			log.Trace("ShardIdForAddress", "error", err.Error())
		}

		txHash, err := core.CalculateHash(rtxh.marshalizer, rtxh.hasher, rTx)
		if err != nil {
			log.Trace("CalculateHash", "error", err.Error())
		}

		// add the reward transaction to the the pool so that the processor can find it
		cacheId := process.ShardCacherIdentifier(rtxh.shardCoordinator.SelfId(), dstShId)
		rtxh.rewardTxPool.AddData(txHash, rTx, cacheId)
	}
}

func (rtxh *rewardsHandler) miniblocksFromRewardTxs(
	rewardTxs []data.TransactionHandler,
) map[uint32]*block.MiniBlock {
	miniBlocks := make(map[uint32]*block.MiniBlock)

	for _, rTx := range rewardTxs {
		dstShId, err := rtxh.address.ShardIdForAddress(rTx.GetRecvAddress())
		if err != nil {
			log.Trace("ShardIdForAddress", "error", err.Error())
			continue
		}

		txHash, err := core.CalculateHash(rtxh.marshalizer, rtxh.hasher, rTx)
		if err != nil {
			log.Trace("CalculateHash", "error", err.Error())
			continue
		}

		var ok bool
		var mb *block.MiniBlock
		if mb, ok = miniBlocks[dstShId]; !ok {
			mb = &block.MiniBlock{
				ReceiverShardID: dstShId,
				SenderShardID:   rtxh.shardCoordinator.SelfId(),
				Type:            block.RewardsBlock,
			}
		}

		mb.TxHashes = append(mb.TxHashes, txHash)
		miniBlocks[dstShId] = mb
	}

	return miniBlocks
}

// GetCreatedInShardMiniBlock will return a clone of the intra shard mini block
func (rtxh *rewardsHandler) GetCreatedInShardMiniBlock() *block.MiniBlock {
	rtxh.mut.Lock()
	defer rtxh.mut.Unlock()

	if rtxh.intraShardMiniBlock == nil {
		return nil
	}

	return rtxh.intraShardMiniBlock.Clone()
}

// VerifyInterMiniBlocks verifies if transaction fees were correctly handled for the block
func (rtxh *rewardsHandler) VerifyInterMiniBlocks(_ block.Body) error {
	err := rtxh.verifyCreatedRewardsTxs()
	return err
}

// CreateBlockStarted does the cleanup before creating a new block
func (rtxh *rewardsHandler) CreateBlockStarted() {
	rtxh.cleanCachedData()
}

// ProcessTransactionFee adds the tx cost to the accumulated amount
func (rtxh *rewardsHandler) ProcessTransactionFee(cost *big.Int) {
	if cost == nil {
		log.Debug("nil cost in ProcessTransactionFee", "error", process.ErrNilValue.Error())
		return
	}

	rtxh.mut.Lock()
	_ = rtxh.accumulatedFees.Add(rtxh.accumulatedFees, cost)
	rtxh.mut.Unlock()
}

// cleanCachedData deletes the cached data
func (rtxh *rewardsHandler) cleanCachedData() {
	rtxh.mut.Lock()
	rtxh.accumulatedFees = big.NewInt(0)
	rtxh.rewardTxsForBlock = make(map[string]*rewardTx.RewardTx)
	rtxh.intraShardMiniBlock = nil
	rtxh.mut.Unlock()

	rtxh.mutGenRewardTxs.Lock()
	rtxh.feeRewards = make([]data.TransactionHandler, 0)
	rtxh.protocolRewards = make([]data.TransactionHandler, 0)
	rtxh.protocolRewardsMeta = make([]data.TransactionHandler, 0)
	rtxh.mutGenRewardTxs.Unlock()

}

func getPercentageOfValue(value *big.Int, percentage float64) *big.Int {
	x := new(big.Float).SetInt(value)
	y := big.NewFloat(percentage)

	z := new(big.Float).Mul(x, y)

	op := big.NewInt(0)
	result, _ := z.Int(op)

	return result
}

func (rtxh *rewardsHandler) createLeaderTx() *rewardTx.RewardTx {
	currTx := &rewardTx.RewardTx{}

	currTx.Value = getPercentageOfValue(rtxh.accumulatedFees, rtxh.economicsRewards.LeaderPercentage())
	currTx.RcvAddr = rtxh.address.LeaderAddress()
	currTx.ShardId = rtxh.shardCoordinator.SelfId()
	currTx.Epoch = rtxh.address.Epoch()
	currTx.Round = rtxh.address.Round()
	currTx.Type = rewardTx.LeaderTx

	return currTx
}

func (rtxh *rewardsHandler) createBurnTx() *rewardTx.RewardTx {
	currTx := &rewardTx.RewardTx{}

	currTx.Value = getPercentageOfValue(rtxh.accumulatedFees, rtxh.economicsRewards.BurnPercentage())
	currTx.RcvAddr = rtxh.address.BurnAddress()
	currTx.ShardId = rtxh.shardCoordinator.SelfId()
	currTx.Epoch = rtxh.address.Epoch()
	currTx.Round = rtxh.address.Round()
	currTx.Type = rewardTx.BurnTx

	return currTx
}

func (rtxh *rewardsHandler) createCommunityTx() *rewardTx.RewardTx {
	currTx := &rewardTx.RewardTx{}

	currTx.Value = getPercentageOfValue(rtxh.accumulatedFees, rtxh.economicsRewards.CommunityPercentage())
	currTx.RcvAddr = rtxh.address.ElrondCommunityAddress()
	currTx.ShardId = rtxh.shardCoordinator.SelfId()
	currTx.Epoch = rtxh.address.Epoch()
	currTx.Round = rtxh.address.Round()
	currTx.Type = rewardTx.CommunityTx

	return currTx
}

// createRewardFromFees creates the reward transactions from accumulated fees
// According to economic paper, out of the block fees 40% are burned, 50% go to the
// leader and 10% go to Elrond community fund.
func (rtxh *rewardsHandler) createRewardFromFees() []data.TransactionHandler {
	rtxh.mut.Lock()
	defer rtxh.mut.Unlock()

	if rtxh.accumulatedFees.Cmp(big.NewInt(1)) < 0 {
		rtxh.accumulatedFees = big.NewInt(0)
		return nil
	}

	leaderTx := rtxh.createLeaderTx()
	communityTx := rtxh.createCommunityTx()
	burnTx := rtxh.createBurnTx()

	currFeeTxs := make([]data.TransactionHandler, 0)
	currFeeTxs = append(currFeeTxs, leaderTx, communityTx, burnTx)

	return currFeeTxs
}

// createProtocolRewards creates the protocol reward transactions
func (rtxh *rewardsHandler) createProtocolRewards() []data.TransactionHandler {
	consensusRewardData := rtxh.address.ConsensusShardRewardData()

	isRewardValueZero := rtxh.economicsRewards.RewardsValue().Cmp(big.NewInt(0)) == 0
	if isRewardValueZero {
		return []data.TransactionHandler{}
	}

	consensusRewardTxs := make([]data.TransactionHandler, 0, len(consensusRewardData.Addresses))
	for index, address := range consensusRewardData.Addresses {
		rTx := &rewardTx.RewardTx{}
		rTx.Value = rtxh.economicsRewards.RewardsValue()
		rTx.RcvAddr = []byte(address)
		rTx.ShardId = rtxh.shardCoordinator.SelfId()
		rTx.Epoch = consensusRewardData.Epoch
		rTx.Round = consensusRewardData.Round
		rTx.CnsIndex = uint16(index)
		rTx.Type = rewardTx.ProtocolRewardsTx

		consensusRewardTxs = append(consensusRewardTxs, rTx)
	}

	return consensusRewardTxs
}

// createProtocolRewardsForMeta creates the protocol reward transactions
func (rtxh *rewardsHandler) createProtocolRewardsForMeta() []data.TransactionHandler {
	metaRewardsData := rtxh.address.ConsensusMetaRewardData()
	consensusRewardTxs := make([]data.TransactionHandler, 0)

	isRewardValueZero := rtxh.economicsRewards.RewardsValue().Cmp(big.NewInt(0)) == 0
	if isRewardValueZero {
		return consensusRewardTxs
	}

	for _, metaConsensusSet := range metaRewardsData {
		for index, address := range metaConsensusSet.Addresses {
			shardId, err := rtxh.address.ShardIdForAddress([]byte(address))
			if err != nil {
				log.Debug("ShardIdForAddress", "error", err.Error())
				continue
			}

			if shardId != rtxh.shardCoordinator.SelfId() {
				continue
			}

			rTx := &rewardTx.RewardTx{}
			rTx.Value = rtxh.economicsRewards.RewardsValue()
			rTx.RcvAddr = []byte(address)
			rTx.ShardId = rtxh.shardCoordinator.SelfId()
			rTx.Epoch = metaConsensusSet.Epoch
			rTx.Round = metaConsensusSet.Round
			rTx.CnsIndex = uint16(index)
			rTx.Type = rewardTx.ProtocolRewardsForMetaTx

			consensusRewardTxs = append(consensusRewardTxs, rTx)
		}
	}

	return consensusRewardTxs
}

// verifyCreatedRewardsTxs verifies if the calculated rewards transactions and the block reward transactions are the same
func (rtxh *rewardsHandler) verifyCreatedRewardsTxs() error {
	calculatedRewardTxs := make([]data.TransactionHandler, 0)
	rtxh.mutGenRewardTxs.RLock()
	calculatedRewardTxs = append(calculatedRewardTxs, rtxh.protocolRewards...)
	calculatedRewardTxs = append(calculatedRewardTxs, rtxh.protocolRewardsMeta...)
	calculatedRewardTxs = append(calculatedRewardTxs, rtxh.feeRewards...)
	rtxh.mutGenRewardTxs.RUnlock()

	rtxh.mut.Lock()
	defer rtxh.mut.Unlock()

	for _, tx := range rtxh.rewardTxsForBlock {
		log.Debug("rewardTxsForBlock",
			"shard", tx.ShardId,
			"epoch", tx.Epoch,
			"round", tx.Round,
			"rcvAddr", tx.RcvAddr,
			"value", tx.Value,
			"cnsIndex", tx.CnsIndex,
			"type", tx.Type)
	}

	totalFeesFromBlock := big.NewInt(0)
	for _, rTx := range rtxh.rewardTxsForBlock {
		totalFeesFromBlock = totalFeesFromBlock.Add(totalFeesFromBlock, rTx.GetValue())
	}

	if len(calculatedRewardTxs) != len(rtxh.rewardTxsForBlock) {
		return process.ErrRewardTxsMismatchCreatedReceived
	}

	totalCalculatedFees := big.NewInt(0)
	for _, rTx := range calculatedRewardTxs {
		totalCalculatedFees = totalCalculatedFees.Add(totalCalculatedFees, rTx.GetValue())

		rewardTxHash, err := core.CalculateHash(rtxh.marshalizer, rtxh.hasher, rTx)
		if err != nil {
			return err
		}

		txFromBlock, ok := rtxh.rewardTxsForBlock[string(rewardTxHash)]
		if !ok {
			return process.ErrRewardTxNotFound
		}
		if txFromBlock.GetValue().Cmp(rTx.GetValue()) != 0 {
			return process.ErrRewardTxsDoNotMatch
		}
	}

	return nil
}

// GetAllCurrentFinishedTxs returns the cached finalized transactions for current round
func (rtxh *rewardsHandler) GetAllCurrentFinishedTxs() map[string]data.TransactionHandler {
	rtxh.mut.Lock()

	rewardTxPool := make(map[string]data.TransactionHandler)
	for txHash, info := range rtxh.rewardTxsForBlock {

		senderShard := info.ShardId
		receiverShard, err := rtxh.address.ShardIdForAddress(info.RcvAddr)
		if err != nil {
			continue
		}
		if receiverShard != rtxh.shardCoordinator.SelfId() {
			continue
		}
		if senderShard != rtxh.shardCoordinator.SelfId() {
			continue
		}
		rewardTxPool[txHash] = info
	}
	rtxh.mut.Unlock()

	return rewardTxPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtxh *rewardsHandler) IsInterfaceNil() bool {
	return rtxh == nil
}

func (rtxh *rewardsHandler) displayCalculatedRewardTxs() {
	rtxh.displayTxs("protocolRewards", rtxh.protocolRewards)
	rtxh.displayTxs("protocolRewardsMeta", rtxh.protocolRewardsMeta)
	rtxh.displayTxs("feeRewards", rtxh.feeRewards)
}

func (rtxh *rewardsHandler) displayTxs(message string, txs []data.TransactionHandler) {
	for _, tx := range txs {
		rtx, ok := tx.(*rewardTx.RewardTx)
		if !ok {
			log.Debug("displayTxs", "error", process.ErrWrongTypeAssertion)
			continue
		}

		log.Debug(message,
			"shard", rtx.ShardId,
			"epoch", rtx.Epoch,
			"round", rtx.Round,
			"rcvAddr", rtx.RcvAddr,
			"value", rtx.Value,
			"cnsIndex", rtx.CnsIndex,
			"type", rtx.Type)
	}
}
