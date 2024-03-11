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
		_, _ = dc.SerializeEventData(sovereign.EventData{})
		_, _ = dc.DeserializeEventData([]byte("data"))
		_, _ = dc.SerializeTokenData(sovereign.EsdtTokenData{})
		_, _ = dc.DeserializeTokenData([]byte("data"))
		_, _ = dc.GetTokenDataBytes([]byte{0x00}, []byte("data"))
		_, _ = dc.SerializeOperation(sovereign.Operation{})
		_ = dc.IsInterfaceNil()
	})
}

func TestDataCodec_SerializeEventData(t *testing.T) {
	t.Parallel()

	dc := NewDisabledDataCodec()
	require.False(t, check.IfNil(dc))

	serialized, err := dc.SerializeEventData(sovereign.EventData{})
	require.NotNil(t, serialized)
	require.NoError(t, err)
}

func TestDataCodec_DeserializeEventData(t *testing.T) {
	t.Parallel()

	dc := NewDisabledDataCodec()
	require.False(t, check.IfNil(dc))

	deserialized, err := dc.DeserializeEventData([]byte("data"))
	require.NotNil(t, deserialized)
	require.NoError(t, err)
}

func TestDataCodec_SerializeTokenData(t *testing.T) {
	t.Parallel()

	dc := NewDisabledDataCodec()
	require.False(t, check.IfNil(dc))

	serialized, err := dc.SerializeTokenData(sovereign.EsdtTokenData{})
	require.NotNil(t, serialized)
	require.NoError(t, err)
}

func TestDataCodec_DeserializeTokenData(t *testing.T) {
	t.Parallel()

	dc := NewDisabledDataCodec()
	require.False(t, check.IfNil(dc))

	deserialized, err := dc.DeserializeTokenData([]byte("data"))
	require.NotNil(t, deserialized)
	require.NoError(t, err)
}

func TestDataCodec_GetTokenDataBytes(t *testing.T) {
	t.Parallel()

	dc := NewDisabledDataCodec()
	require.False(t, check.IfNil(dc))

	deserialized, err := dc.GetTokenDataBytes([]byte{0x00}, []byte("data"))
	require.NotNil(t, deserialized)
	require.NoError(t, err)
}

func TestDataCodec_SerializeOperation(t *testing.T) {
	t.Parallel()

	dc := NewDisabledDataCodec()
	require.False(t, check.IfNil(dc))

	deserialized, err := dc.SerializeOperation(sovereign.Operation{})
	require.NotNil(t, deserialized)
	require.NoError(t, err)
}
