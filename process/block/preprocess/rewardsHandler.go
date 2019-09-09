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

// MinGasPrice is the minimal gas price to be paid for any transaction
// TODO: Set MinGasPrice and MinTxFee to some positive value (TBD)
var MinGasPrice = uint64(0)

// MinTxFee is the minimal fee to be paid for any transaction
var MinTxFee = uint64(0)

const communityPercentage = 0.1 // 1 = 100%, 0 = 0%
const leaderPercentage = 0.5    // 1 = 100%, 0 = 0%
const burnPercentage = 0.4      // 1 = 100%, 0 = 0%

// TODO: Replace with valid reward value
var rewardValue = big.NewInt(1000)

type rewardsHandler struct {
    address          process.SpecialAddressHandler
    hasher           hashing.Hasher
    marshalizer      marshal.Marshalizer
    shardCoordinator sharding.Coordinator
    adrConv          state.AddressConverter
    store            dataRetriever.StorageService
    rewardTxPool     dataRetriever.ShardedDataCacherNotifier

    mutGenRewardTxs sync.RWMutex
    protocolRewards []data.TransactionHandler
    feeRewards      []data.TransactionHandler

    mut               sync.Mutex
    accumulatedFees   *big.Int
    rewardTxsForBlock map[string]*rewardTx.RewardTx
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

    rtxh := &rewardsHandler{
        address:          address,
        shardCoordinator: shardCoordinator,
        adrConv:          adrConv,
        hasher:           hasher,
        marshalizer:      marshalizer,
        store:            store,
        rewardTxPool:     rewardTxPool,
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
            log.Error(errNotCritical.Error())
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

func (rtxh *rewardsHandler) getShardIdsFromAddress(addr []byte) (uint32, error) {
    address, err := rtxh.adrConv.CreateAddressFromPublicKeyBytes(addr)
    if err != nil {
        return rtxh.shardCoordinator.NumberOfShards(), err
    }
    shardId := rtxh.shardCoordinator.ComputeId(address)

    return shardId, nil
}

// CreateAllInterMiniBlocks creates miniblocks from process transactions
func (rtxh *rewardsHandler) CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock {
    rtxh.mutGenRewardTxs.Lock()
    calculatedRewardTxs := make([]data.TransactionHandler, 0)
    rtxh.feeRewards = rtxh.createRewardFromFees()
    rtxh.addTransactionsToPool(rtxh.feeRewards)
    calculatedRewardTxs = append(calculatedRewardTxs, rtxh.protocolRewards...)
    calculatedRewardTxs = append(calculatedRewardTxs, rtxh.feeRewards...)
    rtxh.mutGenRewardTxs.Unlock()

    miniBlocks := rtxh.miniblocksFromRewardTxs(calculatedRewardTxs)

    return miniBlocks
}

func (rtxh *rewardsHandler) addTransactionsToPool(rewardTxs []data.TransactionHandler) {
    for _, rTx := range rewardTxs {
        dstShId, err := rtxh.address.ShardIdForAddress(rTx.GetRecvAddress())
        if err != nil {
            log.Debug(err.Error())
        }

        txHash, err := core.CalculateHash(rtxh.marshalizer, rtxh.hasher, rTx)
        if err != nil {
            log.Debug(err.Error())
        }

        // add the reward transaction to the the pool so that the processor can find it
        cacheId := process.ShardCacherIdentifier(rtxh.shardCoordinator.SelfId(), dstShId)
        rtxh.rewardTxPool.AddData(txHash, rTx, cacheId)
    }
}

func (rtxh *rewardsHandler) miniblocksFromRewardTxs(
    rewardTxs []data.TransactionHandler,
) map[uint32]*block.MiniBlock {
    miniBlocks := make(map[uint32]*block.MiniBlock, 0)

    for _, rTx := range rewardTxs {
        dstShId, err := rtxh.address.ShardIdForAddress(rTx.GetRecvAddress())
        if err != nil {
            log.Debug(err.Error())
            continue
        }

        txHash, err := core.CalculateHash(rtxh.marshalizer, rtxh.hasher, rTx)
        if err != nil {
            log.Debug(err.Error())
            continue
        }

        var ok bool
        var mb *block.MiniBlock
        if mb, ok = miniBlocks[dstShId]; !ok {
            mb = &block.MiniBlock{
                ReceiverShardID: dstShId,
            }
        }

        mb.TxHashes = append(mb.TxHashes, txHash)
        miniBlocks[dstShId] = mb
    }

    return miniBlocks
}

// VerifyInterMiniBlocks verifies if transaction fees were correctly handled for the block
func (rtxh *rewardsHandler) VerifyInterMiniBlocks(body block.Body) error {
    err := rtxh.verifyCreatedRewardsTxs()
    return err
}

// CreateBlockStarted does the cleanup before creating a new block
func (rtxh *rewardsHandler) CreateBlockStarted() {
    rtxh.cleanCachedData()
    rewardTxs := rtxh.createProtocolRewards()
    rtxh.addTransactionsToPool(rewardTxs)
}

// CreateMarshalizedData creates the marshalized data for broadcasting purposes
func (rtxh *rewardsHandler) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
    rtxh.mut.Lock()
    defer rtxh.mut.Unlock()

    marshaledTxs := make([][]byte, 0)
    for _, txHash := range txHashes {
        rTx, ok := rtxh.rewardTxsForBlock[string(txHash)]
        if !ok {
            return nil, process.ErrRewardTxNotFound
        }

        marshaledTx, err := rtxh.marshalizer.Marshal(rTx)
        if err != nil {
            return nil, process.ErrMarshalWithoutSuccess
        }
        marshaledTxs = append(marshaledTxs, marshaledTx)
    }

    return marshaledTxs, nil
}

// ProcessTransactionFee adds the tx cost to the accumulated amount
func (rtxh *rewardsHandler) ProcessTransactionFee(cost *big.Int) {
    if cost == nil {
        log.Debug(process.ErrNilValue.Error())
        return
    }

    rtxh.mut.Lock()
    rtxh.accumulatedFees = rtxh.accumulatedFees.Add(rtxh.accumulatedFees, cost)
    rtxh.mut.Unlock()
}

// cleanCachedData deletes the cached data
func (rtxh *rewardsHandler) cleanCachedData() {
    rtxh.mut.Lock()
    rtxh.accumulatedFees = big.NewInt(0)
    rtxh.rewardTxsForBlock = make(map[string]*rewardTx.RewardTx)
    rtxh.mut.Unlock()

    rtxh.mutGenRewardTxs.Lock()
    rtxh.feeRewards = make([]data.TransactionHandler, 0)
    rtxh.protocolRewards = make([]data.TransactionHandler, 0)
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

    currTx.Value = getPercentageOfValue(rtxh.accumulatedFees, leaderPercentage)
    currTx.RcvAddr = rtxh.address.LeaderAddress()
    currTx.ShardId = rtxh.shardCoordinator.SelfId()
    currTx.Epoch = rtxh.address.Epoch()
    currTx.Round = rtxh.address.Round()

    return currTx
}

func (rtxh *rewardsHandler) createBurnTx() *rewardTx.RewardTx {
    currTx := &rewardTx.RewardTx{}

    currTx.Value = getPercentageOfValue(rtxh.accumulatedFees, burnPercentage)
    currTx.RcvAddr = rtxh.address.BurnAddress()
    currTx.ShardId = rtxh.shardCoordinator.SelfId()
    currTx.Epoch = rtxh.address.Epoch()
    currTx.Round = rtxh.address.Round()

    return currTx
}

func (rtxh *rewardsHandler) createCommunityTx() *rewardTx.RewardTx {
    currTx := &rewardTx.RewardTx{}

    currTx.Value = getPercentageOfValue(rtxh.accumulatedFees, communityPercentage)
    currTx.RcvAddr = rtxh.address.ElrondCommunityAddress()
    currTx.ShardId = rtxh.shardCoordinator.SelfId()
    currTx.Epoch = rtxh.address.Epoch()
    currTx.Round = rtxh.address.Round()

    return currTx
}

// createRewardFromFees creates the reward transactions from accumulated fees
// According to economic paper, out of the block fees 50% are burned, 40% go to the
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
    consensusRewardAddresses := rtxh.address.ConsensusRewardAddresses()

    consensusRewardTxs := make([]data.TransactionHandler, 0)
    for _, address := range consensusRewardAddresses {
        rTx := &rewardTx.RewardTx{}
        rTx.Value = rewardValue
        rTx.RcvAddr = []byte(address)
        rTx.ShardId = rtxh.shardCoordinator.SelfId()
        rTx.Epoch = rtxh.address.Epoch()
        rTx.Round = rtxh.address.Round()

        consensusRewardTxs = append(consensusRewardTxs, rTx)
    }

    rtxh.mutGenRewardTxs.Lock()
    rtxh.protocolRewards = consensusRewardTxs
    rtxh.mutGenRewardTxs.Unlock()

    return consensusRewardTxs
}

// VerifyCreatedRewardsTxs verifies if the calculated rewards transactions and the block reward transactions are the same
func (rtxh *rewardsHandler) verifyCreatedRewardsTxs() error {
    calculatedRewardTxs := make([]data.TransactionHandler, 0)
    rtxh.mutGenRewardTxs.RLock()
    calculatedRewardTxs = append(calculatedRewardTxs, rtxh.protocolRewards...)
    calculatedRewardTxs = append(calculatedRewardTxs, rtxh.feeRewards...)
    rtxh.mutGenRewardTxs.RUnlock()

    rtxh.mut.Lock()
    defer rtxh.mut.Unlock()

    totalFeesFromBlock := big.NewInt(0)
    for _, rTx := range rtxh.rewardTxsForBlock {
        totalFeesFromBlock = totalFeesFromBlock.Add(totalFeesFromBlock, rTx.GetValue())
    }

    totalCalculatedFees := big.NewInt(0)
    for _, value := range calculatedRewardTxs {
        totalCalculatedFees = totalCalculatedFees.Add(totalCalculatedFees, value.GetValue())

        rewardTxHash, err := core.CalculateHash(rtxh.marshalizer, rtxh.hasher, value)
        if err != nil {
            return err
        }

        txFromBlock, ok := rtxh.rewardTxsForBlock[string(rewardTxHash)]
        if !ok {
            return process.ErrRewardTxNotFound
        }
        if txFromBlock.GetValue().Cmp(value.GetValue()) != 0 {
            return process.ErrRewardTxsDoNotMatch
        }
    }

    if totalCalculatedFees.Cmp(totalFeesFromBlock) != 0 {
        return process.ErrTotalTxsFeesDoNotMatch
    }

    return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtxh *rewardsHandler) IsInterfaceNil() bool {
    if rtxh == nil {
        return true
    }
    return false
}
