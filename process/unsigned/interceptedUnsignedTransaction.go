package unsigned

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.TxValidatorHandler = (*InterceptedUnsignedTransaction)(nil)
var _ process.InterceptedData = (*InterceptedUnsignedTransaction)(nil)

// InterceptedUnsignedTransaction holds and manages a transaction based struct with extended functionality
type InterceptedUnsignedTransaction struct {
	uTx               *smartContractResult.SmartContractResult
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	pubkeyConv        core.PubkeyConverter
	coordinator       sharding.Coordinator
	hash              []byte
	rcvShard          uint32
	sndShard          uint32
	isForCurrentShard bool
}

// NewInterceptedUnsignedTransaction returns a new instance of InterceptedUnsignedTransaction
func NewInterceptedUnsignedTransaction(
	uTxBuff []byte,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	pubkeyConv core.PubkeyConverter,
	coordinator sharding.Coordinator,
) (*InterceptedUnsignedTransaction, error) {
	if uTxBuff == nil {
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

	uTx, err := createUtx(marshalizer, uTxBuff)
	if err != nil {
		return nil, err
	}

	inUTx := &InterceptedUnsignedTransaction{
		uTx:         uTx,
		marshalizer: marshalizer,
		hasher:      hasher,
		pubkeyConv:  pubkeyConv,
		coordinator: coordinator,
	}

	err = inUTx.processFields(uTxBuff)
	if err != nil {
		return nil, err
	}

	return inUTx, nil
}

func createUtx(marshalizer marshal.Marshalizer, uTxBuff []byte) (*smartContractResult.SmartContractResult, error) {
	uTx := &smartContractResult.SmartContractResult{}
	err := marshalizer.Unmarshal(uTx, uTxBuff)
	if err != nil {
		return nil, err
	}

	return uTx, nil
}

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (inUTx *InterceptedUnsignedTransaction) CheckValidity() error {
	err := inUTx.uTx.CheckIntegrity()
	if err != nil {
		return err
	}

	return nil
}

func (inUTx *InterceptedUnsignedTransaction) processFields(uTxBuffWithSig []byte) error {
	inUTx.hash = inUTx.hasher.Compute(string(uTxBuffWithSig))

	inUTx.rcvShard = inUTx.coordinator.ComputeId(inUTx.uTx.RcvAddr)
	inUTx.sndShard = inUTx.coordinator.ComputeId(inUTx.uTx.SndAddr)

	isForCurrentShardRecv := inUTx.rcvShard == inUTx.coordinator.SelfId()
	isForCurrentShardSender := inUTx.sndShard == inUTx.coordinator.SelfId()
	inUTx.isForCurrentShard = isForCurrentShardRecv || isForCurrentShardSender

	return nil
}

// Nonce returns the transaction nonce
func (inUTx *InterceptedUnsignedTransaction) Nonce() uint64 {
	return inUTx.uTx.Nonce
}

// SenderAddress returns the transaction sender address
func (inUTx *InterceptedUnsignedTransaction) SenderAddress() []byte {
	return inUTx.uTx.SndAddr
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

// Fee represents the unsigned transaction fee. It is always 0
func (inUTx *InterceptedUnsignedTransaction) Fee() *big.Int {
	return big.NewInt(0)
}

// Hash gets the hash of this transaction
func (inUTx *InterceptedUnsignedTransaction) Hash() []byte {
	return inUTx.hash
}

// Type returns the type of this intercepted data
func (inUTx *InterceptedUnsignedTransaction) Type() string {
	return "intercepted unsigned tx"
}

// String returns the unsigned transaction's most important fields as string
func (inUTx *InterceptedUnsignedTransaction) String() string {
	return fmt.Sprintf("sender=%s, nonce=%d, value=%s, recv=%s, data=%s",
		logger.DisplayByteSlice(inUTx.uTx.SndAddr),
		inUTx.uTx.Nonce,
		inUTx.uTx.Value.String(),
		logger.DisplayByteSlice(inUTx.uTx.RcvAddr),
		hex.EncodeToString(inUTx.uTx.Data),
	)
}

// Identifiers returns the identifiers used in requests
func (inUTx *InterceptedUnsignedTransaction) Identifiers() [][]byte {
	return [][]byte{inUTx.hash}
}

// IsInterfaceNil returns true if there is no value under the interface
func (inUTx *InterceptedUnsignedTransaction) IsInterfaceNil() bool {
	return inUTx == nil
}
