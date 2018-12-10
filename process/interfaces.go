package process

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// Checker checks the integrity of a data structure
type Checker interface {
	Check() bool
}

// SigVerifier provides functionality to verify a signature
type SigVerifier interface {
	VerifySig() bool
}

type Hashed interface {
	SetHash([]byte)
	Hash() []byte
}

// TransactionInterceptorAdapter is the interface used in interception of transactions
type TransactionInterceptorAdapter interface {
	Checker
	SigVerifier
	Hashed
	p2p.Newer
	RcvShard() uint32
	SndShard() uint32
	IsAddressedToOtherShards() bool
	SetAddressConverter(converter state.AddressConverter)
	AddressConverter() state.AddressConverter
	GetTransaction() *transaction.Transaction
}

// HeaderInterceptorAdapter is the interface used in interception of headers
type HeaderInterceptorAdapter interface {
	Checker
	SigVerifier
	Hashed
	p2p.Newer
	GetHeader() *block.Header
	Shard() uint32
}

type TxBlockBodyInterceptorAdapter interface {
	Checker
	Hashed
	p2p.Newer
	GetTxBlockBody() *block.TxBlockBody
	Shard() uint32
}
