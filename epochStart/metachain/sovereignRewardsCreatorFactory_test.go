package metachain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignRewardsCreatorFactory_CreateRewardsCreator(t *testing.T) {
	t.Parallel()

	args := createDefaultRewardsCreatorProxyArgs()
	f := NewSovereignRewardsCreatorFactory()
	require.False(t, f.IsInterfaceNil())

	creator, err := f.CreateRewardsCreator(args)
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%T", creator), "*metachain.sovereignRewards")
}
