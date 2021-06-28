package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestEpochHandler_Epoch(t *testing.T) {
	t.Parallel()

	disabledEpochHandler := NewEpochHandler()

	require.False(t, check.IfNil(disabledEpochHandler))
	require.Equal(t, uint32(0), disabledEpochHandler.MetaEpoch())
}
