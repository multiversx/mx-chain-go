package unsigned

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedUnsignedTransaction holds and manages a transaction based struct with extended functionality
type InterceptedUnsignedTransaction struct {
	uTx               *smartContractResult.SmartContractResult
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	addrConv          state.AddressConverter
	coordinator       sharding.Coordinator
	hash              []byte
	rcvShard          uint32
	sndShard          uint32
	isForCurrentShard bool
	sndAddr           state.AddressContainer
}

// NewInterceptedUnsignedTransaction returns a new instance of InterceptedUnsignedTransaction
func NewInterceptedUnsignedTransaction(
	uTxBuff []byte,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	addrConv state.AddressConverter,
	coordinator sharding.Coordinator,
) (*InterceptedUnsignedTransaction, error) {

	if uTxBuff == nil {
		return nil, process.ErrNilBuffer
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if addrConv == nil || addrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if coordinator == nil || coordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}

	uTx := &smartContractResult.SmartContractResult{}
	err := marshalizer.Unmarshal(uTx, uTxBuff)
	if err != nil {
		return nil, err
	}

	inUTx := &InterceptedUnsignedTransaction{
		uTx:         uTx,
		marshalizer: marshalizer,
		hasher:      hasher,
		addrConv:    addrConv,
		coordinator: coordinator,
	}

	err = inUTx.processFields(uTxBuff)
	if err != nil {
		return nil, err
	}

	return inUTx, nil
}

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (inUTx *InterceptedUnsignedTransaction) CheckValidity() error {
	err := inUTx.integrity()
	if err != nil {
		return err
	}

	return nil
}

func (inUTx *InterceptedUnsignedTransaction) processFields(uTxBuffWithSig []byte) error {
	inUTx.hash = inUTx.hasher.Compute(string(uTxBuffWithSig))

	var err error
	inUTx.sndAddr, err = inUTx.addrConv.CreateAddressFromPublicKeyBytes(inUTx.uTx.SndAddr)
	if err != nil {
		return process.ErrInvalidSndAddr
	}

	rcvAddr, err := inUTx.addrConv.CreateAddressFromPublicKeyBytes(inUTx.uTx.RcvAddr)
	if err != nil {
		return process.ErrInvalidRcvAddr
	}

	inUTx.rcvShard = inUTx.coordinator.ComputeId(rcvAddr)
	inUTx.sndShard = inUTx.coordinator.ComputeId(inUTx.sndAddr)

	isForCurrentShardRecv := inUTx.rcvShard == inUTx.coordinator.SelfId()
	isForCurrentShardSender := inUTx.sndShard == inUTx.coordinator.SelfId()
	inUTx.isForCurrentShard = isForCurrentShardRecv || isForCurrentShardSender

	return nil
}

// integrity checks for not nil fields and negative value
func (inUTx *InterceptedUnsignedTransaction) integrity() error {
	if len(inUTx.uTx.RcvAddr) == 0 {
		return process.ErrNilRcvAddr
	}
	if len(inUTx.uTx.SndAddr) == 0 {
		return process.ErrNilSndAddr
	}
	if inUTx.uTx.Value == nil {
		return process.ErrNilValue
	}
	if inUTx.uTx.Value.Cmp(big.NewInt(0)) < 0 {
		return process.ErrNegativeValue
	}
	if len(inUTx.uTx.TxHash) == 0 {
		return process.ErrNilTxHash
	}

	return nil
}

// Nonce returns the transaction nonce
func (inUTx *InterceptedUnsignedTransaction) Nonce() uint64 {
	return inUTx.uTx.Nonce
}

// SenderAddress returns the transaction sender address
func (inUTx *InterceptedUnsignedTransaction) SenderAddress() state.AddressContainer {
	return inUTx.sndAddr
}

// ReceiverShardId returns the receiver shard
func (inUTx *InterceptedUnsignedTransaction) ReceiverShardId() uint32 {
	return inUTx.rcvShard
}

// SenderShardId returns the sender shard
func (inUTx *InterceptedUnsignedTransaction) SenderShardId() uint32 {
	return inUTx.sndShard
}

// IsForCurrentShard returns true if this transaction is meant to be processed by the node from this shard
func (inUTx *InterceptedUnsignedTransaction) IsForCurrentShard() bool {
	return inUTx.isForCurrentShard
}

// Transaction returns the transaction pointer that actually holds the data
func (inUTx *InterceptedUnsignedTransaction) Transaction() data.TransactionHandler {
	return inUTx.uTx
}

// TotalValue returns the value of the unsigned transaction
func (inUTx *InterceptedUnsignedTransaction) TotalValue() *big.Int {
	copiedVal := big.NewInt(0).Set(inUTx.uTx.Value)
	return copiedVal
}

// Hash gets the hash of this transaction
func (inUTx *InterceptedUnsignedTransaction) Hash() []byte {
	return inUTx.hash
}

// IsInterfaceNil returns true if there is no value under the interface
func (inUTx *InterceptedUnsignedTransaction) IsInterfaceNil() bool {
	if inUTx == nil {
		return true
	}
	return false
}
