package resolvers_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgValidatorInfoResolver() resolvers.ArgValidatorInfoResolver {
	return resolvers.ArgValidatorInfoResolver{
		SenderResolver:       &mock.TopicResolverSenderStub{},
		Marshalizer:          &mock.MarshalizerMock{},
		AntifloodHandler:     &mock.P2PAntifloodHandlerStub{},
		Throttler:            &mock.ThrottlerStub{},
		ValidatorInfoPool:    testscommon.NewCacherStub(),
		ValidatorInfoStorage: &storage.StorerStub{},
		IsFullHistoryNode:    false,
	}
}

func createMockValidatorInfo() state.ValidatorInfo {
	return state.ValidatorInfo{
		PublicKey: []byte("provided pk"),
		ShardId:   123,
		List:      string(common.EligibleList),
		Index:     10,
		Rating:    10,
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
	t.Run("nil Marshalizer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoResolver()
		args.Marshalizer = nil

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
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		res, err := resolvers.NewValidatorInfoResolver(createMockArgValidatorInfoResolver())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(res))

		assert.Nil(t, res.Close())
	})
}

func TestValidatorInfoResolver_RequestDataFromHash(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockArgValidatorInfoResolver()
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
				return expectedErr
			},
		}

		res, _ := resolvers.NewValidatorInfoResolver(args)
		err := res.RequestDataFromHash(nil, 0)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("provided hash")
		providedEpoch := uint32(123)
		args := createMockArgValidatorInfoResolver()
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
				assert.Equal(t, providedHash, originalHashes[0])
				assert.Equal(t, dataRetriever.HashType, rd.Type)
				assert.Equal(t, providedHash, rd.Value)
				assert.Equal(t, providedEpoch, rd.Epoch)

				return nil
			},
		}

		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.RequestDataFromHash(providedHash, providedEpoch)
		assert.Nil(t, err)
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

		expectedErr := errors.New("expected err")
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
		assert.False(t, args.Throttler.(*mock.ThrottlerStub).StartWasCalled)
		assert.False(t, args.Throttler.(*mock.ThrottlerStub).EndWasCalled)
	})
	t.Run("parseReceivedMessage returns error due to marshalizer error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockArgValidatorInfoResolver()
		args.Marshalizer = &mock.MarshalizerStub{
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

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, []byte("hash")), fromConnectedPeer)
		assert.True(t, errors.Is(err, dataRetriever.ErrRequestTypeNotImplemented))
	})
	t.Run("data not found in cache and fetchValidatorInfoByteSlice fails when getting data from storage", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
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

		expectedErr := errors.New("expected err")
		marshalizerMock := testscommon.MarshalizerMock{}
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return []byte("some value"), true
			},
		}
		args.Marshalizer = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return marshalizerMock.Unmarshal(obj, buff)
			},
		}
		res, _ := resolvers.NewValidatorInfoResolver(args)
		require.False(t, check.IfNil(res))

		err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, []byte("hash")), fromConnectedPeer)
		assert.NotNil(t, err)
	})
	t.Run("data found in storage but marshal fails", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		marshalizerMock := testscommon.MarshalizerMock{}
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		args.ValidatorInfoStorage = &storage.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				return []byte("some value"), nil
			},
		}
		args.Marshalizer = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return marshalizerMock.Unmarshal(obj, buff)
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
		providedValue := createMockValidatorInfo()
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return providedValue, true
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				marshalizerMock := testscommon.MarshalizerMock{}
				b := &batch.Batch{}
				_ = marshalizerMock.Unmarshal(b, buff)

				vi := &state.ValidatorInfo{}
				_ = marshalizerMock.Unmarshal(vi, b.Data[0])

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
		providedValue := createMockValidatorInfo()
		args := createMockArgValidatorInfoResolver()
		args.ValidatorInfoPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		args.ValidatorInfoStorage = &storage.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				marshalizerMock := testscommon.MarshalizerMock{}
				return marshalizerMock.Marshal(providedValue)
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				marshalizerMock := testscommon.MarshalizerMock{}
				b := &batch.Batch{}
				_ = marshalizerMock.Unmarshal(b, buff)

				vi := &state.ValidatorInfo{}
				_ = marshalizerMock.Unmarshal(vi, b.Data[0])

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
}

func TestValidatorInfoResolver_SetResolverDebugHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	res, _ := resolvers.NewValidatorInfoResolver(createMockArgValidatorInfoResolver())
	require.False(t, check.IfNil(res))

	_ = res.SetResolverDebugHandler(nil)
}

func TestValidatorInfoResolver_NumPeersToQuery(t *testing.T) {
	t.Parallel()

	providedIntra, providedCross := 5, 10
	receivedIntra, receivedCross := 0, 0
	args := createMockArgValidatorInfoResolver()
	args.SenderResolver = &mock.TopicResolverSenderStub{
		SetNumPeersToQueryCalled: func(intra int, cross int) {
			assert.Equal(t, providedIntra, intra)
			assert.Equal(t, providedCross, cross)
			receivedIntra = intra
			receivedCross = cross
		},
		GetNumPeersToQueryCalled: func() (int, int) {
			return receivedIntra, receivedCross
		},
	}

	res, _ := resolvers.NewValidatorInfoResolver(args)
	require.False(t, check.IfNil(res))

	res.SetNumPeersToQuery(providedIntra, providedCross)
	intra, cross := res.NumPeersToQuery()
	assert.Equal(t, providedIntra, intra)
	assert.Equal(t, providedCross, cross)
}
