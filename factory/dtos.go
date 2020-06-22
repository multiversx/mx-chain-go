package factory

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// StateComponents struct holds the state components of the Elrond protocol
type StateComponents struct {
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	PeerAccounts             state.AccountsAdapter
	AccountsAdapter          state.AccountsAdapter
	InBalanceForShard        map[string]*big.Int
}

// TriesComponents holds the tries components
type TriesComponents struct {
	TriesContainer      state.TriesHolder
	TrieStorageManagers map[string]data.StorageManager
}
