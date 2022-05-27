package enableEpochs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFlagsHolder_NilFlagShouldPanic(t *testing.T) {
	t.Parallel()

	fh := newFlagsHolder()
	require.NotNil(t, fh)

	fh.scDeployFlag = nil
	require.Panicsf(t, func() { fh.IsSCDeployFlagEnabled() }, "")
}
