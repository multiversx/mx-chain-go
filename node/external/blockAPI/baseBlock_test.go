package blockAPI

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/stretchr/testify/require"
)

func createMockBaseBlock() *baseAPIBlockProcessor {
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
		txUnmarshaller:           &mock.TransactionAPIHandlerStub{},
	}
}

func TestBaseBlockGetIntraMiniblocksSCRS(t *testing.T) {
	t.Parallel()

	baseAPIBlockProc := createMockBaseBlock()

	scrHash := []byte("scr1")
	mbScrs := &block.MiniBlock{
		Type:     block.SmartContractResultBlock,
		TxHashes: [][]byte{scrHash},
	}
	mbScrsBytes, _ := baseAPIBlockProc.marshalizer.Marshal(mbScrs)

	receiptsStorer := mock.NewStorerMock()

	batchData := &batch.Batch{
		Data: [][]byte{mbScrsBytes},
	}
	batchDataBytes, _ := baseAPIBlockProc.marshalizer.Marshal(batchData)

	receiptsHash := []byte("recHash")
	_ = receiptsStorer.Put(receiptsHash, batchDataBytes)

	unsignedStorer := mock.NewStorerMock()
	scResult := &smartContractResult.SmartContractResult{
		SndAddr: []byte("snd"),
		RcvAddr: []byte("rcv"),
		Data:    []byte("doSomething"),
	}
	scResultBytes, _ := baseAPIBlockProc.marshalizer.Marshal(scResult)
	_ = unsignedStorer.Put(scrHash, scResultBytes)

	baseAPIBlockProc.store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.ReceiptsUnit:
				return receiptsStorer
			case dataRetriever.UnsignedTransactionUnit:
				return unsignedStorer
			}

			return nil
		},
	}

	baseAPIBlockProc.txUnmarshaller = &mock.TransactionAPIHandlerStub{
		UnmarshalTransactionCalled: func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
			return &transaction.ApiTransactionResult{
				Sender:   hex.EncodeToString(scResult.SndAddr),
				Receiver: hex.EncodeToString(scResult.RcvAddr),
				Data:     scResult.Data,
			}, nil
		},
	}

	intraMbs := baseAPIBlockProc.getIntraMiniblocks(receiptsHash, 0, true)
	require.Equal(t, &api.MiniBlock{
		Hash: "7630a217810d1ad3ea67e32dbff0e8f3ea6d970191f03d3c71761b3b60e57b91",
		Type: "SmartContractResultBlock",
		Transactions: []*transaction.ApiTransactionResult{
			{
				Hash:          "73637231",
				Sender:        "736e64",
				Receiver:      "726376",
				Data:          []byte("doSomething"),
				MiniBlockType: "SmartContractResultBlock",
				MiniBlockHash: "7630a217810d1ad3ea67e32dbff0e8f3ea6d970191f03d3c71761b3b60e57b91",
			},
		},
	}, intraMbs[0])
}

func TestBaseBlockGetIntraMiniblocksReceipts(t *testing.T) {
	t.Parallel()

	baseAPIBlockProc := createMockBaseBlock()

	recHash := []byte("rec1")
	recMb := &block.MiniBlock{
		Type:     block.ReceiptBlock,
		TxHashes: [][]byte{recHash},
	}
	recMbBytes, _ := baseAPIBlockProc.marshalizer.Marshal(recMb)

	receiptsStorer := mock.NewStorerMock()

	batchData := &batch.Batch{
		Data: [][]byte{recMbBytes},
	}
	batchDataBytes, _ := baseAPIBlockProc.marshalizer.Marshal(batchData)

	receiptsHash := []byte("recHash")
	_ = receiptsStorer.Put(receiptsHash, batchDataBytes)

	unsignedStorer := mock.NewStorerMock()
	rec := &receipt.Receipt{
		Value:   big.NewInt(1000),
		SndAddr: []byte("sndAddr"),
		Data:    []byte("refund"),
		TxHash:  []byte("hash"),
	}
	recBytes, _ := baseAPIBlockProc.marshalizer.Marshal(rec)
	_ = unsignedStorer.Put(recHash, recBytes)

	baseAPIBlockProc.txUnmarshaller = &mock.TransactionAPIHandlerStub{
		UnmarshalReceiptCalled: func(receiptBytes []byte) (*transaction.ApiReceipt, error) {
			return &transaction.ApiReceipt{
				Value:   rec.Value,
				SndAddr: baseAPIBlockProc.addressPubKeyConverter.Encode(rec.SndAddr),
				Data:    string(rec.Data),
				TxHash:  hex.EncodeToString(rec.TxHash),
			}, nil
		},
	}

	baseAPIBlockProc.store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.ReceiptsUnit:
				return receiptsStorer
			case dataRetriever.UnsignedTransactionUnit:
				return unsignedStorer
			}

			return nil
		},
	}

	intraMbs := baseAPIBlockProc.getIntraMiniblocks(receiptsHash, 0, true)
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
	}, intraMbs[0])
}

func TestBaseBlock_getAndAttachTxsToMb_MiniblockTxBlock(t *testing.T) {
	t.Parallel()

	baseAPIBlockProc := createMockBaseBlock()

	txHash := []byte("tx1")
	txMb := &block.MiniBlock{
		Type:     block.TxBlock,
		TxHashes: [][]byte{txHash},
	}
	txMbBytes, _ := baseAPIBlockProc.marshalizer.Marshal(txMb)

	mbStorer := mock.NewStorerMock()
	mbHash := []byte("mbHash")
	_ = mbStorer.Put(mbHash, txMbBytes)

	unsignedStorer := mock.NewStorerMock()
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

	baseAPIBlockProc.txUnmarshaller = &mock.TransactionAPIHandlerStub{
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
	baseAPIBlockProc.getAndAttachTxsToMb(mbHeader, 0, apiMB)
	require.Equal(t, &api.MiniBlock{
		Transactions: []*transaction.ApiTransactionResult{
			{
				Nonce:         1,
				Hash:          "747831",
				Sender:        "736e6441646472",
				Receiver:      "72637641646472",
				Data:          []byte("refund"),
				MiniBlockType: "TxBlock",
				MiniBlockHash: "6d6248617368",
			},
		},
	}, apiMB)
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
