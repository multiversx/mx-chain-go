package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func createMockArgInterceptedDataVerifierFactory() InterceptedDataVerifierFactoryArgs {
	return InterceptedDataVerifierFactoryArgs{
		InterceptedDataVerifierConfig: config.InterceptedDataVerifierConfig{
			EnableCaching:    true,
			CacheSpanInSec:   1,
			CacheExpiryInSec: 1,
		},
	}
}

func TestInterceptedDataVerifierFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var factory *interceptedDataVerifierFactory
	require.True(t, factory.IsInterfaceNil())

	factory = NewInterceptedDataVerifierFactory(createMockArgInterceptedDataVerifierFactory())
	require.False(t, factory.IsInterfaceNil())
}

func TestNewInterceptedDataVerifierFactory(t *testing.T) {
	t.Parallel()

	factory := NewInterceptedDataVerifierFactory(createMockArgInterceptedDataVerifierFactory())
	require.NotNil(t, factory)
}

func TestInterceptedDataVerifierFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("should fail if cache creation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedDataVerifierFactory()
		args.InterceptedDataVerifierConfig.CacheSpanInSec = 0
		factory := NewInterceptedDataVerifierFactory(args)
		require.NotNil(t, factory)

		interceptedDataVerifier, err := factory.Create("mockTopic")
		require.Error(t, err)
		require.Nil(t, interceptedDataVerifier)
	})
	t.Run("should work with caching enabled", func(t *testing.T) {
		t.Parallel()

		factory := NewInterceptedDataVerifierFactory(createMockArgInterceptedDataVerifierFactory())
		require.NotNil(t, factory)

		interceptedDataVerifier, err := factory.Create("mockTopic")
		require.NoError(t, err)

		require.False(t, interceptedDataVerifier.IsInterfaceNil())
		require.NoError(t, factory.Close())
	})
	t.Run("should work with caching disabled", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedDataVerifierFactory()
		args.InterceptedDataVerifierConfig.EnableCaching = false
		factory := NewInterceptedDataVerifierFactory(args)
		require.NotNil(t, factory)

		interceptedDataVerifier, err := factory.Create("mockTopic")
		require.NoError(t, err)

		require.False(t, interceptedDataVerifier.IsInterfaceNil())
	})
}
