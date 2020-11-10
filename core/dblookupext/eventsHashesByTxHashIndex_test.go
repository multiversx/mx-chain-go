package dblookupext

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericmocks"
	"github.com/stretchr/testify/require"
)

func TestGetResultsHashesByTxHashShouldErr(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	marshalizerMock := &mock.MarshalizerMock{}
	storerMock := genericmocks.NewStorerMock("EventsHashesByTxHash", epoch)

	eventsHashesIndex := newEventsHashesByTxHash(storerMock, marshalizerMock)

	eventsHashes, err := eventsHashesIndex.getEventsHashesByTxHash([]byte("hash"), 0)
	require.Nil(t, eventsHashes)
	require.Error(t, err)
}

func TestSaveAndGetResultsSCRSHashesByTxHash(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	marshalizerMock := &mock.MarshalizerMock{}
	storerMock := genericmocks.NewStorerMock("EventsHashesByTxHash", epoch)

	eventsHashesIndex := newEventsHashesByTxHash(storerMock, marshalizerMock)

	originalTxHash := []byte("txHash")
	scrHash1 := []byte("scrHash1")
	scrHash2 := []byte("scrHash2")
	scrResults1 := map[string]data.TransactionHandler{
		string(scrHash1): &smartContractResult.SmartContractResult{
			OriginalTxHash: originalTxHash,
		},
		string(scrHash2): &smartContractResult.SmartContractResult{
			OriginalTxHash: originalTxHash,
		},
		"wrongTx": &transaction.Transaction{},
	}
	err := eventsHashesIndex.saveResultsHashes(epoch, scrResults1, nil)
	require.Nil(t, err)

	scrHash3 := []byte("scrHash3")
	scrHash4 := []byte("scrHash4")
	scrResults2 := map[string]data.TransactionHandler{
		string(scrHash3): &smartContractResult.SmartContractResult{
			OriginalTxHash: originalTxHash,
		},
		string(scrHash4): &smartContractResult.SmartContractResult{
			OriginalTxHash: originalTxHash,
		},
	}
	err = eventsHashesIndex.saveResultsHashes(epoch, scrResults2, nil)
	require.Nil(t, err)

	expectedEvents := &ResultsHashesByTxHash{
		ReceiptsHash: nil,
		ScResultsHashesAndEpoch: []*ScResultsHashesAndEpoch{
			{
				Epoch:           epoch,
				ScResultsHashes: [][]byte{scrHash1, scrHash2},
			},
			{
				Epoch:           epoch,
				ScResultsHashes: [][]byte{scrHash3, scrHash4},
			},
		},
	}

	eventsHashes, err := eventsHashesIndex.getEventsHashesByTxHash(originalTxHash, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedEvents, eventsHashes)
}

func TestSaveAndGetResultsReceiptsHashesByTxHash(t *testing.T) {
	epoch := uint32(0)
	marshalizerMock := &mock.MarshalizerMock{}
	storerMock := genericmocks.NewStorerMock("EventsHashesByTxHash", epoch)

	eventsHashesIndex := newEventsHashesByTxHash(storerMock, marshalizerMock)

	txWithReceiptHash := []byte("invalidTxHash")
	recHash1 := []byte("receiptHash")
	receipts := map[string]data.TransactionHandler{
		string(recHash1): &receipt.Receipt{
			TxHash: txWithReceiptHash,
		},
		"wrongTx": &transaction.Transaction{},
	}

	err := eventsHashesIndex.saveResultsHashes(epoch, nil, receipts)
	require.Nil(t, err)

	expectedEvents := &ResultsHashesByTxHash{
		ReceiptsHash:            recHash1,
		ScResultsHashesAndEpoch: nil,
	}

	eventsHashes, err := eventsHashesIndex.getEventsHashesByTxHash(txWithReceiptHash, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedEvents, eventsHashes)
}

func TestGroupSmartContractResults(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	marshalizerMock := &mock.MarshalizerMock{}
	storerMock := genericmocks.NewStorerMock("EventsHashesByTxHash", epoch)

	eventsHashesIndex := newEventsHashesByTxHash(storerMock, marshalizerMock)

	originalTxHash := []byte("txHash")
	scrHash1 := []byte("scrHash1")
	scrHash2 := []byte("scrHash2")
	scrResults1 := map[string]data.TransactionHandler{
		string(scrHash1): &smartContractResult.SmartContractResult{
			OriginalTxHash: originalTxHash,
		},
		string(scrHash2): &smartContractResult.SmartContractResult{
			OriginalTxHash: originalTxHash,
		},
		"wrongTx": &transaction.Transaction{},
	}
	err := eventsHashesIndex.saveResultsHashes(epoch, scrResults1, nil)
	require.Nil(t, err)

	expectedResultHashes1 := &ResultsHashesByTxHash{
		ScResultsHashesAndEpoch: []*ScResultsHashesAndEpoch{
			{
				Epoch:           epoch,
				ScResultsHashes: [][]byte{scrHash1, scrHash2},
			},
		},
	}
	eventsHashes, err := eventsHashesIndex.getEventsHashesByTxHash(originalTxHash, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedResultHashes1, eventsHashes)

	scrHash3 := []byte("scrHash3")
	scrHash4 := []byte("scrHash4")
	scrResults2 := map[string]data.TransactionHandler{
		string(scrHash3): &smartContractResult.SmartContractResult{
			OriginalTxHash: originalTxHash,
		},
		string(scrHash4): &smartContractResult.SmartContractResult{
			OriginalTxHash: originalTxHash,
		},
	}
	err = eventsHashesIndex.saveResultsHashes(epoch, scrResults2, nil)
	require.Nil(t, err)

	expectedResultHashes2 := &ResultsHashesByTxHash{
		ScResultsHashesAndEpoch: []*ScResultsHashesAndEpoch{
			{
				Epoch:           epoch,
				ScResultsHashes: [][]byte{scrHash1, scrHash2},
			},
			{
				Epoch:           epoch,
				ScResultsHashes: [][]byte{scrHash3, scrHash4},
			},
		},
	}
	eventsHashes, err = eventsHashesIndex.getEventsHashesByTxHash(originalTxHash, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedResultHashes2, eventsHashes)
}
