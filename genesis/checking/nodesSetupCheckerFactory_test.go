package checking_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/genesis/checking"
	"github.com/stretchr/testify/require"
)

func TestNodesSetupCheckerFactory_CreateNodesSetupChecker(t *testing.T) {
	t.Parallel()

	factory := checking.NewNodesSetupCheckerFactory()
	require.False(t, factory.IsInterfaceNil())

	args := createArgs()
	nodesSetupChecker, err := factory.CreateNodesSetupChecker(args)
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%T", nodesSetupChecker), "*checking.nodeSetupChecker")
}
