package enableEpochs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFlagsHolder_NilFlagShouldPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	fh := newFlagsHolder()
	require.NotNil(t, fh)

	fh.scDeployFlag = nil
	fh.IsSCDeployFlagEnabled()
}
