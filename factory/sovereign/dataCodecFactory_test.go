package sovereign_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/factory/sovereign"

	"github.com/stretchr/testify/require"
)

func TestNewDataCodecFactory(t *testing.T) {
	t.Parallel()

	dataCodecFactory := sovereign.NewDataCodecFactory()
	require.False(t, dataCodecFactory.IsInterfaceNil())
}

func TestDataCodecFactory_CreateDataCodec(t *testing.T) {
	t.Parallel()

	dataCodecFactory := sovereign.NewDataCodecFactory()
	dataCodec := dataCodecFactory.CreateDataCodec()
	require.NotNil(t, dataCodec)
	require.Equal(t, "*disabled.dataCodec", fmt.Sprintf("%T", dataCodec))
}
