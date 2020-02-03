package indexer

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/stretchr/testify/require"
)

func TestGetTransactionByType_SC(t *testing.T) {
	t.Parallel()

	nonce := uint64(10)
	smartContract := &smartContractResult.SmartContractResult{Nonce: nonce}
	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	blockHash := []byte("blockHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}

	resultTx := getTransactionByType(smartContract, txHash, mbHash, blockHash, mb, header, "")
	require.Equal(t, nonce, resultTx.Nonce)
}

func TestGetTransactionByType_RewardTx(t *testing.T) {
	t.Parallel()

	round := uint64(10)
	rcvAddr := []byte("receiver")
	rwdTx := &rewardTx.RewardTx{Round: round, RcvAddr: rcvAddr}
	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	blockHash := []byte("blockHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}

	resultTx := getTransactionByType(rwdTx, txHash, mbHash, blockHash, mb, header, "")
	require.Equal(t, round, resultTx.Round)
}

func TestGetTransactionByType_Receipt(t *testing.T) {
	t.Parallel()

	receiptTest := &receipt.Receipt{Value: big.NewInt(100)}
	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	blockHash := []byte("blockHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}

	resultTx := getTransactionByType(receiptTest, txHash, mbHash, blockHash, mb, header, "")
	require.Equal(t, receiptTest.Value.String(), resultTx.Value)
}

func TestGetTransactionByType_Nil(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	blockHash := []byte("blockHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}

	resultTx := getTransactionByType(nil, txHash, mbHash, blockHash, mb, header, "")
	require.Nil(t, resultTx)
}
