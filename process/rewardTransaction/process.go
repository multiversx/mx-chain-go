package rewardTransaction

import (
	"math/big"

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

func NewRewardTxProcessor(
	accountsDB state.AccountsAdapter,
	adrConv state.AddressConverter,
	coordinator sharding.Coordinator,
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

	return &rewardTxProcessor{
		accounts:         accountsDB,
		adrConv:          adrConv,
		shardCoordinator: coordinator,
	}, nil
}

func (rtp *rewardTxProcessor) getAccountFromAddress(address []byte) (state.AccountHandler, error) {
	adrSrc, err := rtp.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	shardForCurrentNode := rtp.shardCoordinator.SelfId()
	shardForSrc := rtp.shardCoordinator.ComputeId(adrSrc)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := rtp.accounts.GetAccountWithJournal(adrSrc)
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

	accHandler, err := rtp.getAccountFromAddress(rTx.RcvAddr)
	if err != nil {
		return err
	}
	if accHandler == nil || accHandler.IsInterfaceNil() {
		return process.ErrNilSCDestAccount
	}

	rewardAcc, ok := accHandler.(*state.Account)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if rTx.Value == nil {
		return process.ErrNilValueFromRewardTransaction
	}

	operation := big.NewInt(0)
	operation = operation.Add(rTx.Value, rewardAcc.Balance)
	err = rewardAcc.SetBalanceWithJournal(operation)
	if err != nil {
		return err
	}

	return nil
}
