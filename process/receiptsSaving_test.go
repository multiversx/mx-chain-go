package process_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/stretchr/testify/require"
)

func TestSaveReceipts(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	hasher := blake2b.NewBlake2b()

	emptyReceipts := &batch.Batch{Data: make([][]byte, 0)}
	emptyReceiptsHash, _ := core.CalculateHash(marshaller, hasher, emptyReceipts)
	headerHash := []byte("header-hash")
	receiptsData := []byte("receipts-data")
	nonEmptyReceiptsHash := []byte("receipts-hash")

	transactionCoordinator := &mock.TransactionCoordinatorMock{
		CreateMarshalizedReceiptsCalled: func() ([]byte, error) {
			return receiptsData, nil
		},
	}

	t.Run("when receipts hash is for non-empty receipts", func(t *testing.T) {
		store := genericMocks.NewChainStorerMock(0)

		process.SaveReceipts(process.ArgsSaveReceipts{
			MarshalizedReceiptsProvider: transactionCoordinator,
			Marshaller:                  marshaller,
			Hasher:                      hasher,
			Store:                       store,
			Header:                      &block.Header{ReceiptsHash: nonEmptyReceiptsHash},
			HeaderHash:                  headerHash,
		})

		// Saved at key = nonEmptyReceiptsHash (header.GetReceiptsHash())
		actualData, err := store.Get(dataRetriever.ReceiptsUnit, nonEmptyReceiptsHash)
		require.Nil(t, err)
		require.Equal(t, receiptsData, actualData)

		// Not saved at key = headerHash
		_, err = store.Get(dataRetriever.ReceiptsUnit, headerHash)
		require.NotNil(t, err)
	})

	t.Run("when receipts hash is for empty receipts", func(t *testing.T) {
		store := genericMocks.NewChainStorerMock(0)

		process.SaveReceipts(process.ArgsSaveReceipts{
			MarshalizedReceiptsProvider: transactionCoordinator,
			Marshaller:                  marshaller,
			Hasher:                      hasher,
			Store:                       store,
			Header:                      &block.Header{ReceiptsHash: emptyReceiptsHash},
			HeaderHash:                  headerHash,
		})

		// Saved at key = headerHash
		actualData, err := store.Get(dataRetriever.ReceiptsUnit, headerHash)
		require.Nil(t, err)
		require.Equal(t, receiptsData, actualData)

		// Not saved at key = emptyReceiptsHash
		_, err = store.Get(dataRetriever.ReceiptsUnit, emptyReceiptsHash)
		require.NotNil(t, err)
	})
}
