package data

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// TriePruningIdentifier is the type for trie pruning identifiers
type TriePruningIdentifier byte

const (
	// OldRoot is appended to the key when oldHashes are added to the evictionWaitingList
	OldRoot TriePruningIdentifier = 0
	// NewRoot is appended to the key when newHashes are added to the evictionWaitingList
	NewRoot TriePruningIdentifier = 1
)

// ModifiedHashes is used to memorize all old hashes and new hashes from when a trie is committed
type ModifiedHashes map[string]struct{}

// HeaderHandler defines getters and setters for header data holder
type HeaderHandler interface {
	GetShardID() uint32
	GetNonce() uint64
	GetEpoch() uint32
	GetRound() uint64
	GetRootHash() []byte
	GetValidatorStatsRootHash() []byte
	GetPrevHash() []byte
	GetPrevRandSeed() []byte
	GetRandSeed() []byte
	GetPubKeysBitmap() []byte
	GetSignature() []byte
	GetLeaderSignature() []byte
	GetChainID() []byte
	GetTimeStamp() uint64
	GetTxCount() uint32
	GetReceiptsHash() []byte

	SetShardID(shId uint32)
	SetNonce(n uint64)
	SetEpoch(e uint32)
	SetRound(r uint64)
	SetTimeStamp(ts uint64)
	SetRootHash(rHash []byte)
	SetValidatorStatsRootHash(rHash []byte)
	SetPrevHash(pvHash []byte)
	SetPrevRandSeed(pvRandSeed []byte)
	SetRandSeed(randSeed []byte)
	SetPubKeysBitmap(pkbm []byte)
	SetSignature(sg []byte)
	SetLeaderSignature(sg []byte)
	SetChainID(chainID []byte)
	SetTxCount(txCount uint32)

	IsStartOfEpochBlock() bool
	GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32

	IsInterfaceNil() bool
	ItemsInBody() uint32
	ItemsInHeader() uint32
	Clone() HeaderHandler
	CheckChainID(reference []byte) error
}

// BodyHandler interface for a block body
type BodyHandler interface {
	// IntegrityAndValidity checks the integrity and validity of the block
	IntegrityAndValidity() error
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// ChainHandler is the interface defining the functionality a blockchain should implement
type ChainHandler interface {
	GetGenesisHeader() HeaderHandler
	SetGenesisHeader(gb HeaderHandler) error
	GetGenesisHeaderHash() []byte
	SetGenesisHeaderHash(hash []byte)
	GetCurrentBlockHeader() HeaderHandler
	SetCurrentBlockHeader(bh HeaderHandler) error
	GetCurrentBlockHeaderHash() []byte
	SetCurrentBlockHeaderHash(hash []byte)
	GetCurrentBlockBody() BodyHandler
	SetCurrentBlockBody(body BodyHandler) error
	GetLocalHeight() int64
	SetLocalHeight(height int64)
	GetNetworkHeight() int64
	SetNetworkHeight(height int64)
	HasBadBlock(blockHash []byte) bool
	PutBadBlock(blockHash []byte)
	IsInterfaceNil() bool
}

// TransactionHandler defines the type of executable transaction
type TransactionHandler interface {
	IsInterfaceNil() bool

	GetValue() *big.Int
	GetNonce() uint64
	GetData() []byte
	GetRecvAddress() []byte
	GetSndAddress() []byte
	GetGasLimit() uint64
	GetGasPrice() uint64

	SetValue(*big.Int)
	SetData([]byte)
	SetRecvAddress([]byte)
	SetSndAddress([]byte)
}

//Trie is an interface for Merkle Trees implementations
type Trie interface {
	Get(key []byte) ([]byte, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	Root() ([]byte, error)
	Commit() error
	Recreate(root []byte) (Trie, error)
	String() string
	CancelPrune(rootHash []byte, identifier TriePruningIdentifier)
	Prune(rootHash []byte, identifier TriePruningIdentifier) error
	TakeSnapshot(rootHash []byte)
	SetCheckpoint(rootHash []byte)
	ResetOldHashes() [][]byte
	AppendToOldHashes([][]byte)
	Database() DBWriteCacher
	GetSerializedNodes([]byte, uint64) ([][]byte, error)
	GetAllLeaves() (map[string][]byte, error)
	IsPruningEnabled() bool
	IsInterfaceNil() bool
	ClosePersister() error
}

// DBWriteCacher is used to cache changes made to the trie, and only write to the database when it's needed
type DBWriteCacher interface {
	Put(key, val []byte) error
	Get(key []byte) ([]byte, error)
	Remove(key []byte) error
	Close() error
	IsInterfaceNil() bool
}

// DBRemoveCacher is used to cache keys that will be deleted from the database
type DBRemoveCacher interface {
	Put([]byte, ModifiedHashes) error
	Evict([]byte) (ModifiedHashes, error)
	GetSize() uint
	PresentInNewHashes(hash string) (bool, error)
	IsInterfaceNil() bool
}

// TrieSyncer synchronizes the trie, asking on the network for the missing nodes
type TrieSyncer interface {
	StartSyncing(rootHash []byte) error
	IsInterfaceNil() bool
}

// StorageManager manages all trie storage operations
type StorageManager interface {
	Database() DBWriteCacher
	TakeSnapshot([]byte, marshal.Marshalizer, hashing.Hasher)
	SetCheckpoint([]byte, marshal.Marshalizer, hashing.Hasher)
	Prune([]byte) error
	CancelPrune([]byte)
	MarkForEviction([]byte, ModifiedHashes) error
	GetDbThatContainsHash([]byte) DBWriteCacher
	IsPruningEnabled() bool
	IsInterfaceNil() bool
}

// TrieFactory creates new tries
type TrieFactory interface {
	Create(config.StorageConfig, bool) (Trie, error)
	IsInterfaceNil() bool
}

// MarshalizedBodyAndHeader holds marshalized body and header
type MarshalizedBodyAndHeader struct {
	Body   []byte
	Header []byte
}
