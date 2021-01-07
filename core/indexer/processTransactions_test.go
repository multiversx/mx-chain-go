package indexer

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	"github.com/ElrondNetwork/elrond-go/core/indexer/fees"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDummyEconomicsConfig(gasPriceModifier float64) *config.EconomicsConfig {
	return &config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "2000000000000000000000",
			MinimumInflation:   0,
			YearSettings: []*config.YearSetting{
				{
					Year:             0,
					MaximumInflation: 0.01,
				},
			},
		},
		RewardsSettings: config.RewardsSettings{
			LeaderPercentage:                 0.1,
			ProtocolSustainabilityPercentage: 0.1,
			ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			TopUpGradientPoint:               "300000000000000000000",
			TopUpFactor:                      0.25,
		},
		FeeSettings: config.FeeSettings{
			MaxGasLimitPerBlock:     "100000",
			MaxGasLimitPerMetaBlock: "1000000",
			MinGasPrice:             "1000000000",
			MinGasLimit:             "50000",
			GasPerDataByte:          "1500",
			GasPriceModifier:        gasPriceModifier,
		},
	}
}

func createArgsForEconomicsData(gasPriceModifier float64) economics.ArgsNewEconomicsData {
	args := economics.ArgsNewEconomicsData{
		Economics:                      createDummyEconomicsConfig(gasPriceModifier),
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &mock.EpochNotifierStub{},
	}
	return args
}

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

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	txDbProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		economicData,
		false,
		&mock.ShardCoordinatorMock{},
	)

	transactions, _ := txDbProc.prepareTransactionsForDatabase(body, header, txPool, 0)
	assert.Equal(t, 7, len(transactions))

}

func TestPrepareTxLog(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	txDbProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		economicData,
		false,
		&mock.ShardCoordinatorMock{},
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

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	txDbProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		economicData,
		false,
		&mock.ShardCoordinatorMock{},
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

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	txDbProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		economicData,
		false,
		&mock.ShardCoordinatorMock{},
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

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{txMiniBlock1, txMiniBlock2, rewTxMiniBlock1, rewTxMiniBlock2, scrMiniBlock1},
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

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	txLogProc := disabled.NewNilTxLogsProcessor()
	txProc := &txDatabaseProcessor{
		commonProcessor: &commonProcessor{
			addressPubkeyConverter: mock.NewPubkeyConverterMock(32),
			feesProcessor:          fees.NewFeesProcessor(economicData),
		},
		marshalizer:     &mock.MarshalizerMock{},
		hasher:          &mock.HasherMock{},
		txLogsProcessor: txLogProc,
		shardCoordinator: &mock.ShardCoordinatorMock{
			ComputeIdCalled: func(address []byte) uint32 {
				switch string(address) {
				case string(address1), string(address4), string(address5), string(address7), string(address9):
					return 0
				default:
					return 1
				}
			},
		},
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

func TestIsSCRForSenderWithGasUsed(t *testing.T) {
	t.Parallel()

	txHash := "txHash"
	nonce := uint64(10)
	sender := "sender"

	tx := &Transaction{
		Hash:   txHash,
		Nonce:  nonce,
		Sender: sender,
	}
	sc := ScResult{
		Data:      []byte("@6f6b@something"),
		Nonce:     nonce + 1,
		Receiver:  sender,
		PreTxHash: txHash,
	}

	require.True(t, isSCRForSenderWithRefund(sc, tx))
}

func TestComputeTxGasUsedField(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		GasLimit: 100000,
		GasPrice: 1000000000,
		Data:     []byte("increment"),
	}
	sc := ScResult{
		Value: "30000000000",
	}

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	feeProcessor := fees.NewFeesProcessor(economicData)

	gasUsed, fee := feeProcessor.ComputeGasUsedAndFeeBasedOnRefundValue(tx, sc.Value)
	require.Equal(t, uint64(99970), gasUsed)
	require.Equal(t, big.NewInt(99970000000000), fee)
}

func TestCheckGasUsedTooMuchGasProvidedCase(t *testing.T) {
	t.Parallel()

	tx := &Transaction{
		Hash:     "dad46ed504695598d1e95781e77f224bf4fd829a63f2b05170f1b7f4e5bcd329",
		Nonce:    11,
		GasLimit: 40000000,
		GasPrice: 1000000000,
		Status:   "success",
		Sender:   "erd1xa7lf3kux3ujrzc6ul0t7tq2x0zdph99pcg0zdj8jctftsk7krus3luq53",
	}
	sc := ScResult{
		PreTxHash:      "dad46ed504695598d1e95781e77f224bf4fd829a63f2b05170f1b7f4e5bcd329",
		Nonce:          12,
		Value:          "0",
		Receiver:       "erd1xa7lf3kux3ujrzc6ul0t7tq2x0zdph99pcg0zdj8jctftsk7krus3luq53",
		Data:           []byte("@6f6b@"),
		OriginalTxHash: "dad46ed504695598d1e95781e77f224bf4fd829a63f2b05170f1b7f4e5bcd329",
		ReturnMessage:  "too much gas provided: gas needed = 109013, gas remained = 18609087",
	}

	require.True(t, isSCRForSenderWithRefund(sc, tx))

	// fee multiplier is 1
	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	feeProcessor := fees.NewFeesProcessor(economicData)

	gasUsed, fee := feeProcessor.ComputeGasUsedAndFeeBasedOnRefundValue(tx, sc.Value)
	require.Equal(t, uint64(40000000), gasUsed)
	require.Equal(t, big.NewInt(40000000000000000), fee)
}

func TestCheckGasUsedInvalidTransaction(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	txProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		economicData,
		false,
		&mock.ShardCoordinatorMock{},
	)

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}
	recHash1 := []byte("recHash1")
	rec1 := &receipt.Receipt{
		Value:  big.NewInt(100),
		TxHash: txHash1,
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1},
				Type:     block.InvalidBlock,
			},
			{
				TxHashes: [][]byte{recHash1},
				Type:     block.ReceiptBlock,
			},
		},
	}

	header := &block.Header{}

	txPool := map[string]data.TransactionHandler{
		string(txHash1):  tx1,
		string(recHash1): rec1,
	}

	txs, _ := txProc.prepareTransactionsForDatabase(body, header, txPool, 0)
	require.Len(t, txs, 1)
	require.Equal(t, tx1.GasLimit, txs[0].GasUsed)
}

func TestCheckGasUsedRelayedTransaction(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	txProc := newTxDatabaseProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PubkeyConverterMock{},
		&mock.PubkeyConverterMock{},
		economicData,
		false,
		&mock.ShardCoordinatorMock{},
	)

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 123456,
		Data:     []byte("relayedTx@1231231231239129312"),
	}
	scResHash1 := []byte("scResHash1")
	scRes1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{scResHash1},
				Type:     block.SmartContractResultBlock,
			},
		},
	}

	header := &block.Header{}

	txPool := map[string]data.TransactionHandler{
		string(txHash1):    tx1,
		string(scResHash1): scRes1,
	}

	txs, _ := txProc.prepareTransactionsForDatabase(body, header, txPool, 0)
	require.Len(t, txs, 1)
	require.Equal(t, tx1.GasLimit, txs[0].GasUsed)
}
