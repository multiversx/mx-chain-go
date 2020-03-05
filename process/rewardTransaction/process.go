package rewardTransaction

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type rewardTxProcessor struct {
	accounts         state.AccountsAdapter
	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator
}

// NewRewardTxProcessor creates a rewardTxProcessor instance
func NewRewardTxProcessor(
	accountsDB state.AccountsAdapter,
	adrConv state.AddressConverter,
	coordinator sharding.Coordinator,
) (*rewardTxProcessor, error) {
	if check.IfNil(accountsDB) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(adrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &rewardTxProcessor{
		accounts:         accountsDB,
		adrConv:          adrConv,
		shardCoordinator: coordinator,
	}, nil
}

func (rtp *rewardTxProcessor) getAccountFromAddress(address []byte) (state.UserAccountHandler, error) {
	addr, err := rtp.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	shardForCurrentNode := rtp.shardCoordinator.SelfId()
	shardForAddr := rtp.shardCoordinator.ComputeId(addr)
	if shardForCurrentNode != shardForAddr {
		return nil, nil
	}

	acnt, err := rtp.accounts.LoadAccount(addr)
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

	process.DisplayProcessTxDetails("ProcessRewardTransaction: receiver account details", accHandler, rTx)

	err = accHandler.AddToBalance(rTx.Value)
	return rtp.accounts.SaveAccount(accHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtp *rewardTxProcessor) IsInterfaceNil() bool {
	return rtp == nil
}
