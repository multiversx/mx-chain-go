package factory

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

// StateComponents struct holds the state components of the Elrond protocol
type StateComponents struct {
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
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
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	StatusHandler            core.AppStatusHandler
	ChainID                  []byte
	MinTransactionVersion    uint32
	TxSignHasher             hashing.Hasher
}

// CryptoParams is a DTO for holding block signing parameters
type CryptoParams struct {
	KeyGenerator    crypto.KeyGenerator
	PrivateKey      crypto.PrivateKey
	PublicKey       crypto.PublicKey
	PublicKeyBytes  []byte
	PublicKeyString string
}

// DataComponents struct holds the data components
type DataComponents struct {
	Blkc     data.ChainHandler
	Store    dataRetriever.StorageService
	Datapool dataRetriever.PoolsHolder
}

// TriesComponents holds the tries components
type TriesComponents struct {
	TriesContainer      state.TriesHolder
	TrieStorageManagers map[string]data.StorageManager
}

// CryptoComponents struct holds the crypto components
type CryptoComponents struct {
	TxSingleSigner       crypto.SingleSigner
	SingleSigner         crypto.SingleSigner
	MultiSigner          crypto.MultiSigner
	BlockSignKeyGen      crypto.KeyGenerator
	TxSignKeyGen         crypto.KeyGenerator
	InitialPubKeys       map[uint32][]string
	MessageSignVerifier  vm.MessageSignVerifier
	PeerSignatureHandler crypto.PeerSignatureHandler
}

// NetworkComponents struct holds the network components
type NetworkComponents struct {
	NetMessenger           p2p.Messenger
	InputAntifloodHandler  P2PAntifloodHandler
	OutputAntifloodHandler P2PAntifloodHandler
	PeerBlackListHandler   process.PeerBlackListCacher
	PkTimeCache            process.TimeCacher
}
