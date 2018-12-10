package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// InterceptedTransaction holds and manages a transaction based struct with extended functionality
type InterceptedTransaction struct {
	*transaction.Transaction

	hash                     []byte
	rcvShard                 uint32
	sndShard                 uint32
	isAddressedToOtherShards bool
	addrConv                 state.AddressConverter
}

// NewInterceptedTransaction returns a new instance of InterceptedTransaction
func NewInterceptedTransaction() *InterceptedTransaction {
	return &InterceptedTransaction{
		Transaction: &transaction.Transaction{},
		addrConv:    nil,
	}
}

// New returns a new instance of this struct (used in topics)
func (inTx *InterceptedTransaction) New() p2p.Newer {
	return NewInterceptedTransaction()
}

// ID returns the ID of this object. Set to return the hash of the transaction
func (inTx *InterceptedTransaction) ID() string {
	return string(inTx.hash)
}

// Check returns true if the transaction data pass some sanity check and some validation checks
func (inTx *InterceptedTransaction) Check() bool {
	if !inTx.sanityCheck() {
		return false
	}

	if !inTx.dataValidity() {
		return false
	}

	inTx.rcvShard = inTx.computeShard(inTx.RcvAddr)
	inTx.sndShard = inTx.computeShard(inTx.SndAddr)

	return true
}

// sanityCheck, practically checks for not nil fields
func (inTx *InterceptedTransaction) sanityCheck() bool {
	notNilFields := inTx.Transaction != nil &&
		inTx.Signature != nil &&
		inTx.Challenge != nil &&
		inTx.RcvAddr != nil &&
		inTx.SndAddr != nil

	if !notNilFields {
		log.Debug("tx with nil fields")
	}

	return notNilFields
}

// dataValidity compute snd and rcv shards, checks for sign value and others
func (inTx *InterceptedTransaction) dataValidity() bool {
	//TODO set isAddressedToOtherShards to true and return when we will have sharding in place and the statement applies
	inTx.isAddressedToOtherShards = false

	if inTx.Value.Sign() < 0 {
		log.Debug("tx with negative value")
		return false
	}

	if inTx.addrConv == nil {
		log.Debug("nil AddressConverter so can not verify addresses")
		return false
	}

	_, err := inTx.addrConv.CreateAddressFromPublicKeyBytes(inTx.SndAddr)
	if err != nil {
		log.Debug("tx with invalid sender address")
		return false
	}
	_, err = inTx.addrConv.CreateAddressFromPublicKeyBytes(inTx.RcvAddr)
	if err != nil {
		log.Debug("tx with invalid receiver address")
		return false
	}

	return true
}

// VerifySig checks if the tx is correctly signed
func (inTx *InterceptedTransaction) VerifySig() bool {
	//TODO add real sig verify here
	return true
}

func (inTx *InterceptedTransaction) computeShard(address []byte) uint32 {
	//TODO add real logic here
	return 0
}

// RcvShard returns the receiver shard
func (inTx *InterceptedTransaction) RcvShard() uint32 {
	return inTx.rcvShard
}

// SndShard returns the sender shard
func (inTx *InterceptedTransaction) SndShard() uint32 {
	return inTx.sndShard
}

// IsAddressedToOtherShards returns true if this transaction is not meant to be processed by the node from this shard
func (inTx *InterceptedTransaction) IsAddressedToOtherShards() bool {
	return inTx.isAddressedToOtherShards
}

// SetAddressConverter sets the AddressConverter implementation used in address processing
func (inTx *InterceptedTransaction) SetAddressConverter(converter state.AddressConverter) {
	inTx.addrConv = converter
}

// AddressConverter returns the AddressConverter implementation used in address processing
func (inTx *InterceptedTransaction) AddressConverter() state.AddressConverter {
	return inTx.addrConv
}

// GetTransaction returns the transaction pointer that actually holds the data
func (inTx *InterceptedTransaction) GetTransaction() *transaction.Transaction {
	return inTx.Transaction
}

// SetHash sets the hash of this transaction. The hash will also be the ID of this object
func (inTx *InterceptedTransaction) SetHash(hash []byte) {
	inTx.hash = hash
}

// Hash gets the hash of this transaction
func (inTx *InterceptedTransaction) Hash() []byte {
	return inTx.hash
}
