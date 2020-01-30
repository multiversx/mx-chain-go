package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// MiniBlockData is the data to be saved in shard account for any shard
type MiniBlockData struct {
	Hash            []byte
	ReceiverShardId uint32
	SenderShardId   uint32
	TxCount         uint32
}

// MetaAccount is the struct used in serialization/deserialization
type MetaAccount struct {
	Round         uint64
	Nonce         uint64
	TxCount       *big.Int
	CodeHash      []byte
	RootHash      []byte
	MiniBlocks    []*MiniBlockData
	PubKeyLeader  []byte
	ShardRootHash []byte
	Address       []byte

	addressContainer AddressContainer
	code             []byte
	accountTracker   AccountTracker
	dataTrieTracker  DataTrieTracker
}

// NewMetaAccount creates new simple meta account for an AccountContainer (that has just been initialized)
func NewMetaAccount(addressContainer AddressContainer, tracker AccountTracker) (*MetaAccount, error) {
	if addressContainer == nil || addressContainer.IsInterfaceNil() {
		return nil, ErrNilAddressContainer
	}
	if tracker == nil || tracker.IsInterfaceNil() {
		return nil, ErrNilAccountTracker
	}

	addressBytes := addressContainer.Bytes()

	return &MetaAccount{
		TxCount:          big.NewInt(0),
		addressContainer: addressContainer,
		Address:          addressBytes,
		accountTracker:   tracker,
		dataTrieTracker:  NewTrackableDataTrie(addressBytes, nil),
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ma *MetaAccount) IsInterfaceNil() bool {
	return ma == nil
}

// AddressContainer returns the address associated with the account
func (ma *MetaAccount) AddressContainer() AddressContainer {
	return ma.addressContainer
}

// SetRoundWithJournal sets the account's round, saving the old round before changing
func (ma *MetaAccount) SetRoundWithJournal(round uint64) error {
	entry, err := NewMetaJournalEntryRound(ma, ma.Round)
	if err != nil {
		return err
	}

	ma.accountTracker.Journalize(entry)
	ma.Round = round

	return ma.accountTracker.SaveAccount(ma)
}

// SetTxCountWithJournal sets the total tx count for this shard, saving the old txCount before changing
func (ma *MetaAccount) SetTxCountWithJournal(txCount *big.Int) error {
	entry, err := NewMetaJournalEntryTxCount(ma, ma.TxCount)
	if err != nil {
		return err
	}

	ma.accountTracker.Journalize(entry)
	ma.TxCount = txCount

	return ma.accountTracker.SaveAccount(ma)
}

// SetMiniBlocksDataWithJournal sets the current final mini blocks header data,
// saving the old mini blocks header data before changing
func (ma *MetaAccount) SetMiniBlocksDataWithJournal(miniBlocksData []*MiniBlockData) error {
	entry, err := NewMetaJournalEntryMiniBlocksData(ma, ma.MiniBlocks)
	if err != nil {
		return err
	}

	ma.accountTracker.Journalize(entry)
	ma.MiniBlocks = miniBlocksData

	return ma.accountTracker.SaveAccount(ma)
}

// SetShardRootHashWithJournal sets the account's root hash, saving the old root hash before changing
func (ma *MetaAccount) SetShardRootHashWithJournal(shardRootHash []byte) error {
	entry, err := NewMetaJournalEntryShardRootHash(ma, ma.ShardRootHash)
	if err != nil {
		return err
	}

	ma.accountTracker.Journalize(entry)
	ma.ShardRootHash = shardRootHash

	return ma.accountTracker.SaveAccount(ma)
}

//------- code / code hash

// GetCodeHash returns the code hash associated with this account
func (ma *MetaAccount) GetCodeHash() []byte {
	return ma.CodeHash
}

// SetCodeHash sets the code hash associated with the account
func (ma *MetaAccount) SetCodeHash(roothash []byte) {
	ma.CodeHash = roothash
}

// SetCodeHashWithJournal sets the account's code hash, saving the old code hash before changing
func (ma *MetaAccount) SetCodeHashWithJournal(codeHash []byte) error {
	entry, err := NewBaseJournalEntryCodeHash(ma, ma.CodeHash)
	if err != nil {
		return err
	}

	ma.accountTracker.Journalize(entry)
	ma.CodeHash = codeHash

	return ma.accountTracker.SaveAccount(ma)
}

// GetCode gets the actual code that needs to be run in the VM
func (ma *MetaAccount) GetCode() []byte {
	return ma.code
}

// SetCode sets the actual code that needs to be run in the VM
func (ma *MetaAccount) SetCode(code []byte) {
	ma.code = code
}

//------- data trie / root hash

// GetRootHash returns the root hash associated with this account
func (ma *MetaAccount) GetRootHash() []byte {
	return ma.RootHash
}

// SetRootHash sets the root hash associated with the account
func (ma *MetaAccount) SetRootHash(roothash []byte) {
	ma.RootHash = roothash
}

// SetNonceWithJournal sets the account's nonce, saving the old nonce before changing
func (ma *MetaAccount) SetNonceWithJournal(nonce uint64) error {
	entry, err := NewBaseJournalEntryNonce(ma, ma.Nonce)
	if err != nil {
		return err
	}

	ma.accountTracker.Journalize(entry)
	ma.Nonce = nonce

	return ma.accountTracker.SaveAccount(ma)
}

//SetNonce saves the nonce to the account
func (ma *MetaAccount) SetNonce(nonce uint64) {
	ma.Nonce = nonce
}

// GetNonce gets the nonce of the account
func (ma *MetaAccount) GetNonce() uint64 {
	return ma.Nonce
}

// DataTrie returns the trie that holds the current account's data
func (ma *MetaAccount) DataTrie() data.Trie {
	return ma.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (ma *MetaAccount) SetDataTrie(trie data.Trie) {
	ma.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (ma *MetaAccount) DataTrieTracker() DataTrieTracker {
	return ma.dataTrieTracker
}

//TODO add Cap'N'Proto converter funcs
