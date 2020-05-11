package epochproviders

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/require"
)

func TestSimpleEpochProviderByNonce_EpochForNonce(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	sep := NewSimpleEpochProviderByNonce(&mock.EpochHandlerStub{
		MetaEpochCalled: func() uint32 {
			return epoch
		},
	})
	require.False(t, check.IfNil(sep))

	resEpoch, err := sep.EpochForNonce(0)
	require.Nil(t, err)
	require.Equal(t, epoch, resEpoch)
}
