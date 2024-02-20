package accounts

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"math/big"
)

var _ state.UserAccountHandler = (*sovereignAccount)(nil)

type sovereignAccount struct {
	*userAccount
	esdtBalance state.ESDTAsBalanceHandler
}

// NewSovereignAccount creates a user account for sovereign shards
func NewSovereignAccount(
	address []byte,
	trackableDataTrie state.DataTrieTracker,
	trieLeafParser common.TrieLeafParser,
	esdtBalance state.ESDTAsBalanceHandler,
) (*sovereignAccount, error) {
	if len(address) == 0 {
		return nil, errors.ErrNilAddress
	}
	if check.IfNil(trackableDataTrie) {
		return nil, errors.ErrNilTrackableDataTrie
	}
	if check.IfNil(trieLeafParser) {
		return nil, errors.ErrNilTrieLeafParser
	}

	userAcc := &userAccount{
		UserAccountData: UserAccountData{
			DeveloperReward: big.NewInt(0),
			Balance:         big.NewInt(0),
			Address:         address,
		},
		dataTrieInteractor: trackableDataTrie,
		dataTrieLeafParser: trieLeafParser,
	}

	sovereignAcc := &sovereignAccount{userAcc, esdtBalance}
	return sovereignAcc, nil
}

// AddToBalance adds new value to balance
func (s *sovereignAccount) AddToBalance(value *big.Int) error {
	return s.esdtBalance.AddToBalance(s.dataTrieInteractor, value)
}

// SubFromBalance subtracts new value from balance
func (s *sovereignAccount) SubFromBalance(value *big.Int) error {
	return s.esdtBalance.SubFromBalance(s.dataTrieInteractor, value)
}

// GetBalance returns the actual balance from the account
func (s *sovereignAccount) GetBalance() *big.Int {
	return s.esdtBalance.GetBalance(s.dataTrieInteractor)
}
