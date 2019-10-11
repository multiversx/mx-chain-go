package transaction

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedTransaction holds and manages a transaction based struct with extended functionality
type InterceptedTransaction struct {
	tx                *transaction.Transaction
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	keyGen            crypto.KeyGenerator
	singleSigner      crypto.SingleSigner
	addrConv          state.AddressConverter
	coordinator       sharding.Coordinator
	hash              []byte
	rcvShard          uint32
	sndShard          uint32
	isForCurrentShard bool
	sndAddr           state.AddressContainer
	tx                       *transaction.Transaction
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	keyGen                   crypto.KeyGenerator
	singleSigner             crypto.SingleSigner
	addrConv                 state.AddressConverter
	coordinator              sharding.Coordinator
	hash                     []byte
	rcvShard                 uint32
	sndShard                 uint32
	isAddressedToOtherShards bool
	sndAddr                  state.AddressContainer
	feeHandler               process.FeeHandler
}

// NewInterceptedTransaction returns a new instance of InterceptedTransaction
func NewInterceptedTransaction(
	txBuff []byte,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	keyGen crypto.KeyGenerator,
	signer crypto.SingleSigner,
	addrConv state.AddressConverter,
	coordinator sharding.Coordinator,
	feeHandler process.FeeHandler,
) (*InterceptedTransaction, error) {

	if txBuff == nil {
		return nil, process.ErrNilBuffer
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(keyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(signer) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(addrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if feeHandler == nil || coordinator.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	tx, err := createTx(marshalizer, txBuff)
	if err != nil {
		return nil, err
	}

	inTx := &InterceptedTransaction{
		tx:           tx,
		marshalizer:  marshalizer,
		hasher:       hasher,
		singleSigner: signer,
		addrConv:     addrConv,
		keyGen:       keyGen,
		coordinator:  coordinator,
		feeHandler:   feeHandler,
	}

	err = inTx.processFields(txBuff)
	if err != nil {
		return nil, err
	}

	return inTx, nil
}

func createTx(marshalizer marshal.Marshalizer, txBuff []byte) (*transaction.Transaction, error) {
	tx := &transaction.Transaction{}
	err := marshalizer.Unmarshal(tx, txBuff)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (inTx *InterceptedTransaction) CheckValidity() error {
	err := inTx.integrity()
	if err != nil {
		return err
	}

	err = inTx.verifySig()
	if err != nil {
		return err
	}

	return nil
}

func (inTx *InterceptedTransaction) processFields(txBuff []byte) error {
	inTx.hash = inTx.hasher.Compute(string(txBuff))

	var err error
	inTx.sndAddr, err = inTx.addrConv.CreateAddressFromPublicKeyBytes(inTx.tx.SndAddr)
	if err != nil {
		return process.ErrInvalidSndAddr
	}

	rcvAddr, err := inTx.addrConv.CreateAddressFromPublicKeyBytes(inTx.tx.RcvAddr)
	if err != nil {
		return process.ErrInvalidRcvAddr
	}

	inTx.sndShard = inTx.coordinator.ComputeId(inTx.sndAddr)
	emptyAddr := make([]byte, len(rcvAddr.Bytes()))
	inTx.rcvShard = inTx.coordinator.ComputeId(rcvAddr)
	if bytes.Equal(rcvAddr.Bytes(), emptyAddr) {
		inTx.rcvShard = inTx.sndShard
	}

	isForCurrentShardRecv := inTx.rcvShard == inTx.coordinator.SelfId()
	isForCurrentShardSender := inTx.sndShard == inTx.coordinator.SelfId()
	inTx.isForCurrentShard = isForCurrentShardRecv || isForCurrentShardSender

	return nil
}

// integrity checks for not nil fields and negative value
func (inTx *InterceptedTransaction) integrity() error {
	if inTx.tx.Signature == nil {
		return process.ErrNilSignature
	}
	if inTx.tx.RcvAddr == nil {
		return process.ErrNilRcvAddr
	}
	if inTx.tx.SndAddr == nil {
		return process.ErrNilSndAddr
	}
	if inTx.tx.Value == nil {
		return process.ErrNilValue
	}
	if inTx.tx.Value.Cmp(big.NewInt(0)) < 0 {
		return process.ErrNegativeValue
	}

	return inTx.checkFeeValues()
}

func (inTx *InterceptedTransaction) checkFeeValues() error {
	isLowerGasLimitInTx := inTx.tx.GasLimit < inTx.feeHandler.MinGasLimit()
	if isLowerGasLimitInTx {
		return process.ErrInsufficientGasLimitInTx
	}

	isLowerGasPrice := inTx.tx.GasPrice < inTx.feeHandler.MinGasPrice()
	if isLowerGasPrice {
		return process.ErrInsufficientGasPriceInTx
	}

	return nil
}

// verifySig checks if the tx is correctly signed
func (inTx *InterceptedTransaction) verifySig() error {
	copiedTx := *inTx.tx
	copiedTx.Signature = nil
	buffCopiedTx, err := inTx.marshalizer.Marshal(&copiedTx)
	if err != nil {
		return err
	}

	senderPubKey, err := inTx.keyGen.PublicKeyFromByteArray(inTx.tx.SndAddr)
	if err != nil {
		return err
	}

	err = inTx.singleSigner.Verify(senderPubKey, buffCopiedTx, inTx.tx.Signature)
	if err != nil {
		return err
	}

	return nil
}

// ReceiverShardId returns the receiver shard id
func (inTx *InterceptedTransaction) ReceiverShardId() uint32 {
	return inTx.rcvShard
}

// IsForCurrentShard returns true if this transaction is meant to be processed by the node from this shard
func (inTx *InterceptedTransaction) IsForCurrentShard() bool {
	return inTx.isForCurrentShard
}

// Transaction returns the transaction pointer that actually holds the data
func (inTx *InterceptedTransaction) Transaction() data.TransactionHandler {
	return inTx.tx
}

// Hash gets the hash of this transaction
func (inTx *InterceptedTransaction) Hash() []byte {
	return inTx.hash
}

// SenderShardId returns the transaction sender shard id
func (inTx *InterceptedTransaction) SenderShardId() uint32 {
	return inTx.sndShard
}

// Nonce returns the transaction nonce
func (inTx *InterceptedTransaction) Nonce() uint64 {
	return inTx.tx.Nonce
}

// SenderAddress returns the transaction sender address
func (inTx *InterceptedTransaction) SenderAddress() state.AddressContainer {
	return inTx.sndAddr
}

// TotalValue returns the maximum cost of transaction
// totalValue = txValue + gasPrice*gasLimit
func (inTx *InterceptedTransaction) TotalValue() *big.Int {
	result := big.NewInt(0).Set(inTx.tx.Value)
	gasPrice := big.NewInt(int64(inTx.tx.GasPrice))
	gasLimit := big.NewInt(int64(inTx.tx.GasLimit))
	mulTxCost := big.NewInt(0)
	mulTxCost = mulTxCost.Mul(gasPrice, gasLimit)
	result = result.Add(result, mulTxCost)

	return result
}

// IsInterfaceNil returns true if there is no value under the interface
func (inTx *InterceptedTransaction) IsInterfaceNil() bool {
	if inTx == nil {
		return true
	}
	return false
}
