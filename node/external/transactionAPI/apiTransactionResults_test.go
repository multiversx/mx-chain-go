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
	datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"
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
	dataStore := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					recBytes, _ := json.Marshal(rec)
					return recBytes, nil
				},
			}, nil
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
	logsFacade := &testscommon.LogsFacadeStub{}
	dataFieldParser := &testscommon.DataFieldParserStub{
		ParseCalled: func(dataField []byte, sender, receiver []byte, _ uint32) *datafield.ResponseParseData {
			return &datafield.ResponseParseData{}
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	txUnmarshalerAndPreparer := newTransactionUnmarshaller(marshalizerdMock, pubKeyConverter, dataFieldParser, shardCoordinator)
	n := newAPITransactionResultProcessor(pubKeyConverter, historyRepo, dataStore, marshalizerdMock, txUnmarshalerAndPreparer, logsFacade, shardCoordinator, dataFieldParser)

	epoch := uint32(0)

	tx := &transaction.ApiTransactionResult{}

	expectedRecAPI := &transaction.ApiReceipt{
		Value:   rec.Value,
		Data:    string(rec.Data),
		TxHash:  hex.EncodeToString(txHash),
		SndAddr: pubKeyConverter.Encode(rec.SndAddr),
	}

	err := n.putResultsInTransaction(txHash, tx, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedRecAPI, tx.Receipt)
}

func TestApiTransactionProcessor_PutResultsInTransactionWhenNoResultsShouldWork(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	historyRepo := &dbLookupExtMock.HistoryRepositoryStub{
		GetEventsHashesByTxHashCalled: func(hash []byte, epoch uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return nil, dblookupext.ErrNotFoundInStorage
		},
	}

	dataFieldParser := &testscommon.DataFieldParserStub{
		ParseCalled: func(dataField []byte, sender, receiver []byte, _ uint32) *datafield.ResponseParseData {
			return &datafield.ResponseParseData{}
		},
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	n := newAPITransactionResultProcessor(
		testscommon.RealWorldBech32PubkeyConverter,
		historyRepo,
		genericMocks.NewChainStorerMock(epoch),
		&testscommon.MarshalizerMock{},
		newTransactionUnmarshaller(&testscommon.MarshalizerMock{}, testscommon.RealWorldBech32PubkeyConverter, dataFieldParser, shardCoordinator),
		&testscommon.LogsFacadeStub{},
		shardCoordinator,
		dataFieldParser,
	)

	tx := &transaction.ApiTransactionResult{}
	err := n.putResultsInTransaction([]byte("txHash"), tx, epoch)
	require.Nil(t, err)
	require.Empty(t, tx.SmartContractResults)
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
	dataStore := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
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
				}, nil
			default:
				return genericMocks.NewStorerMock(), nil
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

	logsFacade := &testscommon.LogsFacadeStub{
		GetLogCalled: func(txHash []byte, epoch uint32) (*transaction.ApiLogs, error) {
			if bytes.Equal(txHash, scrHash1) && epoch == testEpoch {
				return logs, nil
			}

			return nil, nil
		},
	}

	dataFieldParser := &testscommon.DataFieldParserStub{
		ParseCalled: func(dataField []byte, sender, receiver []byte, _ uint32) *datafield.ResponseParseData {
			return &datafield.ResponseParseData{}
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	pubKeyConverter := mock.NewPubkeyConverterMock(3)
	txUnmarshalerAndPreparer := newTransactionUnmarshaller(marshalizerdMock, pubKeyConverter, dataFieldParser, shardCoordinator)
	n := newAPITransactionResultProcessor(pubKeyConverter, historyRepo, dataStore, marshalizerdMock, txUnmarshalerAndPreparer, logsFacade, shardCoordinator, dataFieldParser)

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
			Receivers:      []string{},
		},
		{
			Hash:           hex.EncodeToString(scrHash2),
			OriginalTxHash: hex.EncodeToString(scr1.OriginalTxHash),
			Logs:           nil,
			Receivers:      []string{},
		},
	}

	tx := &transaction.ApiTransactionResult{}
	err := n.putResultsInTransaction(testTxHash, tx, testEpoch)
	require.Nil(t, err)
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
	dataStore := &storageStubs.ChainStorerStub{}

	historyRepo := &dbLookupExtMock.HistoryRepositoryStub{
		GetEventsHashesByTxHashCalled: func(hash []byte, e uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return nil, errors.New("local err")
		},
	}

	logsFacade := &testscommon.LogsFacadeStub{
		GetLogCalled: func(txHash []byte, epoch uint32) (*transaction.ApiLogs, error) {
			if bytes.Equal(txHash, testTxHash) && epoch == testEpoch {
				return logs, nil
			}

			return nil, nil
		},
	}

	dataFieldParser := &testscommon.DataFieldParserStub{
		ParseCalled: func(dataField []byte, sender, receiver []byte, _ uint32) *datafield.ResponseParseData {
			return &datafield.ResponseParseData{}
		},
	}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	pubKeyConverter := &mock.PubkeyConverterMock{}
	txUnmarshalerAndPreparer := newTransactionUnmarshaller(marshalizerMock, pubKeyConverter, dataFieldParser, shardCoordinator)
	n := newAPITransactionResultProcessor(pubKeyConverter, historyRepo, dataStore, marshalizerMock, txUnmarshalerAndPreparer, logsFacade, shardCoordinator, dataFieldParser)

	tx := &transaction.ApiTransactionResult{}
	err := n.putResultsInTransaction(testTxHash, tx, testEpoch)
	// TODO: Note that "putResultsInTransaction" produces an effect on "tx" even if it returns an error.
	// TODO: Refactor this package to use less functions with side-effects.
	require.Errorf(t, err, "local err")
	require.Equal(t, logs, tx.Logs)
}
