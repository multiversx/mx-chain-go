package txcache

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/state"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	txcache2 "github.com/multiversx/mx-chain-go/txcache"
	"github.com/stretchr/testify/require"
)

const maxNumBytesUpperBound = 1_073_741_824           // one GB
const maxNumBytesPerSenderUpperBoundTest = 33_554_432 // 32 MB
var oneQuarterOfEGLD = big.NewInt(250000000000000000)

type accountInfo struct {
	balance *big.Int
	nonce   uint64
}

func createMockSelectionSession(accounts map[string]*accountInfo) *txcachemocks.SelectionSessionMock {
	sessionMock := txcachemocks.SelectionSessionMock{
		GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
			return &stateMock.StateUserAccountHandlerStub{
				GetBalanceCalled: func() *big.Int {
					return accounts[string(address)].balance
				},
				GetNonceCalled: func() uint64 {
					return accounts[string(address)].nonce
				},
			}, nil
		},
	}

	return &sessionMock
}

func Test_Selection(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txcache, err := txcache2.NewTxCache(txcache2.ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
		TxCacheBoundsConfig: config.TxCacheBoundsConfig{
			MaxNumBytesPerSenderUpperBound: maxNumBytesPerSenderUpperBoundTest,
		},
	}, host)

	require.Nil(t, err)
	require.NotNil(t, txcache)

	selectionSession := createMockSelectionSession(map[string]*accountInfo{
		"alice": {
			balance: big.NewInt(0),
			nonce:   0,
		},
		"bob": {
			balance: big.NewInt(0),
			nonce:   0,
		},
		"carol": {
			balance: big.NewInt(0),
			nonce:   0,
		},
		"receiver": {
			balance: big.NewInt(0),
			nonce:   0,
		},
		"relayer": {
			balance: big.NewInt(0),
			nonce:   0,
		},
	})
	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		30_000,
		250,
		10,
	)

	// Consume most of relayer's balance. Keep an amount that is enough for the fee of two simple transfer transactions.
	currentRelayerBalance := int64(1000000000000000000)
	feeForTransfer := int64(50_000 * 1_000_000_004)
	feeForRelayingTransactionsOfAliceAndBob := int64(100_000*1_000_000_003 + 100_000*1_000_000_002)

	transactions := make([]*transaction.Transaction, 0)

	transactions = append(transactions, &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(currentRelayerBalance - feeForTransfer - feeForRelayingTransactionsOfAliceAndBob),
		SndAddr:   []byte("relayer"),
		RcvAddr:   []byte("receiver"),
		Data:      []byte{},
		GasLimit:  50_000,
		GasPrice:  1_000_000_004,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	// Transfer from Alice (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("alice"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_003,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Bob (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("bob"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_002,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Carol (relayed) - this one should not be selected due to insufficient balance (of the relayer)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("carol"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_001,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	txHashes := make([][]byte, 0)
	for i, tx := range transactions {
		txHash := []byte(fmt.Sprintf("txHash%d", i))
		txHashes = append(txHashes, txHash)
		txcache.AddTx(&txcache2.WrappedTransaction{
			Tx:               tx,
			TxHash:           txHash,
			SenderShardID:    0,
			ReceiverShardID:  0,
			Size:             0,
			Fee:              big.NewInt(int64(tx.GasLimit * tx.GasPrice)),
			PricePerUnit:     0,
			TransferredValue: tx.Value,
			FeePayer:         tx.RelayerAddr,
		})
	}

	blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
		{
			TxHashes: txHashes,
		},
	}}

	err = txcache.OnProposedBlock([]byte("blockHash1"), &blockBody, &block.Header{
		Nonce:    0,
		PrevHash: nil,
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	})
	require.Nil(t, err)

	require.Equal(t, txcache.CountTx(), uint64(4))
	selectedTransactions, _ := txcache.SelectTransactions(selectionSession, options)
	require.Equal(t, len(selectedTransactions), 3)
}
