package transaction

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.TxValidatorHandler = (*InterceptedTransaction)(nil)
var _ process.InterceptedData = (*InterceptedTransaction)(nil)

// InterceptedTransaction holds and manages a transaction based struct with extended functionality
type InterceptedTransaction struct {
	tx                     *transaction.Transaction
	protoMarshalizer       marshal.Marshalizer
	signMarshalizer        marshal.Marshalizer
	hasher                 hashing.Hasher
	txSignHasher           hashing.Hasher
	keyGen                 crypto.KeyGenerator
	singleSigner           crypto.SingleSigner
	pubkeyConv             core.PubkeyConverter
	coordinator            sharding.Coordinator
	hash                   []byte
	feeHandler             process.FeeHandler
	whiteListerVerifiedTxs process.WhiteListHandler
	argsParser             process.ArgumentsParser
	txVersionChecker       process.TxVersionCheckerHandler
	chainID                []byte
	rcvShard               uint32
	sndShard               uint32
	isForCurrentShard      bool
	enableSignedTxWithHash bool
	enableEpochsHandler    common.EnableEpochsHandler
}

// NewInterceptedTransaction returns a new instance of InterceptedTransaction
func NewInterceptedTransaction(
	txBuff []byte,
	protoMarshalizer marshal.Marshalizer,
	signMarshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	keyGen crypto.KeyGenerator,
	signer crypto.SingleSigner,
	pubkeyConv core.PubkeyConverter,
	coordinator sharding.Coordinator,
	feeHandler process.FeeHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	argsParser process.ArgumentsParser,
	chainID []byte,
	enableSignedTxWithHash bool,
	txSignHasher hashing.Hasher,
	txVersionChecker process.TxVersionCheckerHandler,
	enableEpochsHandler common.EnableEpochsHandler,
) (*InterceptedTransaction, error) {

	if txBuff == nil {
		return nil, process.ErrNilBuffer
	}
	if check.IfNil(protoMarshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(signMarshalizer) {
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
	if check.IfNil(pubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(feeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(whiteListerVerifiedTxs) {
		return nil, process.ErrNilWhiteListHandler
	}
	if check.IfNil(argsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if len(chainID) == 0 {
		return nil, process.ErrInvalidChainID
	}
	if check.IfNil(txSignHasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(txVersionChecker) {
		return nil, process.ErrNilTransactionVersionChecker
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	tx, err := createTx(protoMarshalizer, txBuff)
	if err != nil {
		return nil, err
	}

	inTx := &InterceptedTransaction{
		tx:                     tx,
		protoMarshalizer:       protoMarshalizer,
		signMarshalizer:        signMarshalizer,
		hasher:                 hasher,
		singleSigner:           signer,
		pubkeyConv:             pubkeyConv,
		keyGen:                 keyGen,
		coordinator:            coordinator,
		feeHandler:             feeHandler,
		whiteListerVerifiedTxs: whiteListerVerifiedTxs,
		argsParser:             argsParser,
		chainID:                chainID,
		enableSignedTxWithHash: enableSignedTxWithHash,
		txVersionChecker:       txVersionChecker,
		txSignHasher:           txSignHasher,
		enableEpochsHandler:    enableEpochsHandler,
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

func createRelayedV2(relayedTx *transaction.Transaction, args [][]byte) (*transaction.Transaction, error) {
	if len(args) != 4 {
		return nil, process.ErrInvalidArguments
	}
	tx := &transaction.Transaction{
		Nonce:             big.NewInt(0).SetBytes(args[1]).Uint64(),
		Value:             big.NewInt(0),
		RcvAddr:           args[0],
		SndAddr:           relayedTx.RcvAddr,
		GasPrice:          relayedTx.GasPrice,
		GasLimit:          0, // the user had to sign a transaction with 0 gasLimit - as all gasLimit is coming from the relayer
		Data:              args[2],
		ChainID:           relayedTx.ChainID,
		Version:           relayedTx.Version,
		Signature:         args[3],
		Options:           relayedTx.Options,
		GuardianAddr:      nil,
		GuardianSignature: nil,
	}

	return tx, nil
}

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (inTx *InterceptedTransaction) CheckValidity() error {
	err := inTx.integrity(inTx.tx)
	if err != nil {
		return err
	}

	whiteListedVerified := inTx.whiteListerVerifiedTxs.IsWhiteListed(inTx)
	if !whiteListedVerified {
		err = inTx.verifySig(inTx.tx)
		if err != nil {
			return err
		}

		err = inTx.VerifyGuardianSig(inTx.tx)
		if err != nil {
			return err
		}

		err = inTx.verifyIfRelayedTx(inTx.tx)
		if err != nil {
			return err
		}

		err = inTx.verifyIfRelayedTxV2(inTx.tx)
		if err != nil {
			return err
		}

		err = inTx.verifyIfRelayedTxV3(inTx.tx)
		if err != nil {
			return err
		}

		inTx.whiteListerVerifiedTxs.Add([][]byte{inTx.Hash()})
	}

	return nil
}

func isRelayedTx(funcName string, innerTxs []*transaction.Transaction) bool {
	return core.RelayedTransaction == funcName ||
		core.RelayedTransactionV2 == funcName ||
		len(innerTxs) > 0
}

func (inTx *InterceptedTransaction) verifyIfRelayedTxV3(tx *transaction.Transaction) error {
	if len(tx.InnerTransactions) == 0 {
		return nil
	}
	if !inTx.enableEpochsHandler.IsRelayedTransactionsV3FlagEnabled() {
		return process.ErrRelayedTxV3Disabled
	}

	innerTxs := tx.InnerTransactions
	for _, innerTx := range innerTxs {
		if !bytes.Equal(innerTx.SndAddr, tx.RcvAddr) {
			return process.ErrRelayedTxV3BeneficiaryDoesNotMatchReceiver
		}
		if len(innerTx.RelayerAddr) == 0 {
			return process.ErrRelayedTxV3EmptyRelayer
		}

		err := inTx.integrity(innerTx)
		if err != nil {
			return fmt.Errorf("inner transaction: %w", err)
		}

		err = inTx.verifyUserTx(innerTx)
		if err != nil {
			return fmt.Errorf("inner transaction: %w", err)
		}
	}

	return nil
}

func (inTx *InterceptedTransaction) verifyIfRelayedTxV2(tx *transaction.Transaction) error {
	funcName, userTxArgs, err := inTx.argsParser.ParseCallData(string(tx.Data))
	if err != nil {
		return nil
	}
	if core.RelayedTransactionV2 != funcName {
		return nil
	}

	userTx, err := createRelayedV2(tx, userTxArgs)
	if err != nil {
		return err
	}

	return inTx.verifyUserTx(userTx)
}

func (inTx *InterceptedTransaction) verifyIfRelayedTx(tx *transaction.Transaction) error {
	funcName, userTxArgs, err := inTx.argsParser.ParseCallData(string(tx.Data))
	if err != nil {
		return nil
	}
	if core.RelayedTransaction != funcName {
		return nil
	}

	if len(userTxArgs) != 1 {
		return process.ErrInvalidArguments
	}

	userTx, err := createTx(inTx.signMarshalizer, userTxArgs[0])
	if err != nil {
		return fmt.Errorf("inner transaction: %w", err)
	}

	if !bytes.Equal(userTx.SndAddr, tx.RcvAddr) {
		return process.ErrRelayedTxBeneficiaryDoesNotMatchReceiver
	}

	err = inTx.integrity(userTx)
	if err != nil {
		return fmt.Errorf("inner transaction: %w", err)
	}

	return inTx.verifyUserTx(userTx)
}

func (inTx *InterceptedTransaction) verifyUserTx(userTx *transaction.Transaction) error {
	err := inTx.verifySig(userTx)
	if err != nil {
		return fmt.Errorf("inner transaction: %w", err)
	}

	err = inTx.VerifyGuardianSig(userTx)
	if err != nil {
		return fmt.Errorf("inner transaction: %w", err)
	}

	funcName, _, err := inTx.argsParser.ParseCallData(string(userTx.Data))
	if err != nil {
		return nil
	}

	// recursive relayed transactions are not allowed
	if isRelayedTx(funcName, userTx.InnerTransactions) {
		return process.ErrRecursiveRelayedTxIsNotAllowed
	}

	return nil
}

func (inTx *InterceptedTransaction) processFields(txBuff []byte) error {
	inTx.hash = inTx.hasher.Compute(string(txBuff))

	inTx.sndShard = inTx.coordinator.ComputeId(inTx.tx.SndAddr)
	emptyAddr := make([]byte, len(inTx.tx.RcvAddr))
	inTx.rcvShard = inTx.coordinator.ComputeId(inTx.tx.RcvAddr)
	if bytes.Equal(inTx.tx.RcvAddr, emptyAddr) {
		inTx.rcvShard = inTx.sndShard
	}

	isForCurrentShardRecv := inTx.rcvShard == inTx.coordinator.SelfId()
	isForCurrentShardSender := inTx.sndShard == inTx.coordinator.SelfId()
	inTx.isForCurrentShard = isForCurrentShardRecv || isForCurrentShardSender

	return nil
}

// integrity checks for not nil fields and negative value
func (inTx *InterceptedTransaction) integrity(tx *transaction.Transaction) error {
	err := inTx.txVersionChecker.CheckTxVersion(tx)
	if err != nil {
		return err
	}

	err = tx.CheckIntegrity()
	if err != nil {
		return err
	}

	if !bytes.Equal(tx.ChainID, inTx.chainID) {
		return process.ErrInvalidChainID
	}
	if len(tx.RcvAddr) != inTx.pubkeyConv.Len() {
		return process.ErrInvalidRcvAddr
	}
	if len(tx.SndAddr) != inTx.pubkeyConv.Len() {
		return process.ErrInvalidSndAddr
	}

	err = inTx.checkMaxGasPrice()
	if err != nil {
		return err
	}

	return inTx.feeHandler.CheckValidityTxValues(tx)
}

func (inTx *InterceptedTransaction) checkMaxGasPrice() error {
	tx := inTx.tx
	// no need to check max gas for guarded transactions as they are co-signed
	if inTx.txVersionChecker.IsGuardedTransaction(tx) {
		return nil
	}

	txData := tx.GetData()
	if !process.IsSetGuardianCall(txData) {
		return nil
	}

	if tx.GetGasPrice() > inTx.feeHandler.MaxGasPriceSetGuardian() {
		return process.ErrGasPriceTooHigh
	}

	return nil
}

// verifySig checks if the tx is correctly signed
func (inTx *InterceptedTransaction) verifySig(tx *transaction.Transaction) error {
	txMessageForSigVerification, err := inTx.getTxMessageForGivenTx(tx)
	if err != nil {
		return err
	}

	senderPubKey, err := inTx.keyGen.PublicKeyFromByteArray(tx.SndAddr)
	if err != nil {
		return err
	}

	return inTx.singleSigner.Verify(senderPubKey, txMessageForSigVerification, tx.Signature)
}

// VerifyGuardianSig verifies if the guardian signature is valid
func (inTx *InterceptedTransaction) VerifyGuardianSig(tx *transaction.Transaction) error {
	txMessageForSigVerification, err := inTx.getTxMessageForGivenTx(tx)
	if err != nil {
		return err
	}

	if !inTx.txVersionChecker.IsGuardedTransaction(tx) {
		return verifyConsistencyForNotGuardedTx(tx)
	}

	guardianPubKey, err := inTx.keyGen.PublicKeyFromByteArray(tx.GuardianAddr)
	if err != nil {
		return err
	}

	errVerifySig := inTx.singleSigner.Verify(guardianPubKey, txMessageForSigVerification, tx.GuardianSignature)
	if errVerifySig != nil {
		return fmt.Errorf("%w when checking the guardian's signature", errVerifySig)
	}

	return nil
}

func verifyConsistencyForNotGuardedTx(tx *transaction.Transaction) error {
	if len(tx.GetGuardianAddr()) > 0 {
		return process.ErrGuardianAddressNotExpected
	}
	if len(tx.GetGuardianSignature()) > 0 {
		return process.ErrGuardianSignatureNotExpected
	}

	return nil
}

func (inTx *InterceptedTransaction) getTxMessageForGivenTx(tx *transaction.Transaction) ([]byte, error) {
	if inTx.txVersionChecker.IsSignedWithHash(tx) && !inTx.enableSignedTxWithHash {
		return nil, process.ErrTransactionSignedWithHashIsNotEnabled
	}

	txSigningData, err := tx.GetDataForSigning(inTx.pubkeyConv, inTx.signMarshalizer, inTx.txSignHasher)
	if err != nil {
		return nil, err
	}

	return txSigningData, nil
}

// GetTxMessageForSignatureVerification returns the transaction data that the signature needs to be verified on
func (inTx *InterceptedTransaction) GetTxMessageForSignatureVerification() ([]byte, error) {
	return inTx.getTxMessageForGivenTx(inTx.tx)
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
func (inTx *InterceptedTransaction) SenderAddress() []byte {
	return inTx.tx.SndAddr
}

// Fee returns the estimated cost of the transaction
func (inTx *InterceptedTransaction) Fee() *big.Int {
	return inTx.feeHandler.ComputeTxFee(inTx.tx)
}

// RelayerAddress returns the relayer address from transaction
func (inTx *InterceptedTransaction) RelayerAddress() []byte {
	return inTx.tx.RelayerAddr
}

// Type returns the type of this intercepted data
func (inTx *InterceptedTransaction) Type() string {
	return "intercepted tx"
}

// String returns the transaction's most important fields as string
func (inTx *InterceptedTransaction) String() string {
	return fmt.Sprintf("sender=%s, nonce=%d, value=%s, recv=%s, data=%s",
		logger.DisplayByteSlice(inTx.tx.SndAddr),
		inTx.tx.Nonce,
		inTx.tx.Value.String(),
		logger.DisplayByteSlice(inTx.tx.RcvAddr),
		hex.EncodeToString(inTx.tx.Data),
	)
}

// Identifiers returns the identifiers used in requests
func (inTx *InterceptedTransaction) Identifiers() [][]byte {
	return [][]byte{inTx.hash}
}

// IsInterfaceNil returns true if there is no value under the interface
func (inTx *InterceptedTransaction) IsInterfaceNil() bool {
	return inTx == nil
}
