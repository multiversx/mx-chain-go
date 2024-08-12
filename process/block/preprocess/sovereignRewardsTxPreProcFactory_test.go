package preprocess

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/stretchr/testify/require"
)

func TestSovereignRewardsTxPreProcFactory_CreateRewardsTxPreProcessor(t *testing.T) {
	t.Parallel()

	f := NewSovereignRewardsTxPreProcFactory()
	require.False(t, f.IsInterfaceNil())

	args := createArgsRewardsPreProc()

	container := containers.NewPreProcessorsContainer()
	err := f.CreateRewardsTxPreProcessorAndAddToContainer(args, container)
	require.Nil(t, err)

	preProc, err := container.Get(block.RewardsBlock)
	require.True(t, strings.Contains(err.Error(), process.ErrInvalidContainerKey.Error()))
	require.Nil(t, preProc)
}
