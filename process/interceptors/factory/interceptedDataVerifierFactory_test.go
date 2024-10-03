package factory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createMockArgInterceptedDataVerifierFactory() InterceptedDataVerifierFactoryArgs {
	return InterceptedDataVerifierFactoryArgs{
		CacheSpan:   time.Second,
		CacheExpiry: time.Second,
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

	factory := NewInterceptedDataVerifierFactory(createMockArgInterceptedDataVerifierFactory())
	require.NotNil(t, factory)

	interceptedDataVerifier, err := factory.Create("mockTopic")
	require.NoError(t, err)

	require.False(t, interceptedDataVerifier.IsInterfaceNil())
}
