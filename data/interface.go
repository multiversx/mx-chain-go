package data

import (
	"math/big"
)

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
	GetData() string
	GetRecvAddress() []byte
	GetSndAddress() []byte
	GetGasLimit() uint64
	GetGasPrice() uint64

	SetValue(*big.Int)
	SetData(string)
	SetRecvAddress([]byte)
	SetSndAddress([]byte)
}

//Trie is an interface for Merkle Trees implementations
type Trie interface {
	Get(key []byte) ([]byte, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	Root() ([]byte, error)
	Prove(key []byte) ([][]byte, error)
	VerifyProof(proofs [][]byte, key []byte) (bool, error)
	Commit() error
	Recreate(root []byte) (Trie, error)
	String() string
	DeepClone() (Trie, error)
	GetAllLeaves() (map[string][]byte, error)
	IsInterfaceNil() bool
}

// DBWriteCacher is used to cache changes made to the trie, and only write to the database when it's needed
type DBWriteCacher interface {
	Put(key, val []byte) error
	Get(key []byte) ([]byte, error)
	IsInterfaceNil() bool
}
