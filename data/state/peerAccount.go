package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// PeerAccount is the struct used in serialization/deserialization
type PeerAccount struct {
	BLSPublicKey     []byte
	SchnorrPublicKey []byte
	Address          []byte
	Stake            *big.Int

	JailTime      uint64
	PastJailTimes uint64

	CurrentShardId    uint32
	NextShardId       uint32
	NodeInWaitingList bool

	Uptime            uint64
	NrSignedBlocks    uint32
	NrMissedBlocks    uint32
	LeaderSuccessRate uint32

	Rating   uint32
	RootHash []byte
	Nonce    uint64

	addressContainer AddressContainer
	accountTracker   AccountTracker
	dataTrieTracker  DataTrieTracker
}

// NewPeerAccount creates new simple account wrapper for an PeerAccountContainer (that has just been initialized)
func NewPeerAccount(
	addressContainer AddressContainer,
	tracker AccountTracker,
	stake *big.Int,
	address []byte,
	schnorr []byte,
	bls []byte,
) (*PeerAccount, error) {
	if addressContainer == nil {
		return nil, ErrNilAddressContainer
	}
	if tracker == nil {
		return nil, ErrNilAccountTracker
	}
	if stake == nil {
		return nil, ErrNilStake
	}
	if address == nil {
		return nil, ErrNilAddress
	}
	if schnorr == nil {
		return nil, ErrNilSchnorrPublicKey
	}
	if bls == nil {
		return nil, ErrNilBLSPublicKey
	}

	return &PeerAccount{
		Stake:            big.NewInt(0).Set(stake),
		Address:          address,
		SchnorrPublicKey: schnorr,
		BLSPublicKey:     bls,
		addressContainer: addressContainer,
		accountTracker:   tracker,
		dataTrieTracker:  NewTrackableDataTrie(nil),
	}, nil
}

// IsInterfaceNil return if there is no value under the interface
func (a *PeerAccount) IsInterfaceNil() bool {
	if a == nil {
		return true
	}
	return false
}

// AddressContainer returns the address associated with the account
func (a *PeerAccount) AddressContainer() AddressContainer {
	return a.addressContainer
}

// SetNonceWithJournal sets the account's nonce, saving the old nonce before changing
func (a *PeerAccount) SetNonceWithJournal(nonce uint64) error {
	entry, err := NewJournalEntryNonce(a, a.Nonce)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Nonce = nonce

	return a.accountTracker.SaveAccount(a)
}

//SetNonce saves the nonce to the account
func (a *PeerAccount) SetNonce(nonce uint64) {
	a.Nonce = nonce
}

// GetNonce gets the nonce of the account
func (a *PeerAccount) GetNonce() uint64 {
	return a.Nonce
}

// SetBalanceWithJournal sets the account's balance, saving the old balance before changing
func (a *PeerAccount) SetBalanceWithJournal(balance *big.Int) error {
	entry, err := NewJournalEntryBalance(a, a.Balance)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.Balance = balance

	return a.accountTracker.SaveAccount(a)
}

//------- code / code hash

// GetCodeHash returns the code hash associated with this account
func (a *PeerAccount) GetCodeHash() []byte {
	return a.CodeHash
}

// SetCodeHash sets the code hash associated with the account
func (a *PeerAccount) SetCodeHash(codeHash []byte) {
	a.CodeHash = codeHash
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (a *PeerAccount) SetCodeHashWithJournal(codeHash []byte) error {
	entry, err := NewBaseJournalEntryCodeHash(a, a.CodeHash)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.CodeHash = codeHash

	return a.accountTracker.SaveAccount(a)
}

// GetCode gets the actual code that needs to be run in the VM
func (a *PeerAccount) GetCode() []byte {
	return a.code
}

// SetCode sets the actual code that needs to be run in the VM
func (a *PeerAccount) SetCode(code []byte) {
	a.code = code
}

//------- data trie / root hash

// GetRootHash returns the root hash associated with this account
func (a *PeerAccount) GetRootHash() []byte {
	return a.RootHash
}

// SetRootHash sets the root hash associated with the account
func (a *PeerAccount) SetRootHash(roothash []byte) {
	a.RootHash = roothash
}

// SetRootHashWithJournal sets the account's root hash, saving the old root hash before changing
func (a *PeerAccount) SetRootHashWithJournal(rootHash []byte) error {
	entry, err := NewBaseJournalEntryRootHash(a, a.RootHash)
	if err != nil {
		return err
	}

	a.accountTracker.Journalize(entry)
	a.RootHash = rootHash

	return a.accountTracker.SaveAccount(a)
}

// DataTrie returns the trie that holds the current account's data
func (a *PeerAccount) DataTrie() data.Trie {
	return a.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (a *PeerAccount) SetDataTrie(trie data.Trie) {
	a.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (a *PeerAccount) DataTrieTracker() DataTrieTracker {
	return a.dataTrieTracker
}

//TODO add Cap'N'Proto converter funcs
