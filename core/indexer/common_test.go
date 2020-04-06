package indexer

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
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
	expectedTx := &Transaction{
		Hash:      hex.EncodeToString(txHash),
		MBHash:    hex.EncodeToString(mbHash),
		BlockHash: hex.EncodeToString(blockHash),
		Nonce:     nonce,
		Value:     "<nil>",
		Status:    "Success",
	}
	require.Equal(t, expectedTx, resultTx)
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
	status := "Success"

	resultTx := getTransactionByType(rwdTx, txHash, mbHash, blockHash, mb, header, status)
	expectedTx := &Transaction{
		Hash:      hex.EncodeToString(txHash),
		MBHash:    hex.EncodeToString(mbHash),
		BlockHash: hex.EncodeToString(blockHash),
		Round:     round,
		Receiver:  hex.EncodeToString(rcvAddr),
		Status:    status,
		Value:     "<nil>",
		Sender:    fmt.Sprintf("%d", core.MetachainShardId),
		Data:      []byte(""),
	}
	require.Equal(t, expectedTx, resultTx)
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
	expectedTx := &Transaction{
		Hash:      hex.EncodeToString(txHash),
		MBHash:    hex.EncodeToString(mbHash),
		BlockHash: hex.EncodeToString(blockHash),
		Value:     receiptTest.Value.String(),
		Status:    "Success",
	}
	require.Equal(t, expectedTx, resultTx)
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

func generateTxs(numTxs int) map[string]data.TransactionHandler {
	txs := make(map[string]data.TransactionHandler, numTxs)
	for i := 0; i < numTxs; i++ {
		tx := &transaction.Transaction{
			Nonce:     uint64(i),
			Value:     big.NewInt(int64(i)),
			RcvAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			SndAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			GasPrice:  10000000,
			GasLimit:  1000,
			Data:      []byte("dasjdksakjdksajdjksajkdjkasjdksajk"),
			Signature: []byte("randomSignatureasdasldkasdsahjgdlhjaskldsjkaldjklasjkdjskladjkl;sajkl"),
		}
		txs[fmt.Sprintf("%d", i)] = tx
	}

	return txs
}

func TestMarshalTxs(t *testing.T) {
	const KILOBYTE = 1024
	numTxs := 20000

	start := time.Now()
	txs := generateTxs(numTxs)
	elapsed := time.Since(start)
	fmt.Printf("Generate %d txs took %s \n", numTxs, elapsed)

	marshalizer := &mock.MarshalizerMock{}

	start = time.Now()
	lenTxs := calculateSizeOfTxs(marshalizer, txs)
	elapsed = time.Since(start)

	fmt.Printf("Marhsal %d txs took %s \n", numTxs, elapsed)
	fmt.Printf("Size of %d transactions : %d Kbs \n", numTxs, lenTxs/KILOBYTE)
}
