package dataRetriever_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("err")

func TestSetEpochHandlerToHdrResolver(t *testing.T) {
	t.Parallel()

	t.Run("get function errors should return error", func(t *testing.T) {
		t.Parallel()

		resolverContainer := &dataRetrieverMock.ResolversContainerStub{
			GetCalled: func(key string) (resolver dataRetriever.Resolver, err error) {
				return nil, expectedErr
			},
		}
		epochHandler := &mock.EpochHandlerStub{}

		err := dataRetriever.SetEpochHandlerToHdrResolver(resolverContainer, epochHandler)
		require.Equal(t, expectedErr, err)
	})
	t.Run("set epoch handler errors should return error", func(t *testing.T) {
		t.Parallel()

		resolverContainer := &dataRetrieverMock.ResolversContainerStub{
			GetCalled: func(key string) (resolver dataRetriever.Resolver, err error) {
				return &mock.HeaderResolverStub{
					SetEpochHandlerCalled: func(epochHandler dataRetriever.EpochHandler) error {
						return expectedErr
					},
				}, nil
			},
		}
		epochHandler := &mock.EpochHandlerStub{}

		err := dataRetriever.SetEpochHandlerToHdrResolver(resolverContainer, epochHandler)
		require.Equal(t, expectedErr, err)
	})
	t.Run("wrong type should return error", func(t *testing.T) {
		t.Parallel()

		resolverContainer := &dataRetrieverMock.ResolversContainerStub{
			GetCalled: func(key string) (resolver dataRetriever.Resolver, err error) {
				return nil, nil
			},
		}
		epochHandler := &mock.EpochHandlerStub{}

		err := dataRetriever.SetEpochHandlerToHdrResolver(resolverContainer, epochHandler)
		require.Equal(t, dataRetriever.ErrWrongTypeInContainer, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		resolverContainer := &dataRetrieverMock.ResolversContainerStub{
			GetCalled: func(key string) (resolver dataRetriever.Resolver, err error) {
				return &mock.HeaderResolverStub{}, nil
			},
		}
		epochHandler := &mock.EpochHandlerStub{}

		err := dataRetriever.SetEpochHandlerToHdrResolver(resolverContainer, epochHandler)
		require.Nil(t, err)
	})
}

func TestSetEpochHandlerToHdrRequester(t *testing.T) {
	t.Parallel()

	t.Run("get function errors should return error", func(t *testing.T) {
		t.Parallel()

		requestersContainer := &dataRetrieverMock.RequestersContainerStub{
			GetCalled: func(key string) (requester dataRetriever.Requester, err error) {
				return nil, expectedErr
			},
		}
		epochHandler := &mock.EpochHandlerStub{}

		err := dataRetriever.SetEpochHandlerToHdrRequester(requestersContainer, epochHandler)
		require.Equal(t, expectedErr, err)
	})
	t.Run("set epoch handler errors should return error", func(t *testing.T) {
		t.Parallel()

		requestersContainer := &dataRetrieverMock.RequestersContainerStub{
			GetCalled: func(key string) (requester dataRetriever.Requester, err error) {
				return &mock.HeaderRequesterStub{
					SetEpochHandlerCalled: func(epochHandler dataRetriever.EpochHandler) error {
						return expectedErr
					},
				}, nil
			},
		}
		epochHandler := &mock.EpochHandlerStub{}

		err := dataRetriever.SetEpochHandlerToHdrRequester(requestersContainer, epochHandler)
		require.Equal(t, expectedErr, err)
	})
	t.Run("wrong type should return error", func(t *testing.T) {
		t.Parallel()

		requestersContainer := &dataRetrieverMock.RequestersContainerStub{
			GetCalled: func(key string) (requester dataRetriever.Requester, err error) {
				return nil, nil
			},
		}
		epochHandler := &mock.EpochHandlerStub{}

		err := dataRetriever.SetEpochHandlerToHdrRequester(requestersContainer, epochHandler)
		require.Equal(t, dataRetriever.ErrWrongTypeInContainer, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		requestersContainer := &dataRetrieverMock.RequestersContainerStub{
			GetCalled: func(key string) (resolver dataRetriever.Requester, err error) {
				return &mock.HeaderRequesterStub{}, nil
			},
		}
		epochHandler := &mock.EpochHandlerStub{}

		err := dataRetriever.SetEpochHandlerToHdrRequester(requestersContainer, epochHandler)
		require.Nil(t, err)
	})
}

func TestGetHdrNonceHashDataUnit(t *testing.T) {
	t.Parallel()

	require.Equal(t, dataRetriever.ShardHdrNonceHashDataUnit, dataRetriever.GetHdrNonceHashDataUnit(0))
	require.Equal(t, dataRetriever.ShardHdrNonceHashDataUnit+1, dataRetriever.GetHdrNonceHashDataUnit(1))
	require.Equal(t, dataRetriever.MetaHdrNonceHashDataUnit, dataRetriever.GetHdrNonceHashDataUnit(core.MetachainShardId))
}

func TestGetHeadersDataUnit(t *testing.T) {
	t.Parallel()

	require.Equal(t, dataRetriever.BlockHeaderUnit, dataRetriever.GetHeadersDataUnit(0))
	require.Equal(t, dataRetriever.BlockHeaderUnit, dataRetriever.GetHeadersDataUnit(1))
	require.Equal(t, dataRetriever.MetaBlockUnit, dataRetriever.GetHeadersDataUnit(core.MetachainShardId))
}
