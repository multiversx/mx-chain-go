package preprocess

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
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

func computeGasLimitFromResultAndRefund(result *gasConsumedResult, refundGas uint64) uint64 {
	return result.consumedReceiverShard + result.consumedSenderShard + refundGas
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
		ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
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

func Test_computeGasConsumedSelfSenderMoveBalanceIntra(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(0)
	rcvAddr, _ := hex.DecodeString("addrReceiver" + suffixShard0)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 75000,
		err:                   nil,
	}
	gasRefund := uint64(25000)

	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasConsumedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Nil(t, err)
	require.Equal(t, gasConsumed.consumedSenderShard, gasLimitSender)
	require.Equal(t, gasConsumed.consumedSenderShard, gasLimitReceiver)
}

func Test_computeGasConsumedSelfSenderSCCallIntra(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(0)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard0)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   500000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}
	gasRefund := uint64(25000)

	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("sc invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasConsumedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Nil(t, err)
	require.Equal(t, gasConsumed.consumedSenderShard-gasRefund, gasLimitSender)
	require.Equal(t, gasConsumed.consumedSenderShard-gasRefund, gasLimitReceiver)
}

func Test_computeGasConsumedByTxSelfSenderMoveBalanceCross(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString("addrReceiver" + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 75000,
		err:                   nil,
	}
	gasRefund := uint64(25000)

	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasConsumedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Nil(t, err)
	require.Equal(t, gasConsumed.consumedSenderShard, gasLimitSender)
	require.Equal(t, gasConsumed.consumedReceiverShard, gasLimitReceiver)
}

func Test_computeGasConsumedByTxSelfSenderScCallCross(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   500000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasConsumedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Nil(t, err)
	require.Equal(t, gasConsumed.consumedSenderShard, gasLimitSender)
	require.Equal(t, gasConsumed.consumedReceiverShard, gasLimitReceiver)
}

func Test_computeGasConsumedByTxGasHandlerComputeGasErrors(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   500000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	expectedError := errors.New("expecterd error")
	gt.gasHandler = &testscommon.GasHandlerStub{
		ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, 0, expectedError
		},
	}

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasConsumedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Equal(t, expectedError, err)
	require.Zero(t, gasLimitSender)
	require.Zero(t, gasLimitReceiver)
}

func Test_computeGasConsumedByTxGasHandlerRefundGasLargerThanLimit(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   500000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	gc := uint64(1000000)
	gt.gasHandler = &testscommon.GasHandlerStub{
		ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return gc, gc, nil
		},
		GasRefundedCalled: func(hash []byte) uint64 {
			return gc * 2
		},
	}

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gasLimitSender, gasLimitReceiver, err := gt.computeGasConsumedByTx(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
	)

	require.Equal(t, process.ErrInsufficientGasLimitInTx, err)
	require.Zero(t, gasLimitSender)
	require.Zero(t, gasLimitReceiver)
}

func Test_computeGasConsumedWithErrorForGasConsumedForTx(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 1600000000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	expectedError := errors.New("expecterd error")
	gt.gasHandler = &testscommon.GasHandlerStub{
		ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
			return 0, 0, expectedError
		},
	}

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{}
	err := gt.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Equal(t, expectedError, err)
}

func Test_computeGasConsumedMaxGasLimitInSenderShardReached(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{
		gasConsumedByMiniBlockInReceiverShard: gt.economicsFee.MaxGasLimitPerBlock(receiverShardID) - gasConsumed.consumedReceiverShard/2,
	}
	err := gt.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Equal(t, process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached, err)
}

func Test_computeGasConsumedMaxGasLimitInReceiverShardReached(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(receiverShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{
		gasConsumedByMiniBlocksInSenderShard: gt.economicsFee.MaxGasLimitPerBlock(senderShardID) - gasConsumed.consumedSenderShard/2,
	}
	err := gt.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Equal(t, process.ErrMaxGasLimitPerMiniBlockInSenderShardIsReached, err)
}

func Test_computeGasConsumedMaxGasLimitPerBlockReached(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{
		totalGasConsumedInSelfShard: gt.economicsFee.MaxGasLimitPerBlock(senderShardID) - gasConsumed.consumedSenderShard/2,
	}
	err := gt.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Equal(t, process.ErrMaxGasLimitPerBlockInSelfShardIsReached, err)
}

func Test_computeGasConsumedOK(t *testing.T) {
	t.Parallel()

	senderShardID := uint32(0)
	sndAddr, _ := hex.DecodeString("addrSender" + suffixShard0)
	receiverShardID := uint32(1)
	rcvAddr, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	hasher := &hashingMocks.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	gasConsumed := &gasConsumedResult{
		consumedSenderShard:   75000,
		consumedReceiverShard: 500000,
		err:                   nil,
	}

	gasRefund := uint64(25000)
	gasLimit := computeGasLimitFromResultAndRefund(gasConsumed, gasRefund)
	gt := createDefaultGasTracker(senderShardID, gasConsumed, gasRefund)
	tx := createDefaultTx(sndAddr, rcvAddr, gasLimit)
	tx.Data = []byte("tx invoking data")

	txm, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txm))

	gci := &gasConsumedInfo{}
	err := gt.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		gci,
	)
	require.Nil(t, err)
	require.Equal(t, gasConsumed.consumedSenderShard, gci.gasConsumedByMiniBlocksInSenderShard)
	require.Equal(t, gasConsumed.consumedReceiverShard, gci.gasConsumedByMiniBlockInReceiverShard)
	require.Equal(t, gasConsumed.consumedSenderShard, gci.totalGasConsumedInSelfShard)
}
