package metachain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEconomicsFactory_CreateEndOfEpochEconomics(t *testing.T) {
	t.Parallel()

	f := NewEconomicsFactory()
	require.False(t, f.IsInterfaceNil())

	args := createMockEpochEconomicsArguments()
	econ, err := f.CreateEndOfEpochEconomics(args)
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%T", econ), "*metachain.economics")
}
