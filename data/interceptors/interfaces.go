package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// Checker checks the integrity of a data structure
type Checker interface {
	Check() bool
}

// SigVerifier provides functionality to verify a block Header signature
type SigVerifier interface {
	VerifySig() bool
}

type TransactionInterceptorAdapter interface {
	Checker
	SigVerifier
	p2p.Newer
	RcvShard() uint32
	SndShard() uint32
	IsAddressedToOtherShards() bool
	SetAddressConverter(converter state.AddressConverter)
	AddressConverter() state.AddressConverter
	GetTransaction() *transaction.Transaction
}
