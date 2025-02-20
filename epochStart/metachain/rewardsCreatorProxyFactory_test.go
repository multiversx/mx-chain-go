package metachain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRewardsCreatorFactory_CreateRewardsCreator(t *testing.T) {
	t.Parallel()

	args := createDefaultRewardsCreatorProxyArgs()
	f := NewRewardsCreatorFactory()
	require.False(t, f.IsInterfaceNil())

	creator, err := f.CreateRewardsCreator(args)
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%T", creator), "*metachain.rewardsCreatorProxy")
}
