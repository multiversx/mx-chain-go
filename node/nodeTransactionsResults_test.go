package node_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node"
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

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizerdMock
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterMock{}

	dataComponents := getDefaultDataComponents()
	dataComponents.Store = dataStore

	processComponents := getDefaultProcessComponents()
	processComponents.HistoryRepositoryInternal = historyRepo

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
	)

	epoch := uint32(0)

	tx := &transaction.ApiTransactionResult{}

	expectedRecAPI := &transaction.ApiReceipt{
		Value:   rec.Value,
		Data:    string(rec.Data),
		TxHash:  hex.EncodeToString(txHash),
		SndAddr: n.GetCoreComponents().AddressPubKeyConverter().Encode(rec.SndAddr),
	}

	n.PutResultsInTransaction(txHash, tx, epoch)
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
			switch unitType {
			case dataRetriever.UnsignedTransactionUnit:
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
			default:
				return mock.NewStorerMock()
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

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizerdMock
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterMock{}

	dataComponents := getDefaultDataComponents()
	dataComponents.Store = dataStore

	processComponents := getDefaultProcessComponents()
	processComponents.HistoryRepositoryInternal = historyRepo

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
	)

	addressPubKeyConverter := n.GetCoreComponents().AddressPubKeyConverter()
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
			SndAddr:        addressPubKeyConverter.Encode(scr1.SndAddr),
			RcvAddr:        addressPubKeyConverter.Encode(scr1.RcvAddr),
			RelayerAddr:    addressPubKeyConverter.Encode(scr1.RelayerAddr),
			OriginalSender: addressPubKeyConverter.Encode(scr1.OriginalSender),
			Logs:           nil,
		},
		{
			Hash:           hex.EncodeToString(scrHash2),
			OriginalTxHash: hex.EncodeToString(scr1.OriginalTxHash),
			Logs:           nil,
		},
	}

	tx := &transaction.ApiTransactionResult{}
	n.PutResultsInTransaction(txHash, tx, epoch)
	require.Equal(t, expectedSCRS, tx.SmartContractResults)
}

func TestPutLogsInTransaction(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	txHash := []byte("txHash")

	logsAndEvents := &transaction.Log{
		Address: []byte("sender"),
		Events: []*transaction.Event{
			{
				Address:    []byte("addr1"),
				Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
				Topics:     [][]byte{[]byte("topic1"), []byte("topic2")},
				Data:       []byte("data1"),
			},
			{
				Address:    []byte("addr2"),
				Identifier: []byte(core.BuiltInFunctionESDTBurn),
				Topics:     [][]byte{[]byte("topic1"), []byte("topic2")},
				Data:       []byte("data1"),
			},
		},
	}

	marshalizerMock := &mock.MarshalizerFake{}
	dataStore := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					switch {
					case bytes.Equal(key, txHash):
						return marshalizerMock.Marshal(logsAndEvents)
					default:
						return nil, nil
					}
				},
			}
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizerMock
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterMock{}

	dataComponents := getDefaultDataComponents()
	dataComponents.Store = dataStore

	historyRepo := &testscommon.HistoryRepositoryStub{
		GetEventsHashesByTxHashCalled: func(hash []byte, e uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return nil, errors.New("local err")
		},
	}
	processComponents := getDefaultProcessComponents()
	processComponents.HistoryRepositoryInternal = historyRepo

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
	)

	addressPubKeyConverter := n.GetCoreComponents().AddressPubKeyConverter()
	expectedLogs := &transaction.ApiLogs{
		Address: addressPubKeyConverter.Encode(logsAndEvents.Address),
		Events: []*transaction.Events{
			{
				Address:    addressPubKeyConverter.Encode(logsAndEvents.Events[0].Address),
				Identifier: string(logsAndEvents.Events[0].Identifier),
				Topics:     logsAndEvents.Events[0].Topics,
				Data:       logsAndEvents.Events[0].Data,
			},
			{
				Address:    addressPubKeyConverter.Encode(logsAndEvents.Events[1].Address),
				Identifier: string(logsAndEvents.Events[1].Identifier),
				Topics:     logsAndEvents.Events[1].Topics,
				Data:       logsAndEvents.Events[1].Data,
			},
		},
	}

	tx := &transaction.ApiTransactionResult{}
	n.PutResultsInTransaction(txHash, tx, epoch)
	require.Equal(t, expectedLogs, tx.Logs)
}
