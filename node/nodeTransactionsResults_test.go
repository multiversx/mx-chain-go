package node

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestPutEventsInTransactionReceipt(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	receiptHash := []byte("hash")
	rec := &receipt.Receipt{
		TxHash:  txHash,
		Data:    []byte("invalid tx"),
		Value:   big.NewInt(1000),
		SndAddr: []byte("sndAddr"),
	}

	marshalizerdMock := &mock.MarshalizerFake{}
	dataStore := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					recBytes, _ := json.Marshal(rec)
					return recBytes, nil
				},
			}
		},
	}
	historyRepo := &testscommon.HistoryRepositoryStub{
		GetEventsHashesByTxHashCalled: func(hash []byte, epoch uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return &dblookupext.ResultsHashesByTxHash{
				ReceiptsHash: receiptHash,
			}, nil
		},
	}
	n, _ := NewNode(
		WithInternalMarshalizer(marshalizerdMock, 0),
		WithDataStore(dataStore),
		WithHistoryRepository(historyRepo),
		WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
	)

	epoch := uint32(0)

	tx := &transaction.ApiTransactionResult{}

	expectedRecAPI := &transaction.ReceiptApi{
		Value:   rec.Value,
		Data:    string(rec.Data),
		TxHash:  hex.EncodeToString(txHash),
		SndAddr: n.addressPubkeyConverter.Encode(rec.SndAddr),
	}

	n.putResultsInTransaction(txHash, tx, epoch)
	require.Equal(t, expectedRecAPI, tx.Receipt)
}

func TestPutEventsInTransactionSmartContractResults(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	txHash := []byte("txHash")
	scrHash1 := []byte("scrHash1")
	scrHash2 := []byte("scrHash2")

	scr1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash,
		RelayerAddr:    []byte("relayer"),
		OriginalSender: []byte("originalSender"),
		PrevTxHash:     []byte("prevTxHash"),
		SndAddr:        []byte("sender"),
		RcvAddr:        []byte("receiver"),
		Nonce:          1,
		Value:          big.NewInt(1000),
		GasLimit:       1,
		GasPrice:       5,
		Code:           []byte("code"),
		Data:           []byte("data"),
	}
	scr2 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash,
	}

	marshalizerdMock := &mock.MarshalizerFake{}
	dataStore := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					switch {
					case bytes.Equal(key, scrHash1):
						return marshalizerdMock.Marshal(scr1)
					case bytes.Equal(key, scrHash2):
						return marshalizerdMock.Marshal(scr2)
					default:
						return nil, nil
					}
				},
			}
		},
	}
	historyRepo := &testscommon.HistoryRepositoryStub{
		GetEventsHashesByTxHashCalled: func(hash []byte, e uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return &dblookupext.ResultsHashesByTxHash{
				ReceiptsHash: nil,
				ScResultsHashesAndEpoch: []*dblookupext.ScResultsHashesAndEpoch{
					{
						Epoch:           epoch,
						ScResultsHashes: [][]byte{scrHash1, scrHash2},
					},
				},
			}, nil
		},
	}
	n, _ := NewNode(
		WithInternalMarshalizer(marshalizerdMock, 0),
		WithDataStore(dataStore),
		WithHistoryRepository(historyRepo),
		WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
	)

	expectedSCRS := []*transaction.ApiSmartContractResult{
		{
			Hash:           hex.EncodeToString(scrHash1),
			Nonce:          scr1.Nonce,
			Value:          scr1.Value,
			RelayedValue:   scr1.RelayedValue,
			Code:           string(scr1.Code),
			Data:           string(scr1.Data),
			PrevTxHash:     hex.EncodeToString(scr1.PrevTxHash),
			OriginalTxHash: hex.EncodeToString(scr1.OriginalTxHash),
			GasLimit:       scr1.GasLimit,
			GasPrice:       scr1.GasPrice,
			CallType:       scr1.CallType,
			CodeMetadata:   string(scr1.CodeMetadata),
			ReturnMessage:  string(scr1.ReturnMessage),
			SndAddr:        n.addressPubkeyConverter.Encode(scr1.SndAddr),
			RcvAddr:        n.addressPubkeyConverter.Encode(scr1.RcvAddr),
			RelayerAddr:    n.addressPubkeyConverter.Encode(scr1.RelayerAddr),
			OriginalSender: n.addressPubkeyConverter.Encode(scr1.OriginalSender),
		},
		{
			Hash:           hex.EncodeToString(scrHash2),
			OriginalTxHash: hex.EncodeToString(scr1.OriginalTxHash),
		},
	}

	tx := &transaction.ApiTransactionResult{}
	n.putResultsInTransaction(txHash, tx, epoch)
	require.Equal(t, expectedSCRS, tx.SmartContractResults)
}

func TestSetStatusIfIsESDTTransferFail(t *testing.T) {
	t.Parallel()

	n, _ := NewNode(
		WithShardCoordinator(&mock.ShardCoordinatorMock{
			SelfShardId: 0,
		}),
	)

	// ESDT transfer fail
	tx1 := &transaction.ApiTransactionResult{
		Nonce:            1,
		Hash:             "myHash",
		SourceShard:      1,
		DestinationShard: 0,
		Data:             []byte("ESDTTransfer@42524f2d343663663439@a688906bd8b00000"),
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{
				OriginalTxHash: "myHash",
				Nonce:          1,
				Data:           "ESDTTransfer@42524f2d343663663439@a688906bd8b00000@75736572206572726f72",
			},
		},
	}

	n.setStatusIfIsESDTTransferFail(tx1)
	require.Equal(t, transaction.TxStatusFail, tx1.Status)

	// transaction with no SCR should be ignored
	tx2 := &transaction.ApiTransactionResult{
		Status: transaction.TxStatusSuccess,
	}
	n.setStatusIfIsESDTTransferFail(tx2)
	require.Equal(t, transaction.TxStatusSuccess, tx2.Status)

	// intra shard transaction should be ignored
	tx3 := &transaction.ApiTransactionResult{
		Status: transaction.TxStatusSuccess,
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{},
			{},
		},
	}
	n.setStatusIfIsESDTTransferFail(tx3)
	require.Equal(t, transaction.TxStatusSuccess, tx3.Status)

	// no ESDT transfer should be ignored
	tx4 := &transaction.ApiTransactionResult{
		Status:           transaction.TxStatusSuccess,
		SourceShard:      1,
		DestinationShard: 0,
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{},
			{},
		},
	}
	n.setStatusIfIsESDTTransferFail(tx4)
	require.Equal(t, transaction.TxStatusSuccess, tx4.Status)
}
