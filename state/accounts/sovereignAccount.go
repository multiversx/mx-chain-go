package accounts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
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
	if check.IfNil(esdtBalance) {
		return nil, errors.ErrNilESDTAsBalanceHandler
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

	return &sovereignAccount{
		userAccount: userAcc,
		esdtBalance: esdtBalance,
	}, nil
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

// IsInterfaceNil checks if the underlying pointer is nil
func (s *sovereignAccount) IsInterfaceNil() bool {
	return s == nil
}
