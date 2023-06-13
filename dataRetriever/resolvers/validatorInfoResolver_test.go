package resolvers_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgValidatorInfoResolver() resolvers.ArgValidatorInfoResolver {
	return resolvers.ArgValidatorInfoResolver{
		SenderResolver:       &mock.TopicResolverSenderStub{},
		Marshaller:           &mock.MarshalizerMock{},
		AntifloodHandler:     &mock.P2PAntifloodHandlerStub{},
		Throttler:            &mock.ThrottlerStub{},
		ValidatorInfoPool:    testscommon.NewShardedDataStub(),
		ValidatorInfoStorage: &storage.StorerStub{},
		DataPacker:           &mock.DataPackerStub{},
		IsFullHistoryNode:    false,
	}
}

func createMockValidatorInfo(pk []byte) state.ValidatorInfo {
	return state.ValidatorInfo{
		PublicKey: pk,
		ShardId:   123,
		List:      string(common.EligibleList),
		Index:     10,
		Rating:    11,
	}
}

func TestNewValidatorInfoResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil SenderResolver should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.SenderResolver = nil

		res, err := resolvers.NewValidatorInfoResolver(args)
		assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.Marshaller = nil

		res, err := resolvers.NewValidatorInfoResolver(args)
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("nil AntifloodHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.AntifloodHandler = nil

		res, err := resolvers.NewValidatorInfoResolver(args)
		assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("nil Throttler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.Throttler = nil

		res, err := resolvers.NewValidatorInfoResolver(args)
		assert.Equal(t, dataRetriever.ErrNilThrottler, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("nil ValidatorInfoPool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = nil

		res, err := resolvers.NewValidatorInfoResolver(args)
		assert.Equal(t, dataRetriever.ErrNilValidatorInfoPool, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("nil ValidatorInfoStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoStorage = nil

		res, err := resolvers.NewValidatorInfoResolver(args)
		assert.Equal(t, dataRetriever.ErrNilValidatorInfoStorage, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("nil DataPacker should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.DataPacker = nil

		res, err := resolvers.NewValidatorInfoResolver(args)
		assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
		assert.True(t, check.IfNil(res))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		res, err := resolvers.NewValidatorInfoResolver(createMockArgValidatorInfoResolver())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(res))

		assert.Nil(t, res.Close())
	})
}

func TestValidatorInfoResolver_ProcessReceivedMessage(t *testing.T) {
	t.Parallel()

	t.Run("nil message should error", func(t *testing.T) {
		t.Parallel()

		res, _ := resolvers.NewValidatorInfoResolver(createMockArgValidatorInfoResolver())
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(nil, fromConnectedPeer)
		assert.Equal(t, dataRetriever.ErrNilMessage, err)
	})
	t.Run("canProcessMessage due to antiflood handler error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return expectedErr
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeer)
		assert.True(t, errors.Is(err, expectedErr))
		assert.False(t, args.Throttler.(*mock.ThrottlerStub).StartWasCalled())
		assert.False(t, args.Throttler.(*mock.ThrottlerStub).EndWasCalled())
	})
	t.Run("parseReceivedMessage returns error due to marshalizer error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.Marshaller = &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeer)
		assert.True(t, errors.Is(err, expectedErr))
	})

	t.Run("invalid request type should error", func(t *testing.T) {
		t.Parallel()

		res, _ := resolvers.NewValidatorInfoResolver(createMockArgValidatorInfoResolver())
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("hash")), fromConnectedPeer)
		assert.True(t, errors.Is(err, dataRetriever.ErrRequestTypeNotImplemented))
	})

	// resolveHashRequest
	t.Run("data not found in cache and fetchValidatorInfoByteSlice fails when getting data from storage", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		args.ValidatorInfoStorage = &storage.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, []byte("hash")), fromConnectedPeer)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("data found in cache but marshal fails", func(t *testing.T) {
		t.Parallel()

		marshMock := marshallerMock.MarshalizerMock{}
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return []byte("some value"), true
			},
		}
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return marshMock.Unmarshal(obj, buff)
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, []byte("hash")), fromConnectedPeer)
		assert.NotNil(t, err)
	})
	t.Run("data found in storage but marshal fails", func(t *testing.T) {
		t.Parallel()

		marshMock := marshallerMock.MarshalizerMock{}
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		args.ValidatorInfoStorage = &storage.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				return []byte("some value"), nil
			},
		}
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return marshMock.Unmarshal(obj, buff)
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, []byte("hash")), fromConnectedPeer)
		assert.NotNil(t, err)
	})
	t.Run("should work, data from cache", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedValue := createMockValidatorInfo([]byte("provided pk"))
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return providedValue, true
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				marshMock := marshallerMock.MarshalizerMock{}
				b := &batch.Batch{}
				_ = marshMock.Unmarshal(b, buff)

				vi := &state.ValidatorInfo{}
				_ = marshMock.Unmarshal(vi, b.Data[0])

				assert.Equal(t, &providedValue, vi)
				wasCalled = true

				return nil
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, []byte("hash")), fromConnectedPeer)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("should work, data from storage", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedValue := createMockValidatorInfo([]byte("provided pk"))
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		args.ValidatorInfoStorage = &storage.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				marshMock := marshallerMock.MarshalizerMock{}
				return marshMock.Marshal(providedValue)
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				marshMock := marshallerMock.MarshalizerMock{}
				b := &batch.Batch{}
				_ = marshMock.Unmarshal(b, buff)

				vi := &state.ValidatorInfo{}
				_ = marshMock.Unmarshal(vi, b.Data[0])

				assert.Equal(t, &providedValue, vi)
				wasCalled = true

				return nil
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, []byte("hash")), fromConnectedPeer)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})

	// resolveMultipleHashesRequest
	t.Run("unmarshal fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.Marshaller = &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				switch obj.(type) {
				case *dataRetriever.RequestData:
					return marshallerMock.MarshalizerMock{}.Unmarshal(obj, buff)
				case *batch.Batch:
					return expectedErr
				}
				return nil
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, []byte("hash")), fromConnectedPeer)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("no hash found", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		args.ValidatorInfoStorage = &storage.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				return nil, errors.New("not found")
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		b := &batch.Batch{
			Data: [][]byte{[]byte("hash")},
		}
		buff, _ := args.Marshaller.Marshal(b)
		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, buff), fromConnectedPeer)
		require.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), dataRetriever.ErrValidatorInfoNotFound.Error()))
	})
	t.Run("pack data in chunks returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				return key, true
			},
		}
		args.ValidatorInfoStorage = &storage.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				return nil, errors.New("not found")
			},
		}
		args.DataPacker = &mock.DataPackerStub{
			PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
				return nil, expectedErr
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		b := &batch.Batch{
			Data: [][]byte{[]byte("hash")},
		}
		buff, _ := args.Marshaller.Marshal(b)
		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, buff), fromConnectedPeer)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("send returns error", func(t *testing.T) {
		t.Parallel()

		numOfProvidedData := 3
		providedHashes := make([][]byte, 0)
		providedData := make([]state.ValidatorInfo, 0)
		for i := 0; i < numOfProvidedData; i++ {
			hashStr := fmt.Sprintf("hash%d", i)
			providedHashes = append(providedHashes, []byte(hashStr))
			pkStr := fmt.Sprintf("pk%d", i)
			providedData = append(providedData, createMockValidatorInfo([]byte(pkStr)))
		}
		args := createMockArgValidatorInfoResolver()
		numOfCalls := 0
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				val := providedData[numOfCalls]
				numOfCalls++
				return val, true
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				return expectedErr
			},
		}
		args.DataPacker, _ = partitioning.NewSimpleDataPacker(args.Marshaller)
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		buff, _ := args.Marshaller.Marshal(&batch.Batch{Data: providedHashes})
		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, buff), fromConnectedPeer)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("all hashes in one chunk should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		numOfProvidedData := 3
		providedHashes := make([][]byte, 0)
		providedData := make([]state.ValidatorInfo, 0)
		for i := 0; i < numOfProvidedData; i++ {
			hashStr := fmt.Sprintf("hash%d", i)
			providedHashes = append(providedHashes, []byte(hashStr))
			pkStr := fmt.Sprintf("pk%d", i)
			providedData = append(providedData, createMockValidatorInfo([]byte(pkStr)))
		}
		args := createMockArgValidatorInfoResolver()
		numOfCalls := 0
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				val := providedData[numOfCalls]
				numOfCalls++
				return val, true
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				marshMock := marshallerMock.MarshalizerMock{}
				b := &batch.Batch{}
				_ = marshMock.Unmarshal(b, buff)
				assert.Equal(t, numOfProvidedData, len(b.Data))

				for i := 0; i < numOfProvidedData; i++ {
					vi := &state.ValidatorInfo{}
					_ = marshMock.Unmarshal(vi, b.Data[i])

					assert.Equal(t, &providedData[i], vi)
				}

				wasCalled = true
				return nil
			},
		}
		args.DataPacker, _ = partitioning.NewSimpleDataPacker(args.Marshaller)
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		buff, _ := args.Marshaller.Marshal(&batch.Batch{Data: providedHashes})
		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, buff), fromConnectedPeer)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("multiple chunks should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		numOfProvidedData := 1000
		providedHashes := make([][]byte, 0)
		providedData := make([]state.ValidatorInfo, 0)
		testHasher := hashingMocks.HasherMock{}
		testMarshaller := marshallerMock.MarshalizerMock{}
		providedDataMap := make(map[string]struct{}, 0)
		for i := 0; i < numOfProvidedData; i++ {
			hashStr := fmt.Sprintf("hash%d", i)
			providedHashes = append(providedHashes, []byte(hashStr))
			pkStr := fmt.Sprintf("pk%d", i)
			newValidatorInfo := createMockValidatorInfo([]byte(pkStr))
			providedData = append(providedData, newValidatorInfo)

			buff, err := testMarshaller.Marshal(newValidatorInfo)
			require.Nil(t, err)
			hash := testHasher.Compute(string(buff))
			providedDataMap[string(hash)] = struct{}{}
		}
		numOfCalls := 0
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				val := providedData[numOfCalls]
				numOfCalls++
				return val, true
			},
		}
		numOfCallsSend := 0
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				marshMock := marshallerMock.MarshalizerMock{}
				b := &batch.Batch{}
				_ = marshMock.Unmarshal(b, buff)

				dataLen := len(b.Data)
				for i := 0; i < dataLen; i++ {
					vi := &state.ValidatorInfo{}
					_ = marshMock.Unmarshal(vi, b.Data[i])

					// remove this info from the provided map
					validatorInfoBuff, err := testMarshaller.Marshal(vi)
					require.Nil(t, err)
					hash := testHasher.Compute(string(validatorInfoBuff))
					delete(providedDataMap, string(hash))
				}

				numOfCallsSend++
				return nil
			},
		}
		args.DataPacker, _ = partitioning.NewSimpleDataPacker(args.Marshaller)
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		buff, _ := args.Marshaller.Marshal(&batch.Batch{Data: providedHashes})
		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, buff), fromConnectedPeer)
		assert.Nil(t, err)
		assert.Equal(t, 2, numOfCallsSend)       // ~677 messages in a chunk
		assert.Equal(t, 0, len(providedDataMap)) // all items should have been deleted on Send
	})
}

func TestValidatorInfoResolver_SetDebugHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	res, _ := resolvers.NewValidatorInfoResolver(createMockArgValidatorInfoResolver())
	require.False(t, check.IfNil(res))

	_ = res.SetDebugHandler(nil)
}
