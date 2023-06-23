package preprocess

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/require"
)

func moveBalanceGas(data []byte) uint64 {
	return uint64(50000 + 1500*len(data))
}

type gasConsumedResult struct {
	consumedSenderShard   uint64
	consumedReceiverShard uint64
	err                   error
}

const (
	smartContractAddressStart = "00000000000000000000"
	suffixShard0              = "0000"
	suffixShard1              = "0001"
)

func computeGasLimitFromResultAndRefund(gcr *gasConsumedResult, refundGas uint64) uint64 {
	return gcr.consumedReceiverShard + gcr.consumedSenderShard + refundGas
}

func createDefaultTx(
	sndAddr []byte,
	rcvAddr []byte,
	gasLimit uint64,
) *transaction.Transaction {
	return &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(10000),
		RcvAddr:   rcvAddr,
		SndAddr:   sndAddr,
		GasPrice:  100000000,
		GasLimit:  gasLimit,
		Data:      []byte("data"),
		ChainID:   []byte("T"),
		Version:   0,
		Signature: []byte("signature"),
		Options:   0,
	}
}

func createDefaultGasTracker(
	selfShardID uint32,
	gcr *gasConsumedResult,
	gasRefunded uint64,
) *gasTracker {
	shardCoordinator := &testscommon.ShardsCoordinatorMock{
		CurrentShard: selfShardID,
	}
	economicsFee := &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
			return 1500000000
		},
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return moveBalanceGas(tx.GetData())
		},
	}

	gasHandler := &testscommon.GasHandlerStub{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return gcr.consumedSenderShard, gcr.consumedReceiverShard, gcr.err
		},
		GasRefundedCalled: func(hash []byte) uint64 {
			return gasRefunded
		},
	}

	gt := &gasTracker{
		shardCoordinator: shardCoordinator,
		economicsFee:     economicsFee,
		gasHandler:       gasHandler,
	}

	return gt
}

func Test_computeGasProvidedSelfSenderMoveBalanceIntra(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(0)
	rcvAddr, _ := hex.DecodeString("addrReceiver" + suffixShard0)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 75000,
		err:                   nil,
	}
	gasRefund := uint64(25000)

	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasProvidedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Nil(t, err)
	require.Equal(t, gcr.consumedSenderShard, gasLimitSender)
	require.Equal(t, gcr.consumedSenderShard, gasLimitReceiver)
}

func Test_computeGasProvidedSelfSenderSCCallIntra(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(0)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard0)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   500000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}
	gasRefund := uint64(25000)

	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("sc invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasProvidedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Nil(t, err)
	require.Equal(t, gcr.consumedSenderShard-gasRefund, gasLimitSender)
	require.Equal(t, gcr.consumedSenderShard-gasRefund, gasLimitReceiver)
}

func Test_computeGasProvidedByTxSelfSenderMoveBalanceCross(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString("addrReceiver" + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 75000,
		err:                   nil,
	}
	gasRefund := uint64(25000)

	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasProvidedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Nil(t, err)
	require.Equal(t, gcr.consumedSenderShard, gasLimitSender)
	require.Equal(t, gcr.consumedReceiverShard, gasLimitReceiver)
}

func Test_computeGasProvidedByTxSelfSenderScCallCross(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   500000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasProvidedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Nil(t, err)
	require.Equal(t, gcr.consumedSenderShard, gasLimitSender)
	require.Equal(t, gcr.consumedReceiverShard, gasLimitReceiver)
}

func Test_computeGasProvidedByTxGasHandlerComputeGasErrors(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   500000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	expectedError := errors.New("expecterd error")
	gt.gasHandler = &testscommon.GasHandlerStub{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, 0, expectedError
		},
	}

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasProvidedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Equal(t, expectedError, err)
	require.Zero(t, gasLimitSender)
	require.Zero(t, gasLimitReceiver)
}

func Test_computeGasProvidedByTxGasHandlerRefundGasLargerThanLimit(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   500000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	gc := uint64(1000000)
	gt.gasHandler = &testscommon.GasHandlerStub{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return gc, gc, nil
		},
		GasRefundedCalled: func(hash []byte) uint64 {
			return gc * 2
		},
	}

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasProvidedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Equal(t, process.ErrInsufficientGasLimitInTx, err)
	require.Zero(t, gasLimitSender)
	require.Zero(t, gasLimitReceiver)
}

func Test_computeGasProvidedWithErrorForGasConsumedForTx(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 1600000000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	expectedError := errors.New("expecterd error")
	gt.gasHandler = &testscommon.GasHandlerStub{
		ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, 0, expectedError
		},
	}

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{}
	_, err := gt.computeGasProvided(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Equal(t, expectedError, err)
}

func Test_computeGasProvidedMaxGasLimitInSenderShardReached(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{
		gasConsumedByMiniBlockInReceiverShard: gt.economicsFee.MaxGasLimitPerBlock(receiverShardID) - gcr.consumedReceiverShard/2,
	}
	_, err := gt.computeGasProvided(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Equal(t, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached, err)
}

func Test_computeGasProvidedMaxGasLimitInReceiverShardReached(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(receiverShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{
		gasConsumedByMiniBlocksInSenderShard: gt.economicsFee.MaxGasLimitPerBlock(senderShardID) - gcr.consumedSenderShard/2,
	}
	_, err := gt.computeGasProvided(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Equal(t, nil, err)
}

func Test_computeGasProvidedMaxGasLimitPerBlockReached(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{
		totalGasConsumedInSelfShard: gt.economicsFee.MaxGasLimitPerBlock(senderShardID) - gcr.consumedSenderShard/2,
	}
	_, err := gt.computeGasProvided(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Equal(t, process.ErrMaxGasLimitPerBlockInSelfShardIsReached, err)
}

func Test_computeGasProvidedOK(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
	gcr := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gcr, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gcr, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{}
	_, err := gt.computeGasProvided(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Nil(t, err)
	require.Equal(t, gcr.consumedSenderShard, gci.gasConsumedByMiniBlocksInSenderShard)
	require.Equal(t, gcr.consumedReceiverShard, gci.gasConsumedByMiniBlockInReceiverShard)
	require.Equal(t, gcr.consumedSenderShard, gci.totalGasConsumedInSelfShard)
}
