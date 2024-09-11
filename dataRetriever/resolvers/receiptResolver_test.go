package resolvers_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/state"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func createMockArgReceiptResolver() resolvers.ArgReceiptResolver {
	return resolvers.ArgReceiptResolver{
		ArgBaseResolver: createMockArgBaseResolver(),
		ReceiptPool:     testscommon.NewCacherStub(),
		ReceiptStorage:  &storageStubs.StorerStub{},
		DataPacker:      &mock.DataPackerStub{},
	}
}

func TestNewReceiptResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil receipt pool should error", func(t *testing.T) {
		args := createMockArgReceiptResolver()
		args.ReceiptPool = nil

		receiptResolver, err := resolvers.NewReceiptResolver(args)
		require.Equal(t, dataRetriever.ErrNilReceiptsPool, err)
		require.Nil(t, receiptResolver)
	})
	t.Run("nil receipt storage should error", func(t *testing.T) {
		args := createMockArgReceiptResolver()
		args.ReceiptStorage = nil

		receiptResolver, err := resolvers.NewReceiptResolver(args)
		require.Equal(t, dataRetriever.ErrNilReceiptsStorage, err)
		require.Nil(t, receiptResolver)
	})

	t.Run("should work", func(t *testing.T) {
		args := createMockArgReceiptResolver()

		receiptResolver, err := resolvers.NewReceiptResolver(args)
		require.Nil(t, err)
		require.NotNil(t, receiptResolver)
	})
}

func TestReceiptResolver_ProcessReceivedMessageUnmarshalFails(t *testing.T) {
	t.Parallel()

	goodMarshalizer := &mock.MarshalizerMock{}
	cnt := 0
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: goodMarshalizer.Marshal,
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			cnt++
			if cnt > 1 {
				return expectedErr
			}
			return goodMarshalizer.Unmarshal(obj, buff)
		},
	}
	receiptHash := []byte("aaa")
	receiptList := make([][]byte, 0)
	receiptList = append(receiptList, receiptHash)
	requestedBuff, merr := goodMarshalizer.Marshal(&batch.Batch{Data: receiptList})
	require.Nil(t, merr)

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	arg := createMockArgReceiptResolver()
	arg.ReceiptPool = cache
	arg.ReceiptStorage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			body := state.Receipt{}
			buff, _ := goodMarshalizer.Marshal(&body)
			return buff, nil
		},
	}
	arg.Marshaller = marshalizer
	arg.DataPacker = &mock.DataPackerStub{
		PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
			require.Fail(t, "should not have been called")
			return nil, nil
		},
	}
	mbRes, _ := resolvers.NewReceiptResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
		&p2pmocks.MessengerStub{},
	)

	require.True(t, errors.Is(err, expectedErr))
	require.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	require.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestReceiptResolver_Close(t *testing.T) {
	t.Parallel()

	arg := createMockArgReceiptResolver()
	mbRes, _ := resolvers.NewReceiptResolver(arg)

	require.Nil(t, mbRes.Close())
}
