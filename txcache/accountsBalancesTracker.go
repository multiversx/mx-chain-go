package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-storage-go/types"
)

type accountsBalancesTracker struct {
	recordByAddress map[string]*accountBalanceRecord
}

type accountBalanceRecord struct {
	initialBalance  *big.Int
	consumedBalance *big.Int
}

func newAccountsBalancesTracker() *accountsBalancesTracker {
	return &accountsBalancesTracker{
		recordByAddress: make(map[string]*accountBalanceRecord),
	}
}

func (tracker *accountsBalancesTracker) setupAccount(address []byte, state *types.AccountState) {
	tracker.recordByAddress[string(address)] = &accountBalanceRecord{
		initialBalance:  state.Balance,
		consumedBalance: big.NewInt(0),
	}
}
