package indexer

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/assert"
)

func TestPrepareTransactionsForDatabase(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}
	txHash2 := []byte("txHash2")
	tx2 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}
	txHash3 := []byte("txHash3")
	tx3 := &transaction.Transaction{}
	txHash4 := []byte("txHash4")
	tx4 := &transaction.Transaction{}
	txHash5 := []byte("txHash5")
	tx5 := &transaction.Transaction{}

	rTx1Hash := []byte("rTxHash1")
	rTx1 := &rewardTx.RewardTx{}
	rTx2Hash := []byte("rTxHash2")
	rTx2 := &rewardTx.RewardTx{}

	recHash1 := []byte("recHash1")
	rec1 := &receipt.Receipt{
		Value:  big.NewInt(100),
		TxHash: txHash1,
	}
	recHash2 := []byte("recHash2")
	rec2 := &receipt.Receipt{
		Value:  big.NewInt(200),
		TxHash: txHash2,
	}

	scHash1 := []byte("scHash1")
	scResult1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}
	scHash2 := []byte("scHash2")
	scResult2 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}
	scHash3 := []byte("scHash3")
	scResult3 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash3,
		Data:           []byte("@" + "6F6B"),
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1, txHash2, txHash3},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{txHash4},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{scHash1, scHash2},
				Type:     block.SmartContractResultBlock,
			},
			{
				TxHashes: [][]byte{scHash3},
				Type:     block.SmartContractResultBlock,
			},
			{
				TxHashes: [][]byte{recHash1, recHash2},
				Type:     block.ReceiptBlock,
			},
			{
				TxHashes: [][]byte{rTx1Hash, rTx2Hash},
				Type:     block.RewardsBlock,
			},
			{
				TxHashes: [][]byte{txHash5},
				Type:     block.InvalidBlock,
			},
		},
	}
	header := &block.Header{}
	txPool := map[string]data.TransactionHandler{
		string(txHash1):  tx1,
		string(txHash2):  tx2,
		string(txHash3):  tx3,
		string(txHash4):  tx4,
		string(txHash5):  tx5,
		string(rTx1Hash): rTx1,
		string(rTx2Hash): rTx2,
		string(recHash1): rec1,
		string(recHash2): rec2,
		string(scHash1):  scResult1,
		string(scHash2):  scResult2,
		string(scHash3):  scResult3,
	}

	txDbProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		&config.FeeSettings{},
	)

	transactions, _ := txDbProc.prepareTransactionsForDatabase(body, header, txPool, 0)
	assert.Equal(t, 7, len(transactions))

}

func TestPrepareTxLog(t *testing.T) {
	t.Parallel()

	txDbProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		&config.FeeSettings{},
	)

	scAddr := []byte("addr")
	addr := []byte("addr")
	identifier := []byte("id")
	top1, top2 := []byte("t1"), []byte("t2")
	dt := []byte("dt")
	txLog := &transaction.Log{
		Address: scAddr,
		Events: []*transaction.Event{
			{
				Address:    addr,
				Identifier: identifier,
				Topics:     [][]byte{top1, top2},
				Data:       dt,
			},
		},
	}
	expectedTxLog := TxLog{
		Address: txDbProc.addressPubkeyConverter.Encode(scAddr),
		Events: []Event{
			{
				Address:    hex.EncodeToString(addr),
				Identifier: hex.EncodeToString(identifier),
				Topics:     []string{hex.EncodeToString(top1), hex.EncodeToString(top2)},
				Data:       hex.EncodeToString(dt),
			},
		},
	}

	dbTxLog := txDbProc.prepareTxLog(txLog)
	assert.Equal(t, expectedTxLog, dbTxLog)
}

func TestRelayedTransactions(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
		Data:     []byte("relayedTx@blablabllablalba"),
	}

	scHash1 := []byte("scHash1")
	scResult1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}
	scHash2 := []byte("scHash2")
	scResult2 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}
	scHash3 := []byte("scHash3")
	scResult3 := &smartContractResult.SmartContractResult{
		OriginalTxHash: scHash1,
		Data:           []byte("@" + "6F6B"),
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{scHash1, scHash2, scHash3},
				Type:     block.SmartContractResultBlock,
			},
		},
	}

	header := &block.Header{}
	txPool := map[string]data.TransactionHandler{
		string(txHash1): tx1,
		string(scHash1): scResult1,
		string(scHash2): scResult2,
		string(scHash3): scResult3,
	}

	txDbProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		&config.FeeSettings{},
	)

	transactions, _ := txDbProc.prepareTransactionsForDatabase(body, header, txPool, 0)
	assert.Equal(t, 1, len(transactions))
	assert.Equal(t, 3, len(transactions[0].SmartContractResults))
	assert.Equal(t, txStatusSuccess, transactions[0].Status)
}

func TestSetTransactionSearchOrder(t *testing.T) {
	t.Parallel()
	txHash1 := []byte("txHash1")
	tx1 := &Transaction{}

	txHash2 := []byte("txHash2")
	tx2 := &Transaction{}

	txPool := map[string]*Transaction{
		string(txHash1): tx1,
		string(txHash2): tx2,
	}

	txDbProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		&config.FeeSettings{},
	)

	transactions := txDbProc.setTransactionSearchOrder(txPool)
	assert.True(t, txPoolHasSearchOrder(transactions, 0))
	assert.True(t, txPoolHasSearchOrder(transactions, 1))

	transactions = txDbProc.setTransactionSearchOrder(txPool)
	assert.True(t, txPoolHasSearchOrder(transactions, 0))
	assert.True(t, txPoolHasSearchOrder(transactions, 1))

	transactions = txDbProc.setTransactionSearchOrder(txPool)
	assert.True(t, txPoolHasSearchOrder(transactions, 0))
	assert.True(t, txPoolHasSearchOrder(transactions, 1))
}

func txPoolHasSearchOrder(txPool map[string]*Transaction, searchOrder uint32) bool {
	for _, tx := range txPool {
		if tx.SearchOrder == searchOrder {
			return true
		}
	}

	return false
}
