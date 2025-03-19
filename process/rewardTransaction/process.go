package rewardTransaction

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.RewardTransactionProcessor = (*rewardTxProcessor)(nil)

var log = logger.GetOrCreate("process/reward")

const rewardKey = "reward"

type rewardTxProcessor struct {
	accounts         state.AccountsAdapter
	pubkeyConv       core.PubkeyConverter
	shardCoordinator sharding.Coordinator
	marshaller       marshal.Marshalizer
	hasher           hashing.Hasher
}

// NewRewardTxProcessor creates a rewardTxProcessor instance
func NewRewardTxProcessor(
	accountsDB state.AccountsAdapter,
	pubkeyConv core.PubkeyConverter,
	coordinator sharding.Coordinator,
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
) (*rewardTxProcessor, error) {
	if check.IfNil(accountsDB) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(pubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}

	return &rewardTxProcessor{
		accounts:         accountsDB,
		pubkeyConv:       pubkeyConv,
		shardCoordinator: coordinator,
		marshaller:       marshaller,
		hasher:           hasher,
	}, nil
}

func (rtp *rewardTxProcessor) getAccountFromAddress(address []byte) (state.UserAccountHandler, error) {
	shardForCurrentNode := rtp.shardCoordinator.SelfId()
	shardForAddr := rtp.shardCoordinator.ComputeId(address)
	if shardForCurrentNode != shardForAddr {
		return nil, nil
	}

	acnt, err := rtp.accounts.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	userAcnt, ok := acnt.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return userAcnt, nil
}

// ProcessRewardTransaction updates the account state from the reward transaction
func (rtp *rewardTxProcessor) ProcessRewardTransaction(rTx *rewardTx.RewardTx) error {
	if rTx == nil {
		return process.ErrNilRewardTransaction
	}
	if rTx.Value == nil {
		return process.ErrNilValueFromRewardTransaction
	}

	accHandler, err := rtp.getAccountFromAddress(rTx.RcvAddr)
	if err != nil {
		return err
	}

	if check.IfNil(accHandler) {
		// address from different shard
		return nil
	}

	process.DisplayProcessTxDetails(
		"ProcessRewardTransaction: receiver account details",
		accHandler,
		rTx,
		nil,
		rtp.pubkeyConv,
	)

	txHash, err := core.CalculateHash(rtp.marshaller, rtp.hasher, rTx)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return err
	}

	// TODO refactor to set the tx hash for the following state changes before the processing occurs
	defer func() {
		rtp.accounts.SetTxHashForLatestStateChanges(txHash, rTx)
		log.Debug("SetTxHashForLatestStateChanges", "txHash", txHash)
	}()

	err = accHandler.AddToBalance(rTx.Value)
	if err != nil {
		return err
	}

	err = rtp.saveAccumulatedRewards(rTx, accHandler)
	if err != nil {
		return err
	}

	return rtp.accounts.SaveAccount(accHandler)
}

func (rtp *rewardTxProcessor) saveAccumulatedRewards(
	rtx *rewardTx.RewardTx,
	userAccount state.UserAccountHandler,
) error {
	if !core.IsSmartContractAddress(rtx.RcvAddr) {
		return nil
	}

	existingReward := big.NewInt(0)
	fullRewardKey := core.ProtectedKeyPrefix + rewardKey
	val, _, err := userAccount.RetrieveValue([]byte(fullRewardKey))
	if err == nil {
		existingReward.SetBytes(val)
	}

	if core.IsGetNodeFromDBError(err) {
		return err
	}

	existingReward.Add(existingReward, rtx.Value)
	_ = userAccount.SaveKeyValue([]byte(fullRewardKey), existingReward.Bytes())

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtp *rewardTxProcessor) IsInterfaceNil() bool {
	return rtp == nil
}
