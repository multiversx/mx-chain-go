package cutoff

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestDisabledBlockProcessingCutoff_FunctionsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.Nil(t, r)
	}()
	d := NewDisabledBlockProcessingCutoff()

	d.HandlePauseCutoff(&block.MetaBlock{Nonce: 37})
	err := d.HandleProcessErrorCutoff(&block.MetaBlock{Round: 37})
	require.NoError(t, err)
	require.False(t, d.IsInterfaceNil())

	var nilObj *disabledBlockProcessingCutoff
	require.True(t, nilObj.IsInterfaceNil())
}
