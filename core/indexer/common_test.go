package indexer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func createCommonProcessor() commonProcessor {
	return commonProcessor{
		addressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		validatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
	}
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
