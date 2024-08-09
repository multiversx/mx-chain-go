package preprocess

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignRewardsTxPreProcFactory_CreateRewardsTxPreProcessor(t *testing.T) {
	t.Parallel()

	f := NewSovereignRewardsTxPreProcFactory()
	require.False(t, f.IsInterfaceNil())

	args := createArgsRewardsPreProc()
	preProc, err := f.CreateRewardsTxPreProcessor(args)
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%T", preProc), "*preprocess.sovereignRewardsTxPreProcessor")
}
