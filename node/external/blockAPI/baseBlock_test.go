package blockAPI

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/stretchr/testify/require"
)

func createBaseBlockProcessor() *baseAPIBlockProcessor {
	return &baseAPIBlockProcessor{
		hasDbLookupExtensions:    true,
		selfShardID:              0,
		emptyReceiptsHash:        nil,
		store:                    &mock.ChainStorerMock{},
		marshalizer:              &mock.MarshalizerFake{},
		uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		historyRepo:              &dblookupext.HistoryRepositoryStub{},
		hasher:                   &mock.HasherFake{},
		addressPubKeyConverter:   mock.NewPubkeyConverterMock(32),
		txStatusComputer:         &mock.StatusComputerStub{},
		apiTransactionHandler:    &mock.TransactionAPIHandlerStub{},
		logsFacade:               &testscommon.LogsFacadeStub{},
		receiptsRepository:       &testscommon.ReceiptsRepositoryStub{},
	}
}

func TestBaseBlockGetIntraMiniblocksSCRS(t *testing.T) {
	t.Parallel()

	baseAPIBlockProc := createBaseBlockProcessor()

	scrHash := []byte("scr1")
	miniblock := &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{scrHash},
	}

	scResult := &smartContractResult.SmartContractResult{
		SndAddr: []byte("snd"),
		RcvAddr: []byte("rcv"),
		Data:    []byte("doSomething"),
	}
	scResultBytes, _ := baseAPIBlockProc.marshalizer.Marshal(scResult)

	baseAPIBlockProc.store = genericMocks.NewChainStorerMock(0)
	baseAPIBlockProc.store.GetStorer(dataRetriever.UnsignedTransactionUnit).Put(scrHash, scResultBytes)

	baseAPIBlockProc.receiptsRepository = &testscommon.ReceiptsRepositoryStub{
		LoadReceiptsCalled: func(header data.HeaderHandler, headerHash []byte) (*process.ReceiptsHolder, error) {
			return &process.ReceiptsHolder{Miniblocks: []*block.MiniBlock{miniblock}}, nil
		},
	}

	baseAPIBlockProc.apiTransactionHandler = &mock.TransactionAPIHandlerStub{
		UnmarshalTransactionCalled: func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
			return &transaction.ApiTransactionResult{
				Sender:   hex.EncodeToString(scResult.SndAddr),
				Receiver: hex.EncodeToString(scResult.RcvAddr),
				Data:     scResult.Data,
			}, nil
		},
	}

	blockHeader := &block.Header{ReceiptsHash: []byte("aaaa"), Epoch: 0}
	intraMbs, err := baseAPIBlockProc.getIntrashardMiniblocksFromReceiptsStorage(blockHeader, []byte{}, api.BlockQueryOptions{WithTransactions: true})
	require.Nil(t, err)
	require.Equal(t, &api.MiniBlock{
		Hash: "7630a217810d1ad3ea67e32dbff0e8f3ea6d970191f03d3c71761b3b60e57b91",
		Type: "SmartContractResultBlock",
		Transactions: []*transaction.ApiTransactionResult{
			{
				Hash:          "73637231",
				HashBytes:     []byte{0x73, 0x63, 0x72, 0x31},
				Sender:        "736e64",
				Receiver:      "726376",
				Data:          []byte("doSomething"),
				MiniBlockType: "SmartContractResultBlock",
				MiniBlockHash: "7630a217810d1ad3ea67e32dbff0e8f3ea6d970191f03d3c71761b3b60e57b91",
			},
		},
		ProcessingType:        block.Normal.String(),
		IsFromReceiptsStorage: true,
	}, intraMbs[0])
}

func TestBaseBlockGetIntraMiniblocksReceipts(t *testing.T) {
	t.Parallel()

	baseAPIBlockProc := createBaseBlockProcessor()

	receiptHash := []byte("rec1")
	miniblock := &block.MiniBlock{
		Type:     block.ReceiptBlock,
		TxHashes: [][]byte{receiptHash},
	}

	receipt := &receipt.Receipt{
		Value:   big.NewInt(1000),
		SndAddr: []byte("sndAddr"),
		Data:    []byte("refund"),
		TxHash:  []byte("hash"),
	}
	receiptBytes, _ := baseAPIBlockProc.marshalizer.Marshal(receipt)

	baseAPIBlockProc.store = genericMocks.NewChainStorerMock(0)
	baseAPIBlockProc.store.GetStorer(dataRetriever.UnsignedTransactionUnit).Put(receiptHash, receiptBytes)

	baseAPIBlockProc.receiptsRepository = &testscommon.ReceiptsRepositoryStub{
		LoadReceiptsCalled: func(header data.HeaderHandler, headerHash []byte) (*process.ReceiptsHolder, error) {
			return &process.ReceiptsHolder{Miniblocks: []*block.MiniBlock{miniblock}}, nil
		},
	}

	baseAPIBlockProc.apiTransactionHandler = &mock.TransactionAPIHandlerStub{
		UnmarshalReceiptCalled: func(receiptBytes []byte) (*transaction.ApiReceipt, error) {
			return &transaction.ApiReceipt{
				Value:   receipt.Value,
				SndAddr: baseAPIBlockProc.addressPubKeyConverter.Encode(receipt.SndAddr),
				Data:    string(receipt.Data),
				TxHash:  hex.EncodeToString(receipt.TxHash),
			}, nil
		},
	}

	blockHeader := &block.Header{ReceiptsHash: []byte("aaaa"), Epoch: 0}
	intraMbs, err := baseAPIBlockProc.getIntrashardMiniblocksFromReceiptsStorage(blockHeader, []byte{}, api.BlockQueryOptions{WithTransactions: true})
	require.Nil(t, err)
	require.Equal(t, &api.MiniBlock{
		Hash: "262b3023ca9ba61e90a60932b4db7f8b0d1dec7c2a00261cf0c5d43785f17f6f",
		Type: "ReceiptBlock",
		Receipts: []*transaction.ApiReceipt{
			{
				SndAddr: "736e6441646472",
				Data:    "refund",
				TxHash:  "68617368",
				Value:   big.NewInt(1000),
			},
		},
		ProcessingType:        block.Normal.String(),
		IsFromReceiptsStorage: true,
	}, intraMbs[0])
}

func TestBaseBlock_getAndAttachTxsToMb_MiniblockTxBlock(t *testing.T) {
	t.Parallel()

	baseAPIBlockProc := createBaseBlockProcessor()

	txHash := []byte("tx1")
	txMb := &block.MiniBlock{
		Type:     block.TxBlock,
		TxHashes: [][]byte{txHash},
	}
	txMbBytes, _ := baseAPIBlockProc.marshalizer.Marshal(txMb)

	mbStorer := genericMocks.NewStorerMock()
	mbHash := []byte("mbHash")
	_ = mbStorer.Put(mbHash, txMbBytes)

	unsignedStorer := genericMocks.NewStorerMock()
	tx := &transaction.Transaction{
		Value:   big.NewInt(1000),
		SndAddr: []byte("sndAddr"),
		RcvAddr: []byte("rcvAddr"),
		Data:    []byte("refund"),
		Nonce:   1,
	}
	txBytes, _ := baseAPIBlockProc.marshalizer.Marshal(tx)
	_ = unsignedStorer.Put(txHash, txBytes)

	baseAPIBlockProc.store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.MiniBlockUnit:
				return mbStorer
			case dataRetriever.TransactionUnit:
				return unsignedStorer
			}

			return nil
		},
	}

	baseAPIBlockProc.apiTransactionHandler = &mock.TransactionAPIHandlerStub{
		UnmarshalTransactionCalled: func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
			return &transaction.ApiTransactionResult{
				Sender:   hex.EncodeToString(tx.SndAddr),
				Receiver: hex.EncodeToString(tx.RcvAddr),
				Data:     tx.Data,
				Nonce:    tx.Nonce,
			}, nil
		},
	}

	mbhr := &block.MiniBlockHeaderReserved{
		IndexOfFirstTxProcessed: 0,
		IndexOfLastTxProcessed:  1,
	}
	marshalizer := testscommon.ProtobufMarshalizerMock{}
	mbhrBytes, _ := marshalizer.Marshal(mbhr)

	mbHeader := &block.MiniBlockHeader{
		Hash:     mbHash,
		Reserved: mbhrBytes,
	}

	apiMB := &api.MiniBlock{}
	err := baseAPIBlockProc.getAndAttachTxsToMb(mbHeader, 0, apiMB, api.BlockQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, &api.MiniBlock{
		Transactions: []*transaction.ApiTransactionResult{
			{
				Nonce:         1,
				Hash:          "747831",
				HashBytes:     []byte{0x74, 0x78, 0x31},
				Sender:        "736e6441646472",
				Receiver:      "72637641646472",
				Data:          []byte("refund"),
				MiniBlockType: "TxBlock",
				MiniBlockHash: "6d6248617368",
			},
		},
	}, apiMB)
}

func TestBaseBlock_getAndAttachTxsToMbShouldIncludeLogsAsSpecified(t *testing.T) {
	t.Parallel()

	testEpoch := uint32(7)

	marshalizer := &marshal.GogoProtoMarshalizer{}

	storageService := genericMocks.NewChainStorerMock(testEpoch)
	processor := createBaseBlockProcessor()
	processor.marshalizer = marshalizer
	processor.store = storageService

	// Setup a dummy transformer for "txBytes" -> "ApiTransactionResult" (only "Nonce" is handled)
	processor.apiTransactionHandler = &mock.TransactionAPIHandlerStub{
		UnmarshalTransactionCalled: func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
			tx := &transaction.Transaction{}
			err := marshalizer.Unmarshal(tx, txBytes)
			if err != nil {
				return nil, err
			}

			return &transaction.ApiTransactionResult{Nonce: tx.Nonce}, nil
		},
	}

	// Setup a miniblock
	miniblockHash := []byte{0xff}
	miniblock := &block.MiniBlock{
		Type:     block.TxBlock,
		TxHashes: [][]byte{{0xaa}, {0xbb}, {0xcc}},
	}
	miniblockBytes, _ := processor.marshalizer.Marshal(miniblock)
	_ = storageService.Miniblocks.Put(miniblockHash, miniblockBytes)

	// Setup some transactions
	firstTx := &transaction.Transaction{Nonce: 42}
	secondTx := &transaction.Transaction{Nonce: 43}
	thirdTx := &transaction.Transaction{Nonce: 44}

	firstTxBytes, _ := marshalizer.Marshal(firstTx)
	secondTxBytes, _ := marshalizer.Marshal(secondTx)
	thirdTxBytes, _ := marshalizer.Marshal(thirdTx)

	_ = storageService.Transactions.Put([]byte{0xaa}, firstTxBytes)
	_ = storageService.Transactions.Put([]byte{0xbb}, secondTxBytes)
	_ = storageService.Transactions.Put([]byte{0xcc}, thirdTxBytes)

	// Setup some logs for 1st and 3rd transactions (none for 2nd)
	processor.logsFacade = &testscommon.LogsFacadeStub{
		IncludeLogsInTransactionsCalled: func(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error {
			// Check the input arguments to match our scenario
			if len(txs) != 3 || len(logsKeys) != 3 {
				return nil
			}
			if !bytes.Equal(logsKeys[0], []byte{0xaa}) ||
				!bytes.Equal(logsKeys[1], []byte{0xbb}) ||
				!bytes.Equal(logsKeys[2], []byte{0xcc}) {
				return nil
			}
			if epoch != testEpoch {
				return nil
			}

			txs[0].Logs = &transaction.ApiLogs{
				Events: []*transaction.Events{
					{Identifier: "first"},
				},
			}

			txs[2].Logs = &transaction.ApiLogs{
				Events: []*transaction.Events{
					{Identifier: "third"},
				},
			}

			return nil
		},
	}

	// Now let's test the loading of transaction and logs
	miniblockHeader := &block.MiniBlockHeader{Hash: miniblockHash}
	miniblockOnApi := &api.MiniBlock{}
	err := processor.getAndAttachTxsToMb(miniblockHeader, testEpoch, miniblockOnApi, api.BlockQueryOptions{WithLogs: true})

	require.Nil(t, err)
	require.Len(t, miniblockOnApi.Transactions, 3)
	require.Equal(t, uint64(42), miniblockOnApi.Transactions[0].Nonce)
	require.Equal(t, uint64(43), miniblockOnApi.Transactions[1].Nonce)
	require.Equal(t, uint64(44), miniblockOnApi.Transactions[2].Nonce)
	require.Equal(t, "first", miniblockOnApi.Transactions[0].Logs.Events[0].Identifier)
	require.Nil(t, miniblockOnApi.Transactions[1].Logs)
	require.Equal(t, "third", miniblockOnApi.Transactions[2].Logs.Events[0].Identifier)
}

func TestExtractExecutedTxHashes(t *testing.T) {
	t.Parallel()

	array := make([][]byte, 10)
	res := extractExecutedTxHashes(array, 0, int32(len(array))-1)
	require.Len(t, res, 10)

	res = extractExecutedTxHashes(array, 0, int32(len(array)))
	require.Equal(t, res, array)

	res = extractExecutedTxHashes(array, -1, int32(len(array)))
	require.Equal(t, res, array)

	res = extractExecutedTxHashes(array, 20, int32(len(array)))
	require.Equal(t, res, array)

	res = extractExecutedTxHashes(array, 0, int32(len(array))+1)
	require.Equal(t, res, array)

	array = make([][]byte, 0, 10)
	for idx := 0; idx < 10; idx++ {
		array = append(array, []byte{byte(idx)})
	}
	res = extractExecutedTxHashes(array, 0, 5)
	require.Equal(t, res, [][]byte{{byte(0)}, {byte(1)}, {byte(2)}, {byte(3)}, {byte(4)}, {byte(5)}})
}
