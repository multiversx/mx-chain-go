package dataRetriever_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/require"
)

func TestSetEpochHandlerToHdrResolver_GetErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	resolverContainer := &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver dataRetriever.Resolver, err error) {
			return nil, localErr
		},
	}
	epochHandler := &mock.EpochHandlerStub{}

	err := dataRetriever.SetEpochHandlerToHdrResolver(resolverContainer, epochHandler)
	require.Equal(t, localErr, err)
}

func TestSetEpochHandlerToHdrResolver_CannotSetEpoch(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	resolverContainer := &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{
				SetEpochHandlerCalled: func(epochHandler dataRetriever.EpochHandler) error {
					return localErr
				},
			}, nil
		},
	}
	epochHandler := &mock.EpochHandlerStub{}

	err := dataRetriever.SetEpochHandlerToHdrResolver(resolverContainer, epochHandler)
	require.Equal(t, localErr, err)
}

func TestSetEpochHandlerToHdrResolver_WrongType(t *testing.T) {
	t.Parallel()

	resolverContainer := &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver dataRetriever.Resolver, err error) {
			return nil, nil
		},
	}
	epochHandler := &mock.EpochHandlerStub{}

	err := dataRetriever.SetEpochHandlerToHdrResolver(resolverContainer, epochHandler)
	require.Equal(t, dataRetriever.ErrWrongTypeInContainer, err)
}

func TestSetEpochHandlerToHdrResolver_Ok(t *testing.T) {
	t.Parallel()

	resolverContainer := &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}
	epochHandler := &mock.EpochHandlerStub{}

	err := dataRetriever.SetEpochHandlerToHdrResolver(resolverContainer, epochHandler)
	require.Nil(t, err)
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
