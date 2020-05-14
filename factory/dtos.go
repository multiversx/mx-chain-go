package factory

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// StateComponents struct holds the state components of the Elrond protocol
type StateComponents struct {
	AddressPubkeyConverter   state.PubkeyConverter
	ValidatorPubkeyConverter state.PubkeyConverter
	PeerAccounts             state.AccountsAdapter
	AccountsAdapter          state.AccountsAdapter
	InBalanceForShard        map[string]*big.Int
}

// TriesComponents holds the tries components
type TriesComponents struct {
	TriesContainer      state.TriesHolder
	TrieStorageManagers map[string]data.StorageManager
}

// NetworkComponents struct holds the network components
type NetworkComponents struct {
	NetMessenger           p2p.Messenger
	InputAntifloodHandler  P2PAntifloodHandler
	OutputAntifloodHandler P2PAntifloodHandler
	PeerBlackListHandler   process.BlackListHandler
}
