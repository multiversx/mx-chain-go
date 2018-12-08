package interceptTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"math/big"
)

type InterceptedTransaction struct {
	*transaction.Transaction

	rcvShard                 uint32
	sndShard                 uint32
	isAddressedToOtherShards bool
	addrConv                 state.AddressConverter
}

func NewInterceptedTransaction() *InterceptedTransaction {
	return &InterceptedTransaction{
		Transaction: &transaction.Transaction{},
		addrConv:    nil,
	}
}

func (inTx *InterceptedTransaction) New() p2p.Newer {
	return NewInterceptedTransaction()
}

func (inTx *InterceptedTransaction) ID() string {
	return string(inTx.Transaction.Signature)
}

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

func (inTx *InterceptedTransaction) dataValidity() bool {
	//TODO set isAddressedToOtherShards to true and return when we will have sharding in place and the statement applies
	inTx.isAddressedToOtherShards = false

	//TODO change to big int implementation
	value := big.NewInt(0)
	value.SetUint64(inTx.Value)
	if inTx.Value < 0 {
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

func (inTx *InterceptedTransaction) VerifySig() bool {
	//TODO add real checking here
	return true
}

func (inTx *InterceptedTransaction) computeShard(address []byte) uint32 {
	//TODO add real logic here
	return 0
}

func (inTx *InterceptedTransaction) RcvShard() uint32 {
	return inTx.rcvShard
}

func (inTx *InterceptedTransaction) SndShard() uint32 {
	return inTx.sndShard
}

func (inTx *InterceptedTransaction) IsAddressedToOtherShards() bool {
	return inTx.isAddressedToOtherShards
}

func (inTx *InterceptedTransaction) SetAddressConverter(converter state.AddressConverter) {
	inTx.addrConv = converter
}

func (inTx *InterceptedTransaction) AddressConverter() state.AddressConverter {
	return inTx.addrConv
}

func (inTx *InterceptedTransaction) GetTransaction() *transaction.Transaction {
	return inTx.Transaction
}
