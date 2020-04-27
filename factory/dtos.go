package factory

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// StateComponents struct holds the state components of the Elrond protocol
type StateComponents struct {
	AddressPubkeyConverter   state.PubkeyConverter
	ValidatorPubkeyConverter state.PubkeyConverter
	PeerAccounts             state.AccountsAdapter
	AccountsAdapter          state.AccountsAdapter
	InBalanceForShard        map[string]*big.Int
}

// CoreComponents is the DTO used for core components
type CoreComponents struct {
	Hasher                   hashing.Hasher
	InternalMarshalizer      marshal.Marshalizer
	VmMarshalizer            marshal.Marshalizer
	TxSignMarshalizer        marshal.Marshalizer
	TriesContainer           state.TriesHolder
	TrieStorageManagers      map[string]data.StorageManager
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	StatusHandler            core.AppStatusHandler
	ChainID                  []byte
}
