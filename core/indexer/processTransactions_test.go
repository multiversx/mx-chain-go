package indexer

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		false,
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
		false,
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
		false,
	)

	transactions, _ := txDbProc.prepareTransactionsForDatabase(body, header, txPool, 0)
	assert.Equal(t, 1, len(transactions))
	assert.Equal(t, 3, len(transactions[0].SmartContractResults))
	assert.Equal(t, transaction.TxStatusSuccess.String(), transactions[0].Status)
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
		false,
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

func TestGetGasUsedFromReceipt_RefundedGas(t *testing.T) {
	t.Parallel()

	gasPrice := uint64(1000)
	gasLimit := uint64(10000)
	recValue := big.NewInt(10000)
	txHash := []byte("tx-hash")
	rec := &receipt.Receipt{
		Value:   recValue,
		SndAddr: nil,
		Data:    []byte(processTransaction.RefundGasMessage),
		TxHash:  txHash,
	}
	tx := &Transaction{
		Hash: hex.EncodeToString(txHash),

		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	expectedGasUsed := uint64(9990)

	gasUsed := getGasUsedFromReceipt(rec, tx)
	assert.Equal(t, expectedGasUsed, gasUsed)
}

func TestGetGasUsedFromReceipt_DataError(t *testing.T) {
	t.Parallel()

	gasPrice := uint64(1000)
	gasLimit := uint64(10000)
	recValue := big.NewInt(100000)
	txHash := []byte("tx-hash")
	rec := &receipt.Receipt{
		Value:   recValue,
		SndAddr: nil,
		Data:    []byte("error"),
		TxHash:  txHash,
	}
	tx := &Transaction{
		Hash: hex.EncodeToString(txHash),

		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	expectedGasUsed := uint64(100)

	gasUsed := getGasUsedFromReceipt(rec, tx)
	assert.Equal(t, expectedGasUsed, gasUsed)
}

func TestAlteredAddresses(t *testing.T) {
	selfShardID := uint32(0)

	expectedAlteredAccounts := make(map[string]struct{})
	// addresses marked with a comment should be added to the altered addresses map

	// normal txs
	address1 := []byte("address1") // should be added
	address2 := []byte("address2")
	expectedAlteredAccounts[hex.EncodeToString(address1)] = struct{}{}
	tx1 := &transaction.Transaction{
		SndAddr: address1,
		RcvAddr: address2,
	}
	tx1Hash := []byte("tx1Hash")

	address3 := []byte("address3")
	address4 := []byte("address4") // should be added
	expectedAlteredAccounts[hex.EncodeToString(address4)] = struct{}{}
	tx2 := &transaction.Transaction{
		SndAddr: address3,
		RcvAddr: address4,
	}
	tx2Hash := []byte("tx2hash")

	txMiniBlock1 := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{tx1Hash},
		SenderShardID:   0,
		ReceiverShardID: 1,
	}
	txMiniBlock2 := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{tx2Hash},
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	// reward txs
	address5 := []byte("address5") // should be added
	expectedAlteredAccounts[hex.EncodeToString(address5)] = struct{}{}
	rwdTx1 := &rewardTx.RewardTx{
		RcvAddr: address5,
	}
	rwdTx1Hash := []byte("rwdTx1")

	address6 := []byte("address6")
	rwdTx2 := &rewardTx.RewardTx{
		RcvAddr: address6,
	}
	rwdTx2Hash := []byte("rwdTx2")

	rewTxMiniBlock1 := &block.MiniBlock{
		Type:            block.RewardsBlock,
		TxHashes:        [][]byte{rwdTx1Hash},
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 0,
	}
	rewTxMiniBlock2 := &block.MiniBlock{
		Type:            block.RewardsBlock,
		TxHashes:        [][]byte{rwdTx2Hash},
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 1,
	}

	// smart contract results
	address7 := []byte("address7") // should be added
	address8 := []byte("address8")
	expectedAlteredAccounts[hex.EncodeToString(address7)] = struct{}{}
	scr1 := &smartContractResult.SmartContractResult{
		RcvAddr: address7,
		SndAddr: address8,
	}
	scr1Hash := []byte("scr1Hash")

	address9 := []byte("address9") // should be added
	address10 := []byte("address10")
	expectedAlteredAccounts[hex.EncodeToString(address9)] = struct{}{}
	scr2 := &smartContractResult.SmartContractResult{
		RcvAddr: address9,
		SndAddr: address10,
	}
	scr2Hash := []byte("scr2Hash")

	scrMiniBlock1 := &block.MiniBlock{
		Type:            block.SmartContractResultBlock,
		TxHashes:        [][]byte{scr1Hash, scr2Hash},
		SenderShardID:   1,
		ReceiverShardID: 0,
	}
	scrMiniBlock2 := &block.MiniBlock{
		Type:            block.SmartContractResultBlock,
		TxHashes:        [][]byte{scr1Hash, scr2Hash},
		SenderShardID:   0,
		ReceiverShardID: 1,
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{txMiniBlock1, txMiniBlock2, rewTxMiniBlock1, rewTxMiniBlock2, scrMiniBlock1, scrMiniBlock2},
	}

	hdr := &block.Header{}
	txPool := map[string]data.TransactionHandler{
		string(tx1Hash):    tx1,
		string(tx2Hash):    tx2,
		string(rwdTx1Hash): rwdTx1,
		string(rwdTx2Hash): rwdTx2,
		string(scr1Hash):   scr1,
		string(scr2Hash):   scr2,
	}

	txLogProc := disabled.NewNilTxLogsProcessor()
	txProc := &txDatabaseProcessor{
		commonProcessor: &commonProcessor{
			addressPubkeyConverter: mock.NewPubkeyConverterMock(32),
		},
		marshalizer:     &mock.MarshalizerMock{},
		hasher:          &mock.HasherMock{},
		txLogsProcessor: txLogProc,
	}

	_, alteredAddresses := txProc.prepareTransactionsForDatabase(body, hdr, txPool, selfShardID)
	require.Equal(t, len(expectedAlteredAccounts), len(alteredAddresses))

	for addrActual := range alteredAddresses {
		_, found := expectedAlteredAccounts[addrActual]
		if !found {
			assert.Fail(t, fmt.Sprintf("address %s not found", addrActual))
		}
	}
}

func txPoolHasSearchOrder(txPool map[string]*Transaction, searchOrder uint32) bool {
	for _, tx := range txPool {
		if tx.SearchOrder == searchOrder {
			return true
		}
	}

	return false
}
