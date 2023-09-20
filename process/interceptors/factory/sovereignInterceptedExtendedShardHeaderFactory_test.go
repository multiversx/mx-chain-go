package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/require"
)

func createSovArgs() ArgsSovereignInterceptedExtendedHeaderFactory {
	return ArgsSovereignInterceptedExtendedHeaderFactory{
		Marshaller: &mock.MarshalizerMock{},
		Hasher:     &hashingMocks.HasherMock{},
	}
}

func TestNewInterceptedShardHeaderDataFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createSovArgs()
		args.Marshaller = nil
		sovFactory, err := NewSovereignInterceptedShardHeaderDataFactory(args)
		require.Equal(t, process.ErrNilMarshalizer, err)
		require.Nil(t, sovFactory)
	})

	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createSovArgs()
		args.Hasher = nil
		sovFactory, err := NewSovereignInterceptedShardHeaderDataFactory(args)
		require.Equal(t, process.ErrNilHasher, err)
		require.Nil(t, sovFactory)
	})

	t.Run("should work", func(t *testing.T) {
		args := createSovArgs()
		sovFactory, err := NewSovereignInterceptedShardHeaderDataFactory(args)
		require.Nil(t, err)
		require.False(t, sovFactory.IsInterfaceNil())
	})
}

func TestSovereignInterceptedShardHeaderDataFactory_Create(t *testing.T) {
	t.Parallel()

	extendedHeader := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Nonce: 4,
			},
		},
	}

	args := createSovArgs()
	headerBytes, err := args.Marshaller.Marshal(extendedHeader)
	require.Nil(t, err)

	sovFactory, _ := NewSovereignInterceptedShardHeaderDataFactory(args)
	interceptedData, err := sovFactory.Create(headerBytes)
	require.Nil(t, err)

	interceptedSovHeader, castOk := interceptedData.(process.ExtendedHeaderValidatorHandler)
	require.True(t, castOk)
	require.Equal(t, extendedHeader, interceptedSovHeader.GetExtendedHeader())
}
