package facade

import (
	"context"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common/checkpoint"
	"github.com/multiversx/mx-chain-go/facade/mock"
	"github.com/stretchr/testify/require"
)

func TestTrieScanUsesEarlierDeadline(t *testing.T) {
	for _, scanTimeout := range []uint32{100, 10000} {
		args := createMockArguments()
		args.WsAntifloodConfig.WebServerAntifloodEnabled = true
		args.WsAntifloodConfig.TrieOperationsDeadlineMilliseconds = scanTimeout
		parentDeadline := time.Now().Add(5 * time.Second)
		parent, cancel := context.WithDeadline(context.Background(), parentDeadline)
		var scanContext context.Context
		args.Node = &mock.NodeStub{GetKeyValuePairsCalled: func(_ string, _ api.AccountQueryOptions, _ uint, _ *checkpoint.Checkpoint, ctx context.Context) (map[string]string, *checkpoint.Checkpoint, api.BlockInfo, error) {
			scanContext = ctx
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			if scanTimeout == 100 {
				require.True(t, deadline.Before(parentDeadline))
				require.LessOrEqual(t, time.Until(deadline), 100*time.Millisecond)
			} else {
				require.Equal(t, parentDeadline, deadline)
			}
			return nil, nil, api.BlockInfo{}, nil
		}}
		nf, err := NewNodeFacade(args)
		require.NoError(t, err)
		_, _, _, err = nf.GetKeyValuePairs(parent, "alice", api.AccountQueryOptions{}, 0, nil)
		require.NoError(t, err)
		require.NotNil(t, scanContext)
		require.ErrorIs(t, scanContext.Err(), context.Canceled)
		require.NoError(t, parent.Err())
		cancel()
	}
}
