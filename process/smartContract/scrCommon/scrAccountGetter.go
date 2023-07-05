package scrCommon

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

type scrAccountGetter struct {
	accounts         state.AccountsAdapter
	shardCoordinator sharding.Coordinator
}

func NewSCRAccountGetter(accAdapter state.AccountsAdapter, coordinator sharding.Coordinator) (AccountGetter, error) {
	if check.IfNil(accAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &scrAccountGetter{
		accounts:         accAdapter,
		shardCoordinator: coordinator,
	}, nil
}

func (sc *scrAccountGetter) GetAccountFromAddress(address []byte) (state.UserAccountHandler, error) {
	shardForCurrentNode := sc.shardCoordinator.SelfId()
	shardForSrc := sc.shardCoordinator.ComputeId(address)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := sc.accounts.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	stAcc, ok := acnt.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return stAcc, nil
}

func (sc *scrAccountGetter) IsInterfaceNil() bool {
	return sc == nil
}
