package metachain

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestSysSCFactory_CreateSystemSCProcessor(t *testing.T) {
	t.Parallel()

	f := NewSysSCFactory()
	require.False(t, check.IfNil(f))

	args := createMockArgsForSystemSCProcessor()
	sysSC, err := f.CreateSystemSCProcessor(args)
	require.Nil(t, err)
	require.Equal(t, "*metachain.systemSCProcessor", fmt.Sprintf("%T", sysSC))
}
