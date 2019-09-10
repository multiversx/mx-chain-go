package rewardTransaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedRewardTransaction holds and manages a transaction based struct with extended functionality
type InterceptedRewardTransaction struct {
	rTx                      *rewardTx.RewardTx
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	addrConv                 state.AddressConverter
	coordinator              sharding.Coordinator
	hash                     []byte
	rcvShard                 uint32
	sndShard                 uint32
	isAddressedToOtherShards bool
}

// NewInterceptedRewardTransaction returns a new instance of InterceptedRewardTransaction
func NewInterceptedRewardTransaction(
	rewardTxBuff []byte,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	addrConv state.AddressConverter,
	coordinator sharding.Coordinator,
) (*InterceptedRewardTransaction, error) {

	if rewardTxBuff == nil {
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

	rTx := &rewardTx.RewardTx{}
	err := marshalizer.Unmarshal(rTx, rewardTxBuff)
	if err != nil {
		return nil, err
	}

	inRewardTx := &InterceptedRewardTransaction{
		rTx:         rTx,
		marshalizer: marshalizer,
		hasher:      hasher,
		addrConv:    addrConv,
		coordinator: coordinator,
	}

	err = inRewardTx.processFields(rewardTxBuff)
	if err != nil {
		return nil, err
	}

	err = inRewardTx.integrity()
	if err != nil {
		return nil, err
	}

	err = inRewardTx.verifyIfNotarized(inRewardTx.hash)
	if err != nil {
		return nil, err
	}

	return inRewardTx, nil
}

func (inRTx *InterceptedRewardTransaction) processFields(rewardTxBuff []byte) error {
	inRTx.hash = inRTx.hasher.Compute(string(rewardTxBuff))

	rcvAddr, err := inRTx.addrConv.CreateAddressFromPublicKeyBytes(inRTx.rTx.RcvAddr)
	if err != nil {
		return process.ErrInvalidRcvAddr
	}

	inRTx.rcvShard = inRTx.coordinator.ComputeId(rcvAddr)
	inRTx.sndShard = inRTx.rTx.ShardId

	inRTx.isAddressedToOtherShards = inRTx.rcvShard != inRTx.coordinator.SelfId() &&
		inRTx.sndShard != inRTx.coordinator.SelfId()

	return nil
}

// integrity checks for not nil fields and negative value
func (inRTx *InterceptedRewardTransaction) integrity() error {
	if len(inRTx.rTx.RcvAddr) == 0 {
		return process.ErrNilRcvAddr
	}

	if inRTx.rTx.Value == nil {
		return process.ErrNilValue
	}

	if inRTx.rTx.Value.Cmp(big.NewInt(0)) < 0 {
		return process.ErrNegativeValue
	}

	return nil
}

// verifyIfNotarized checks if the rewardTx was already notarized
func (inRTx *InterceptedRewardTransaction) verifyIfNotarized(rTxBuff []byte) error {
	// TODO: implement this for flood protection purposes
	// could verify if the epoch/round is behind last committed metachain block
	return nil
}

// RcvShard returns the receiver shard
func (inRTx *InterceptedRewardTransaction) RcvShard() uint32 {
	return inRTx.rcvShard
}

// SndShard returns the sender shard
func (inRTx *InterceptedRewardTransaction) SndShard() uint32 {
	return inRTx.sndShard
}

// IsAddressedToOtherShards returns true if this transaction is not meant to be processed by the node from this shard
func (inRTx *InterceptedRewardTransaction) IsAddressedToOtherShards() bool {
	return inRTx.isAddressedToOtherShards
}

// RewardTransaction returns the reward transaction pointer that actually holds the data
func (inRTx *InterceptedRewardTransaction) RewardTransaction() data.TransactionHandler {
	return inRTx.rTx
}

// Hash gets the hash of this transaction
func (inRTx *InterceptedRewardTransaction) Hash() []byte {
	return inRTx.hash
}
