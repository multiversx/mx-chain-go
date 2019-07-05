package unsigned

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedUnsignedTransaction holds and manages a transaction based struct with extended functionality
type InterceptedUnsignedTransaction struct {
	uTx                      *smartContractResult.SmartContractResult
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	addrConv                 state.AddressConverter
	coordinator              sharding.Coordinator
	hash                     []byte
	rcvShard                 uint32
	sndShard                 uint32
	isAddressedToOtherShards bool
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
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if addrConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if coordinator == nil {
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

	err = inUTx.integrity()
	if err != nil {
		return nil, err
	}

	err = inUTx.verifyIfNotarized(inUTx.hash)
	if err != nil {
		return nil, err
	}

	return inUTx, nil
}

func (inUTx *InterceptedUnsignedTransaction) processFields(uTxBuffWithSig []byte) error {
	inUTx.hash = inUTx.hasher.Compute(string(uTxBuffWithSig))

	sndAddr, err := inUTx.addrConv.CreateAddressFromPublicKeyBytes(inUTx.uTx.SndAddr)
	if err != nil {
		return process.ErrInvalidSndAddr
	}

	rcvAddr, err := inUTx.addrConv.CreateAddressFromPublicKeyBytes(inUTx.uTx.RcvAddr)
	if err != nil {
		return process.ErrInvalidRcvAddr
	}

	inUTx.rcvShard = inUTx.coordinator.ComputeId(rcvAddr)
	inUTx.sndShard = inUTx.coordinator.ComputeId(sndAddr)

	inUTx.isAddressedToOtherShards = inUTx.rcvShard != inUTx.coordinator.SelfId() &&
		inUTx.sndShard != inUTx.coordinator.SelfId()

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

// verifyIfNotarized checks if the uTx was already notarized
func (inUTx *InterceptedUnsignedTransaction) verifyIfNotarized(uTxBuff []byte) error {
	// TODO: implement this for flood protection purposes
	return nil
}

// RcvShard returns the receiver shard
func (inUTx *InterceptedUnsignedTransaction) RcvShard() uint32 {
	return inUTx.rcvShard
}

// SndShard returns the sender shard
func (inUTx *InterceptedUnsignedTransaction) SndShard() uint32 {
	return inUTx.sndShard
}

// IsAddressedToOtherShards returns true if this transaction is not meant to be processed by the node from this shard
func (inUTx *InterceptedUnsignedTransaction) IsAddressedToOtherShards() bool {
	return inUTx.isAddressedToOtherShards
}

// UnsignedTransaction returns the unsigned transaction pointer that actually holds the data
func (inUTx *InterceptedUnsignedTransaction) UnsignedTransaction() data.TransactionHandler {
	return inUTx.uTx
}

// Hash gets the hash of this transaction
func (inUTx *InterceptedUnsignedTransaction) Hash() []byte {
	return inUTx.hash
}
