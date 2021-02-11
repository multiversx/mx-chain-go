package rewardTransaction

import (
	"fmt"
	"math/big"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.InterceptedData = (*InterceptedRewardTransaction)(nil)

// InterceptedRewardTransaction holds and manages a transaction based struct with extended functionality
type InterceptedRewardTransaction struct {
	rTx               *rewardTx.RewardTx
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	pubkeyConv        core.PubkeyConverter
	coordinator       sharding.Coordinator
	hash              []byte
	rcvShard          uint32
	sndShard          uint32
	isForCurrentShard bool
}

// NewInterceptedRewardTransaction returns a new instance of InterceptedRewardTransaction
func NewInterceptedRewardTransaction(
	rewardTxBuff []byte,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	pubkeyConv core.PubkeyConverter,
	coordinator sharding.Coordinator,
) (*InterceptedRewardTransaction, error) {
	if rewardTxBuff == nil {
		return nil, process.ErrNilBuffer
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(pubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(coordinator) {
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
		pubkeyConv:  pubkeyConv,
		coordinator: coordinator,
	}

	err = inRewardTx.processFields(rewardTxBuff)
	if err != nil {
		return nil, err
	}

	return inRewardTx, nil
}

func (inRTx *InterceptedRewardTransaction) processFields(rewardTxBuff []byte) error {
	inRTx.hash = inRTx.hasher.Compute(string(rewardTxBuff))

	inRTx.rcvShard = inRTx.coordinator.ComputeId(inRTx.rTx.RcvAddr)
	inRTx.sndShard = core.MetachainShardId

	if inRTx.coordinator.SelfId() == core.MetachainShardId {
		inRTx.isForCurrentShard = false
		return nil
	}

	isForCurrentShardRecv := inRTx.rcvShard == inRTx.coordinator.SelfId()
	isForCurrentShardSender := inRTx.sndShard == inRTx.coordinator.SelfId()
	inRTx.isForCurrentShard = isForCurrentShardRecv || isForCurrentShardSender

	return nil
}

// integrity checks for not nil fields and negative value
func (inRTx *InterceptedRewardTransaction) integrity() error {
	err := inRTx.rTx.CheckIntegrity()
	if err != nil {
		return err
	}

	return nil
}

// Nonce returns the transaction nonce
func (inRTx *InterceptedRewardTransaction) Nonce() uint64 {
	return inRTx.rTx.GetNonce()
}

// Fee represents the reward transaction fee. It is always 0
func (inRTx *InterceptedRewardTransaction) Fee() *big.Int {
	return big.NewInt(0)
}

// SenderAddress returns the transaction sender address
func (inRTx *InterceptedRewardTransaction) SenderAddress() []byte {
	return nil
}

// ReceiverShardId returns the receiver shard
func (inRTx *InterceptedRewardTransaction) ReceiverShardId() uint32 {
	return inRTx.rcvShard
}

// SenderShardId returns the sender shard
func (inRTx *InterceptedRewardTransaction) SenderShardId() uint32 {
	return inRTx.sndShard
}

// Transaction returns the reward transaction pointer that actually holds the data
func (inRTx *InterceptedRewardTransaction) Transaction() data.TransactionHandler {
	return inRTx.rTx
}

// Hash gets the hash of this transaction
func (inRTx *InterceptedRewardTransaction) Hash() []byte {
	return inRTx.hash
}

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (inRTx *InterceptedRewardTransaction) CheckValidity() error {
	err := inRTx.integrity()
	if err != nil {
		return err
	}

	return nil
}

// IsForCurrentShard returns true if this transaction is meant to be processed by the node from this shard
func (inRTx *InterceptedRewardTransaction) IsForCurrentShard() bool {
	return inRTx.isForCurrentShard
}

// Type returns the type of this intercepted data
func (inRTx *InterceptedRewardTransaction) Type() string {
	return "intercepted reward tx"
}

// String returns the reward's most important fields as string
func (inRTx *InterceptedRewardTransaction) String() string {
	return fmt.Sprintf("epoch=%d, round=%d, address=%s, value=%s",
		inRTx.rTx.Epoch,
		inRTx.rTx.Round,
		logger.DisplayByteSlice(inRTx.rTx.RcvAddr),
		inRTx.rTx.Value.String(),
	)
}

// Identifiers returns the identifiers used in requests
func (inRTx *InterceptedRewardTransaction) Identifiers() [][]byte {
	return [][]byte{inRTx.hash}
}

// IsInterfaceNil returns true if there is no value under the interface
func (inRTx *InterceptedRewardTransaction) IsInterfaceNil() bool {
	return inRTx == nil
}
