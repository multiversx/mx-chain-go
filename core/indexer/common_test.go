package indexer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/require"
)

func createCommonProcessor() commonProcessor {
	return commonProcessor{
		addressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		validatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		txFeeCalculator: &economicsmocks.EconomicsHandlerStub{
			ComputeTxFeeBasedOnGasUsedCalled: func(tx process.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
				return big.NewInt(100)
			},
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 500
			},
		},
		shardCoordinator: &mock.ShardCoordinatorMock{},
	}
}

func TestGetMoveBalanceTransaction(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}
	status := "Success"
	gasPrice := uint64(1000)
	gasLimit := uint64(1000)
	cp := createCommonProcessor()

	tx := &transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(1000),
		RcvAddr:   []byte("receiver"),
		SndAddr:   []byte("sender"),
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte("data"),
		ChainID:   []byte("1"),
		Version:   1,
		Signature: []byte("signature"),
	}

	expectedTx := &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         tx.Nonce,
		Round:         header.Round,
		Value:         tx.Value.String(),
		Receiver:      cp.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:        cp.addressPubkeyConverter.Encode(tx.SndAddr),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      gasPrice,
		GasLimit:      gasLimit,
		GasUsed:       uint64(500),
		Data:          tx.Data,
		Signature:     hex.EncodeToString(tx.Signature),
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        status,
		rcvAddrBytes:  []byte("receiver"),
		Fee:           "100",
	}

	dbTx := cp.buildTransaction(tx, txHash, mbHash, mb, header, status)
	require.Equal(t, expectedTx, dbTx)
}

func TestGetTransactionByType_SC(t *testing.T) {
	t.Parallel()

	cp := createCommonProcessor()

	nonce := uint64(10)
	txHash := []byte("txHash")
	code := []byte("code")
	sndAddr, rcvAddr := []byte("snd"), []byte("rec")
	scHash := "scHash"
	smartContractRes := &smartContractResult.SmartContractResult{
		Nonce:      nonce,
		PrevTxHash: txHash,
		Code:       code,
		Data:       []byte(""),
		SndAddr:    sndAddr,
		RcvAddr:    rcvAddr,
		CallType:   vmcommon.CallType(1),
	}

	scRes := cp.convertScResultInDatabaseScr(scHash, smartContractRes)
	expectedTx := ScResult{
		Nonce:        nonce,
		Hash:         hex.EncodeToString([]byte(scHash)),
		PreTxHash:    hex.EncodeToString(txHash),
		Code:         string(code),
		Data:         make([]byte, 0),
		Sender:       cp.addressPubkeyConverter.Encode(sndAddr),
		Receiver:     cp.addressPubkeyConverter.Encode(rcvAddr),
		Value:        "<nil>",
		RelayedValue: "<nil>",
		CallType:     "1",
	}

	require.Equal(t, expectedTx, scRes)
}

func TestGetTransactionByType_RewardTx(t *testing.T) {
	t.Parallel()

	cp := createCommonProcessor()

	round := uint64(10)
	rcvAddr := []byte("receiver")
	rwdTx := &rewardTx.RewardTx{Round: round, RcvAddr: rcvAddr}
	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}
	status := "Success"

	resultTx := cp.buildRewardTransaction(rwdTx, txHash, mbHash, mb, header, status)
	expectedTx := &Transaction{
		Hash:     hex.EncodeToString(txHash),
		MBHash:   hex.EncodeToString(mbHash),
		Round:    round,
		Receiver: hex.EncodeToString(rcvAddr),
		Status:   status,
		Value:    "<nil>",
		Sender:   fmt.Sprintf("%d", core.MetachainShardId),
		Data:     make([]byte, 0),
	}

	require.Equal(t, expectedTx, resultTx)
}

func TestPrepareBufferMiniblocks(t *testing.T) {
	var buff bytes.Buffer

	meta := []byte("test1")
	serializedData := []byte("test2")

	buff = prepareBufferMiniblocks(buff, meta, serializedData)

	var expectedBuff bytes.Buffer
	serializedData = append(serializedData, "\n"...)
	expectedBuff.Grow(len(meta) + len(serializedData))
	_, _ = expectedBuff.Write(meta)
	_, _ = expectedBuff.Write(serializedData)

	require.Equal(t, expectedBuff, buff)
}
