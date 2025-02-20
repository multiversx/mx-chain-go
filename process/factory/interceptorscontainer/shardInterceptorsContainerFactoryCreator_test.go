package interceptorscontainer_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/stretchr/testify/require"
)

func TestNewShardInterceptorsContainerFactoryCreator(t *testing.T) {
	t.Parallel()

	factoryCreator := interceptorscontainer.NewShardInterceptorsContainerFactoryCreator()
	require.Implements(t, new(interceptorscontainer.InterceptorsContainerFactoryCreator), factoryCreator)
	require.False(t, check.IfNil(factoryCreator))
}

func TestShardInterceptorsContainerFactoryCreator_CreateInterceptorsContainerFactory(t *testing.T) {
	t.Parallel()

	factoryCreator := interceptorscontainer.NewShardInterceptorsContainerFactoryCreator()
	coreComp, cryptoComp := createMockComponentHolders()
	args := getArgumentsShard(coreComp, cryptoComp)
	factory, err := factoryCreator.CreateInterceptorsContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*interceptorscontainer.shardInterceptorsContainerFactory", fmt.Sprintf("%T", factory))
}
