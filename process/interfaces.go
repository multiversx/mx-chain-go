package process

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

const (
	// TxInterceptor is the name for the transaction interceptor topic
	TxInterceptor = "tx"
	// HeaderInterceptor is the name for the header interceptor topic
	HeaderInterceptor = "hdr"
	// TxBlockBodyInterceptor is the name for the tx block body interceptor topic
	TxBlockBodyInterceptor = "txBlk"
	// StateBlockBodyInterceptor is the name for the state block body interceptor topic
	StateBlockBodyInterceptor = "stateBlk"
	// PeerBlockBodyInterceptor is the name for the peer block body interceptor topic
	PeerBlockBodyInterceptor = "peerBlk"
)

// Checker checks the integrity of a data structure
type Checker interface {
	Check() bool
}

// SigVerifier provides functionality to verify a signature
type SigVerifier interface {
	VerifySig() bool
}

// Hashed interface provides functionality over hashable objects
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
	Shard() uint32
	GetHeader() *block.Header
}

// BlockBodyInterceptorAdapter defines what a block body object should do
type BlockBodyInterceptorAdapter interface {
	Checker
	Hashed
	p2p.Newer
	Shard() uint32
}

type Interceptor interface {
	Name() string
	SetCheckReceivedObjectHandler(func(newer p2p.Newer, rawData []byte) bool)
	CheckReceivedObjectHandler() func(newer p2p.Newer, rawData []byte) bool
}
