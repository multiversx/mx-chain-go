package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"
)

func TestDataCodec_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	dc := NewDisabledDataCodec()
	require.False(t, check.IfNil(dc))

	require.NotPanics(t, func() {
		serializedEventData, err := dc.SerializeEventData(sovereign.EventData{})
		require.Equal(t, make([]byte, 0), serializedEventData)
		require.NoError(t, err)

		deserializedEventData, err := dc.DeserializeEventData([]byte("data"))
		require.Equal(t, &sovereign.EventData{}, deserializedEventData)
		require.NoError(t, err)

		serializedTokenData, err := dc.SerializeTokenData(sovereign.EsdtTokenData{})
		require.Equal(t, make([]byte, 0), serializedTokenData)
		require.NoError(t, err)

		deserializedTokenData, err := dc.DeserializeTokenData([]byte("data"))
		require.Equal(t, &sovereign.EsdtTokenData{}, deserializedTokenData)
		require.NoError(t, err)

		deserializedOperation, err := dc.SerializeOperation(sovereign.Operation{})
		require.Equal(t, make([]byte, 0), deserializedOperation)
		require.NoError(t, err)
	})
}
