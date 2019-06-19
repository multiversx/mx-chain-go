package smartContract

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// InterceptedSmartContractResult holds and manages a transaction based struct with extended functionality
type InterceptedSmartContractResult struct {
	scr                      *smartContractResult.SmartContractResult
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	addrConv                 state.AddressConverter
	coordinator              sharding.Coordinator
	hash                     []byte
	rcvShard                 uint32
	sndShard                 uint32
	isAddressedToOtherShards bool
}

// NewInterceptedSmartContractResult returns a new instance of InterceptedSmartContractResult
func NewInterceptedSmartContractResult(
	scrBuff []byte,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	addrConv state.AddressConverter,
	coordinator sharding.Coordinator,
) (*InterceptedSmartContractResult, error) {

	if scrBuff == nil {
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

	scr := &smartContractResult.SmartContractResult{}
	err := marshalizer.Unmarshal(scr, scrBuff)
	if err != nil {
		return nil, err
	}

	inScr := &InterceptedSmartContractResult{
		scr:         scr,
		marshalizer: marshalizer,
		hasher:      hasher,
		addrConv:    addrConv,
		coordinator: coordinator,
	}

	err = inScr.processFields(scrBuff)
	if err != nil {
		return nil, err
	}

	err = inScr.integrity()
	if err != nil {
		return nil, err
	}

	err = inScr.verifyIfNotarized(inScr.hash)
	if err != nil {
		return nil, err
	}

	return inScr, nil
}

func (inScr *InterceptedSmartContractResult) processFields(scrBuffWithSig []byte) error {
	inScr.hash = inScr.hasher.Compute(string(scrBuffWithSig))

	sndAddr, err := inScr.addrConv.CreateAddressFromPublicKeyBytes(inScr.scr.SndAddr)
	if err != nil {
		return process.ErrInvalidSndAddr
	}

	rcvAddr, err := inScr.addrConv.CreateAddressFromPublicKeyBytes(inScr.scr.RcvAddr)
	if err != nil {
		return process.ErrInvalidRcvAddr
	}

	inScr.rcvShard = inScr.coordinator.ComputeId(rcvAddr)
	inScr.sndShard = inScr.coordinator.ComputeId(sndAddr)

	inScr.isAddressedToOtherShards = inScr.rcvShard != inScr.coordinator.SelfId() &&
		inScr.sndShard != inScr.coordinator.SelfId()

	return nil
}

// integrity checks for not nil fields and negative value
func (inScr *InterceptedSmartContractResult) integrity() error {
	if len(inScr.scr.RcvAddr) == 0 {
		return process.ErrNilRcvAddr
	}

	if len(inScr.scr.SndAddr) == 0 {
		return process.ErrNilSndAddr
	}

	if inScr.scr.Value == nil {
		return process.ErrNilValue
	}

	if inScr.scr.Value.Cmp(big.NewInt(0)) < 0 {
		return process.ErrNegativeValue
	}

	if len(inScr.scr.TxHash) == 0 {
		return process.ErrNilTxHash
	}

	return nil
}

// verifyIfNotarized checks if the scr was already notarized
func (inScr *InterceptedSmartContractResult) verifyIfNotarized(scrBuff []byte) error {
	// TODO: implement this for flood protection purposes
	return nil
}

// RcvShard returns the receiver shard
func (inScr *InterceptedSmartContractResult) RcvShard() uint32 {
	return inScr.rcvShard
}

// SndShard returns the sender shard
func (inScr *InterceptedSmartContractResult) SndShard() uint32 {
	return inScr.sndShard
}

// IsAddressedToOtherShards returns true if this transaction is not meant to be processed by the node from this shard
func (inScr *InterceptedSmartContractResult) IsAddressedToOtherShards() bool {
	return inScr.isAddressedToOtherShards
}

// SmartContractResult returns the transaction pointer that actually holds the data
func (inScr *InterceptedSmartContractResult) SmartContractResult() *smartContractResult.SmartContractResult {
	return inScr.scr
}

// Hash gets the hash of this transaction
func (inScr *InterceptedSmartContractResult) Hash() []byte {
	return inScr.hash
}
