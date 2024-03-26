package sovereign_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/sovereign"
	sovereignMock "github.com/multiversx/mx-chain-go/testscommon/sovereign"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignDataCodecFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil data codec should fail", func(t *testing.T) {
		sovereignDataCodecFactory, err := sovereign.NewSovereignDataCodecFactory(nil)
		require.Equal(t, errors.ErrNilDataCodec, err)
		require.True(t, sovereignDataCodecFactory.IsInterfaceNil())
	})

	t.Run("should work", func(t *testing.T) {
		sovereignDataCodecFactory, err := sovereign.NewSovereignDataCodecFactory(&sovereignMock.DataCodecMock{})
		require.Nil(t, err)
		require.False(t, sovereignDataCodecFactory.IsInterfaceNil())
	})
}

func TestSovereignDataCodecFactory_CreateDataCodec(t *testing.T) {
	t.Parallel()

	sovereignDataCodecFactory, _ := sovereign.NewSovereignDataCodecFactory(&sovereignMock.DataCodecMock{})
	sovereignDataCodec := sovereignDataCodecFactory.CreateDataCodec()
	require.NotNil(t, sovereignDataCodec)
	require.Equal(t, "*sovereign.DataCodecMock", fmt.Sprintf("%T", sovereignDataCodec))
}
