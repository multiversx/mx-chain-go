package transactionAPI

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dbLookupExtMock "github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
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
			return &storageStubs.StorerStub{
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					recBytes, _ := json.Marshal(rec)
					return recBytes, nil
				},
			}
		},
	}
	historyRepo := &dbLookupExtMock.HistoryRepositoryStub{
		GetEventsHashesByTxHashCalled: func(hash []byte, epoch uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return &dblookupext.ResultsHashesByTxHash{
				ReceiptsHash: receiptHash,
			}, nil
		},
	}

	pubKeyConverter := &mock.PubkeyConverterMock{}
	txUnmarshalerAndPreparer := newTransactionUnmarshaller(marshalizerdMock, pubKeyConverter)
	logsFacade := &testscommon.LogsFacadeStub{}

	n := newAPITransactionResultProcessor(pubKeyConverter, historyRepo, dataStore, marshalizerdMock, txUnmarshalerAndPreparer, logsFacade, 0)

	epoch := uint32(0)

	tx := &transaction.ApiTransactionResult{}

	expectedRecAPI := &transaction.ApiReceipt{
		Value:   rec.Value,
		Data:    string(rec.Data),
		TxHash:  hex.EncodeToString(txHash),
		SndAddr: pubKeyConverter.Encode(rec.SndAddr),
	}

	n.putResultsInTransaction(txHash, tx, epoch)
	require.Equal(t, expectedRecAPI, tx.Receipt)
}

func TestPutEventsInTransactionSmartContractResults(t *testing.T) {
	t.Parallel()

	testEpoch := uint32(0)
	testTxHash := []byte("txHash")
	scrHash1 := []byte("scrHash1")
	scrHash2 := []byte("scrHash2")

	scr1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: testTxHash,
		RelayerAddr:    []byte("rlr"),
		OriginalSender: []byte("osn"),
		PrevTxHash:     []byte("prevTxHash"),
		SndAddr:        []byte("snd"),
		RcvAddr:        []byte("rcv"),
		Nonce:          1,
		Value:          big.NewInt(1000),
		GasLimit:       1,
		GasPrice:       5,
		Code:           []byte("code"),
		Data:           []byte("data"),
	}
	scr2 := &smartContractResult.SmartContractResult{
		OriginalTxHash: testTxHash,
	}

	logs := &transaction.ApiLogs{
		Address: "erd1contract",
		Events: []*transaction.Events{
			{
				Address:    "erd1alice",
				Identifier: "first",
				Topics:     [][]byte{[]byte("hello")},
				Data:       []byte("data1"),
			},
			{
				Address:    "erd1bob",
				Identifier: "second",
				Topics:     [][]byte{[]byte("world")},
				Data:       []byte("data2"),
			},
		},
	}

	marshalizerdMock := &mock.MarshalizerFake{}
	dataStore := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.UnsignedTransactionUnit:
				return &storageStubs.StorerStub{
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
			default:
				return genericMocks.NewStorerMock()
			}
		},
	}

	historyRepo := &dbLookupExtMock.HistoryRepositoryStub{
		GetEventsHashesByTxHashCalled: func(hash []byte, e uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return &dblookupext.ResultsHashesByTxHash{
				ReceiptsHash: nil,
				ScResultsHashesAndEpoch: []*dblookupext.ScResultsHashesAndEpoch{
					{
						Epoch:           testEpoch,
						ScResultsHashes: [][]byte{scrHash1, scrHash2},
					},
				},
			}, nil
		},
	}

	pubKeyConverter := mock.NewPubkeyConverterMock(3)
	txUnmarshalerAndPreparer := newTransactionUnmarshaller(marshalizerdMock, pubKeyConverter)

	logsFacade := &testscommon.LogsFacadeStub{
		GetLogCalled: func(txHash []byte, epoch uint32) (*transaction.ApiLogs, error) {
			if bytes.Equal(txHash, scrHash1) && epoch == testEpoch {
				return logs, nil
			}

			return nil, nil
		},
	}

	n := newAPITransactionResultProcessor(pubKeyConverter, historyRepo, dataStore, marshalizerdMock, txUnmarshalerAndPreparer, logsFacade, 0)

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
			SndAddr:        pubKeyConverter.Encode(scr1.SndAddr),
			RcvAddr:        pubKeyConverter.Encode(scr1.RcvAddr),
			RelayerAddr:    pubKeyConverter.Encode(scr1.RelayerAddr),
			OriginalSender: pubKeyConverter.Encode(scr1.OriginalSender),
			Logs:           logs,
		},
		{
			Hash:           hex.EncodeToString(scrHash2),
			OriginalTxHash: hex.EncodeToString(scr1.OriginalTxHash),
			Logs:           nil,
		},
	}

	tx := &transaction.ApiTransactionResult{}
	n.putResultsInTransaction(testTxHash, tx, testEpoch)
	require.Equal(t, expectedSCRS, tx.SmartContractResults)
}

func TestPutLogsInTransaction(t *testing.T) {
	t.Parallel()

	testEpoch := uint32(7)
	testTxHash := []byte("txHash")

	logs := &transaction.ApiLogs{
		Address: "erd1contract",
		Events: []*transaction.Events{
			{
				Address:    "erd1alice",
				Identifier: "first",
				Topics:     [][]byte{[]byte("hello")},
				Data:       []byte("data1"),
			},
			{
				Address:    "erd1bob",
				Identifier: "second",
				Topics:     [][]byte{[]byte("world")},
				Data:       []byte("data2"),
			},
		},
	}

	marshalizerMock := &mock.MarshalizerFake{}
	dataStore := &mock.ChainStorerMock{}

	historyRepo := &dbLookupExtMock.HistoryRepositoryStub{
		GetEventsHashesByTxHashCalled: func(hash []byte, e uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return nil, errors.New("local err")
		},
	}

	pubKeyConverter := &mock.PubkeyConverterMock{}
	txUnmarshalerAndPreparer := newTransactionUnmarshaller(marshalizerMock, pubKeyConverter)
	logsFacade := &testscommon.LogsFacadeStub{
		GetLogCalled: func(txHash []byte, epoch uint32) (*transaction.ApiLogs, error) {
			if bytes.Equal(txHash, testTxHash) && epoch == testEpoch {
				return logs, nil
			}

			return nil, nil
		},
	}

	n := newAPITransactionResultProcessor(pubKeyConverter, historyRepo, dataStore, marshalizerMock, txUnmarshalerAndPreparer, logsFacade, 0)
	tx := &transaction.ApiTransactionResult{}
	n.putResultsInTransaction(testTxHash, tx, testEpoch)
	require.Equal(t, logs, tx.Logs)
}
