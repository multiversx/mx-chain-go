package preprocess

import (
    "math/big"
    "sync"

    "github.com/ElrondNetwork/elrond-go/core"
    "github.com/ElrondNetwork/elrond-go/data"
    "github.com/ElrondNetwork/elrond-go/data/block"
    "github.com/ElrondNetwork/elrond-go/data/rewardTx"
    "github.com/ElrondNetwork/elrond-go/hashing"
    "github.com/ElrondNetwork/elrond-go/marshal"
    "github.com/ElrondNetwork/elrond-go/process"
    "github.com/ElrondNetwork/elrond-go/sharding"
)

const communityPercentage = 0.1 // 1 = 100%, 0 = 0%
const leaderPercentage = 0.4    // 1 = 100%, 0 = 0%
const burnPercentage = 0.5      // 1 = 100%, 0 = 0%

// TODO: Replace with valid reward value
var rewardValue = big.NewInt(1000)

type rewardsHandler struct {
    address          process.SpecialAddressHandler
    shardCoordinator sharding.Coordinator
    hasher           hashing.Hasher
    marshalizer      marshal.Marshalizer
    mut              sync.Mutex
    accumulatedFees  *big.Int

    rewardTxsFromBlock map[string]*rewardTx.RewardTx
}

// NewRewardTxHandler constructor for the reward transaction handler
func NewRewardTxHandler(
    address process.SpecialAddressHandler,
    shardCoordinator sharding.Coordinator,
    hasher hashing.Hasher,
    marshalizer marshal.Marshalizer,
) (*rewardsHandler, error) {
    if address == nil || address.IsInterfaceNil() {
        return nil, process.ErrNilSpecialAddressHandler
    }
    if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
        return nil, process.ErrNilShardCoordinator
    }
    if hasher == nil || hasher.IsInterfaceNil() {
        return nil, process.ErrNilHasher
    }
    if marshalizer == nil || marshalizer.IsInterfaceNil() {
        return nil, process.ErrNilMarshalizer
    }

    rtxh := &rewardsHandler{
        address:          address,
        shardCoordinator: shardCoordinator,
        hasher:           hasher,
        marshalizer:      marshalizer,
    }
    rtxh.accumulatedFees = big.NewInt(0)
    rtxh.rewardTxsFromBlock = make(map[string]*rewardTx.RewardTx)

    return rtxh, nil
}

// SaveCurrentIntermediateTxToStorage saves current cached data into storage - already saved for txs
func (rtxh *rewardsHandler) SaveCurrentIntermediateTxToStorage() error {
    //TODO implement me - save only created accumulatedFees
    return nil
}

// AddIntermediateTransactions adds intermediate transactions to local cache
func (rtxh *rewardsHandler) AddIntermediateTransactions(txs []data.TransactionHandler) error {
    return nil
}

// CreateAllInterMiniBlocks creates miniblocks from process transactions
func (rtxh *rewardsHandler) CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock {
    calculatedRewardTxs := rtxh.CreateAllUTxs()

    miniBlocks := make(map[uint32]*block.MiniBlock)
    for _, rTx := range calculatedRewardTxs {
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
                SenderShardID:   rtxh.shardCoordinator.SelfId(),
                Type:            block.RewardsBlock,
            }
        }

        mb.TxHashes = append(mb.TxHashes, txHash)
        miniBlocks[dstShId] = mb
    }

    return miniBlocks
}

// VerifyInterMiniBlocks verifies if transaction fees were correctly handled for the block
func (rtxh *rewardsHandler) VerifyInterMiniBlocks(body block.Body) error {
    err := rtxh.VerifyCreatedUTxs()
    rtxh.CleanProcessedUTxs()

    return err
}

// CreateBlockStarted does the cleanup before creating a new block
func (rtxh *rewardsHandler) CreateBlockStarted() {
    rtxh.CleanProcessedUTxs()
}

// CleanProcessedUTxs deletes the cached data
func (rtxh *rewardsHandler) CleanProcessedUTxs() {
    rtxh.mut.Lock()
    rtxh.accumulatedFees = big.NewInt(0)
    rtxh.rewardTxsFromBlock = make(map[string]*rewardTx.RewardTx)
    rtxh.mut.Unlock()
}

// AddRewardTxFromBlock adds an existing reward transaction from block into local cache
func (rtxh *rewardsHandler) AddRewardTxFromBlock(tx data.TransactionHandler) {
    currRewardTx, ok := tx.(*rewardTx.RewardTx)
    if !ok {
        log.Error(process.ErrWrongTypeAssertion.Error())
        return
    }

    rtxh.mut.Lock()
    rtxh.rewardTxsFromBlock[string(tx.GetRecvAddress())] = currRewardTx
    rtxh.mut.Unlock()
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

    return currTx
}

func (rtxh *rewardsHandler) createBurnTx() *rewardTx.RewardTx {
    currTx := &rewardTx.RewardTx{}

    currTx.Value = getPercentageOfValue(rtxh.accumulatedFees, burnPercentage)
    currTx.RcvAddr = rtxh.address.BurnAddress()

    return currTx
}

func (rtxh *rewardsHandler) createCommunityTx() *rewardTx.RewardTx {
    currTx := &rewardTx.RewardTx{}

    currTx.Value = getPercentageOfValue(rtxh.accumulatedFees, communityPercentage)
    currTx.RcvAddr = rtxh.address.ElrondCommunityAddress()

    return currTx
}

// CreateAllUTxs creates all the needed reward transactions
// According to economic paper, out of the block fees 50% are burned, 40% go to the leader and 10% go
// to Elrond community fund. Fixed rewards for every validator in consensus are generated by the system.
func (rtxh *rewardsHandler) CreateAllUTxs() []data.TransactionHandler {

    rewardTxs := make([]data.TransactionHandler, 0)
    rewardsFromFees := rtxh.createRewardTxsFromFee()
    rewardsForConsensus := rtxh.createRewardTxsForConsensusGroup()

    rewardTxs = append(rewardTxs, rewardsFromFees...)
    rewardTxs = append(rewardTxs, rewardsForConsensus...)

    return rewardTxs
}

func (rtxh *rewardsHandler) createRewardTxsFromFee() []data.TransactionHandler {
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

    rtxh.accumulatedFees = big.NewInt(0)

    return currFeeTxs
}

func (rtxh *rewardsHandler) createRewardTxsForConsensusGroup() []data.TransactionHandler {
    consensusRewardAddresses := rtxh.address.ConsensusRewardAddresses()

    consensusRewardTxs := make([]data.TransactionHandler, 0)
    for _, address := range consensusRewardAddresses {
        rTx := &rewardTx.RewardTx{}
        rTx.Value = rewardValue
        rTx.RcvAddr = []byte(address)

        consensusRewardTxs = append(consensusRewardTxs, rTx)
    }

    return consensusRewardTxs
}

// VerifyCreatedUTxs creates all fee txs from added values, than verifies if in block the values are the same
func (rtxh *rewardsHandler) VerifyCreatedUTxs() error {
    calculatedFeeTxs := rtxh.CreateAllUTxs()

    rtxh.mut.Lock()
    defer rtxh.mut.Unlock()

    totalFeesFromBlock := big.NewInt(0)
    for _, value := range rtxh.rewardTxsFromBlock {
        totalFeesFromBlock = totalFeesFromBlock.Add(totalFeesFromBlock, value.Value)
    }

    totalCalculatedFees := big.NewInt(0)
    for _, value := range calculatedFeeTxs {
        totalCalculatedFees = totalCalculatedFees.Add(totalCalculatedFees, value.GetValue())

        txFromBlock, ok := rtxh.rewardTxsFromBlock[string(value.GetRecvAddress())]
        if !ok {
            return process.ErrTxsFeesNotFound
        }
        if txFromBlock.Value.Cmp(value.GetValue()) != 0 {
            return process.ErrTxsFeesDoNotMatch
        }
    }

    if totalCalculatedFees.Cmp(totalFeesFromBlock) != 0 {
        return process.ErrTotalTxsFeesDoNotMatch
    }

    return nil
}

// CreateMarshalizedData creates the marshalized data for broadcasting purposes
func (rtxh *rewardsHandler) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
    // TODO: implement me

    return make([][]byte, 0), nil
}

// GetAllCurrentFinishedTxs returns the cached finalized transactions for current round
func (rtxh *rewardsHandler) GetAllCurrentFinishedTxs() map[string]data.TransactionHandler {
    rtxh.mut.Lock()

    rewardTxPool := make(map[string]data.TransactionHandler)
    for txHash, txInfo := range rtxh.rewardTxsFromBlock {

        senderShard := txInfo.ShardId
        receiverShard, err := rtxh.address.ShardIdForAddress(txInfo.RcvAddr)
        if err != nil {
            continue
        }
        if receiverShard != rtxh.shardCoordinator.SelfId() {
            continue
        }
        if senderShard != rtxh.shardCoordinator.SelfId() {
            continue
        }
        rewardTxPool[txHash] = txInfo
    }
    rtxh.mut.Unlock()

    return rewardTxPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtxh *rewardsHandler) IsInterfaceNil() bool {
    if rtxh == nil {
        return true
    }
    return false
}
