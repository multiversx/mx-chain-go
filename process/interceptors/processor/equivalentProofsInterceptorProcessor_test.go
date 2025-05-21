package processor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEquivalentProofsInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var epip *equivalentProofsInterceptorProcessor
	require.True(t, epip.IsInterfaceNil())

	epip = NewEquivalentProofsInterceptorProcessor()
	require.False(t, epip.IsInterfaceNil())
}

func TestNewEquivalentProofsInterceptorProcessor(t *testing.T) {
	t.Parallel()

	epip := NewEquivalentProofsInterceptorProcessor()
	require.NotNil(t, epip)

	// coverage only
	require.Nil(t, epip.Validate(nil, ""))

	// coverage only
	err := epip.Save(nil, "", "")
	require.Nil(t, err)

	// coverage only
	epip.RegisterHandler(nil)
}
