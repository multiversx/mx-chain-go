package state

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
)

// Account is the struct used in serialization/deserialization
type Account struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash []byte
	RootHash []byte
	Address  []byte

	DeveloperReward *big.Int
	OwnerAddress    []byte

	addressContainer AddressContainer
	code             []byte
	accountTracker   AccountTracker
	dataTrieTracker  DataTrieTracker
}

// NewAccount creates new simple account wrapper for an AccountContainer (that has just been initialized)
func NewAccount(addressContainer AddressContainer, tracker AccountTracker) (*Account, error) {
	if check.IfNil(addressContainer) {
		return nil, ErrNilAddressContainer
	}
	if check.IfNil(tracker) {
		return nil, ErrNilAccountTracker
	}

	addressBytes := addressContainer.Bytes()

	return &Account{
		DeveloperReward:  big.NewInt(0),
		Balance:          big.NewInt(0),
		addressContainer: addressContainer,
		Address:          addressBytes,
		accountTracker:   tracker,
		dataTrieTracker:  NewTrackableDataTrie(addressBytes, nil),
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *Account) IsInterfaceNil() bool {
	return a == nil
}

// AddressContainer returns the address associated with the account
func (a *Account) AddressContainer() AddressContainer {
	return a.addressContainer
}

// SetNonceWithJournal sets the account's nonce, saving the old nonce before changing
func (a *Account) SetNonceWithJournal(nonce uint64) error {
	entry, err := NewBaseJournalEntryNonce(a, a.Nonce)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Nonce = nonce

	return a.accountTracker.SaveAccount(a)
}

//SetNonce saves the nonce to the account
func (a *Account) SetNonce(nonce uint64) {
	a.Nonce = nonce
}

// GetNonce gets the nonce of the account
func (a *Account) GetNonce() uint64 {
	return a.Nonce
}

// SetBalanceWithJournal sets the account's balance, saving the old balance before changing
func (a *Account) SetBalanceWithJournal(balance *big.Int) error {
	entry, err := NewJournalEntryBalance(a, a.Balance)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Balance = balance

	return a.accountTracker.SaveAccount(a)
}

func (a *Account) setDeveloperRewardWithJournal(developerReward *big.Int) error {
	entry, err := NewJournalEntryDeveloperReward(a, a.DeveloperReward)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.DeveloperReward = developerReward

	return a.accountTracker.SaveAccount(a)
}

// SetOwnerAddressWithJournal sets the owner address of an account, saving the old before changing
func (a *Account) SetOwnerAddressWithJournal(ownerAddress []byte) error {
	entry, err := NewJournalEntryOwnerAddress(a, a.OwnerAddress)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.OwnerAddress = ownerAddress

	return a.accountTracker.SaveAccount(a)
}

//------- code / code hash

// GetCodeHash returns the code hash associated with this account
func (a *Account) GetCodeHash() []byte {
	return a.CodeHash
}

// SetCodeHash sets the code hash associated with the account
func (a *Account) SetCodeHash(codeHash []byte) {
	a.CodeHash = codeHash
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (a *Account) SetCodeHashWithJournal(codeHash []byte) error {
	entry, err := NewBaseJournalEntryCodeHash(a, a.CodeHash)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.CodeHash = codeHash

	return a.accountTracker.SaveAccount(a)
}

// GetCode gets the actual code that needs to be run in the VM
func (a *Account) GetCode() []byte {
	return a.code
}

// SetCode sets the actual code that needs to be run in the VM
func (a *Account) SetCode(code []byte) {
	a.code = code
}

//------- data trie / root hash

// GetRootHash returns the root hash associated with this account
func (a *Account) GetRootHash() []byte {
	return a.RootHash
}

// SetRootHash sets the root hash associated with the account
func (a *Account) SetRootHash(roothash []byte) {
	a.RootHash = roothash
}

// DataTrie returns the trie that holds the current account's data
func (a *Account) DataTrie() data.Trie {
	return a.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (a *Account) SetDataTrie(trie data.Trie) {
	a.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (a *Account) DataTrieTracker() DataTrieTracker {
	return a.dataTrieTracker
}

// ClaimDeveloperRewards returns the accumulated developer rewards and sets it to 0 in the account
func (a *Account) ClaimDeveloperRewards(sndAddress []byte) (*big.Int, error) {
	if !bytes.Equal(sndAddress, a.OwnerAddress) {
		return nil, ErrOperationNotPermitted
	}

	oldValue := big.NewInt(0).Set(a.DeveloperReward)
	err := a.setDeveloperRewardWithJournal(big.NewInt(0))
	if err != nil {
		return nil, err
	}

	return oldValue, nil
}

// ChangeOwnerAddress changes the owner account if operation is permitted
func (a *Account) ChangeOwnerAddress(sndAddress []byte, newAddress []byte) error {
	if !bytes.Equal(sndAddress, a.OwnerAddress) {
		return ErrOperationNotPermitted
	}
	if len(newAddress) != len(a.addressContainer.Bytes()) {
		return ErrInvalidAddressLength
	}

	err := a.SetOwnerAddressWithJournal(newAddress)
	if err != nil {
		return err
	}

	return nil
}

// AddToDeveloperReward adds new value to developer reward
func (a *Account) AddToDeveloperReward(value *big.Int) error {
	newDeveloperReward := big.NewInt(0).Add(a.DeveloperReward, value)
	err := a.setDeveloperRewardWithJournal(newDeveloperReward)
	if err != nil {
		log.Debug("SetDeveloperRewardWithJournal error", "error", err.Error())
		return nil
	}

	return nil
}

// AddToBalance adds new value to balance
func (a *Account) AddToBalance(value *big.Int) error {
	newBalance := big.NewInt(0).Add(a.Balance, value)
	if newBalance.Cmp(big.NewInt(0)) < 0 {
		return ErrInsufficientFunds
	}

	err := a.SetBalanceWithJournal(newBalance)
	if err != nil {
		log.Debug("SetDeveloperRewardWithJournal error", "error", err.Error())
		return nil
	}

	return nil
}

// GetOwnerAddress returns the owner address
func (a *Account) GetOwnerAddress() []byte {
	return a.OwnerAddress
}

// GetBalance returns the actual balance from the account
func (a *Account) GetBalance() *big.Int {
	return big.NewInt(0).Set(a.Balance)
}

//TODO add Cap'N'Proto converter funcs
