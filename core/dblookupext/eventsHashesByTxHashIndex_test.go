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

func TestGetEventsHashesByTxHashShouldErr(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	marshalizerMock := &mock.MarshalizerMock{}
	storerMock := genericmocks.NewStorerMock("EventsHashesByTxHash", epoch)

	eventsHashesIndex := newEventsHashesByTxHash(storerMock, marshalizerMock)

	eventsHashes, err := eventsHashesIndex.getEventsHashesByTxHash([]byte("hash"), 0)
	require.Nil(t, eventsHashes)
	require.Error(t, err)
}

func TestSaveAndGetEventsSCRSHashesByTxHash(t *testing.T) {
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
	err := eventsHashesIndex.saveEventsHashes(epoch, scrResults1, nil)
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
	err = eventsHashesIndex.saveEventsHashes(epoch, scrResults2, nil)
	require.Nil(t, err)

	expectedEvents := &EventsHashesByTxHash{
		ReceiptsHash: nil,
		ScrHashesEpoch: []*ScrHashesAndEpoch{
			{
				Epoch:                      epoch,
				SmartContractResultsHashes: [][]byte{scrHash3, scrHash4},
			},
			{
				Epoch:                      epoch,
				SmartContractResultsHashes: [][]byte{scrHash1, scrHash2},
			},
		},
	}

	eventsHashes, err := eventsHashesIndex.getEventsHashesByTxHash(originalTxHash, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedEvents, eventsHashes)
}

func TestSaveAndGetEventsReceiptsHashesByTxHash(t *testing.T) {
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

	err := eventsHashesIndex.saveEventsHashes(epoch, nil, receipts)
	require.Nil(t, err)

	expectedEvents := &EventsHashesByTxHash{
		ReceiptsHash:   recHash1,
		ScrHashesEpoch: nil,
	}

	eventsHashes, err := eventsHashesIndex.getEventsHashesByTxHash(txWithReceiptHash, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedEvents, eventsHashes)
}
