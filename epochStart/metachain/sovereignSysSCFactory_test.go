package metachain

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestSovereignSysSCFactory_CreateSystemSCProcessor(t *testing.T) {
	t.Parallel()

	f := NewSovereignSysSCFactory()
	require.False(t, check.IfNil(f))

	args := createMockArgsForSystemSCProcessor()
	sysSC, err := f.CreateSystemSCProcessor(args)
	require.Nil(t, err)
	require.Equal(t, "*metachain.sovereignSystemSC", fmt.Sprintf("%T", sysSC))
}
