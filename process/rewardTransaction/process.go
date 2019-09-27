package rewardTransaction

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type rewardTxProcessor struct {
	accounts         state.AccountsAdapter
	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator

	mutRewardsForwarder sync.Mutex
	rewardTxForwarder   process.IntermediateTransactionHandler
}

// NewRewardTxProcessor creates a rewardTxProcessor instance
func NewRewardTxProcessor(
	accountsDB state.AccountsAdapter,
	adrConv state.AddressConverter,
	coordinator sharding.Coordinator,
	rewardTxForwarder process.IntermediateTransactionHandler,
) (*rewardTxProcessor, error) {
	if accountsDB == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if adrConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if coordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if rewardTxForwarder == nil {
		return nil, process.ErrNilIntermediateTransactionHandler
	}

	return &rewardTxProcessor{
		accounts:          accountsDB,
		adrConv:           adrConv,
		shardCoordinator:  coordinator,
		rewardTxForwarder: rewardTxForwarder,
	}, nil
}

func (rtp *rewardTxProcessor) getAccountFromAddress(address []byte) (state.AccountHandler, error) {
	addr, err := rtp.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	shardForCurrentNode := rtp.shardCoordinator.SelfId()
	shardForAddr := rtp.shardCoordinator.ComputeId(addr)
	if shardForCurrentNode != shardForAddr {
		return nil, nil
	}

	acnt, err := rtp.accounts.GetAccountWithJournal(addr)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

// ProcessRewardTransaction updates the account state from the reward transaction
func (rtp *rewardTxProcessor) ProcessRewardTransaction(rTx *rewardTx.RewardTx) error {
	if rTx == nil {
		return process.ErrNilRewardTransaction
	}
	if rTx.Value == nil {
		return process.ErrNilValueFromRewardTransaction
	}

	rtp.mutRewardsForwarder.Lock()
	err := rtp.rewardTxForwarder.AddIntermediateTransactions([]data.TransactionHandler{rTx})
	rtp.mutRewardsForwarder.Unlock()
	if err != nil {
		return err
	}

	accHandler, err := rtp.getAccountFromAddress(rTx.RcvAddr)
	if err != nil {
		return err
	}

	if accHandler == nil || accHandler.IsInterfaceNil() {
		// address from different shard
		return nil
	}

	rewardAcc, ok := accHandler.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	operation := big.NewInt(0)
	operation = operation.Add(rTx.Value, rewardAcc.Balance)
	err = rewardAcc.SetBalanceWithJournal(operation)

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtp *rewardTxProcessor) IsInterfaceNil() bool {
	if rtp == nil {
		return true
	}
	return false
}
