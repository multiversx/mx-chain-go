package broadcastFactory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChainMessengerFactory_CreateShardChainMessenger(t *testing.T) {
	t.Parallel()

	f := NewShardChainMessengerFactory()
	require.False(t, f.IsInterfaceNil())

	args := createDefaultShardChainArgs()
	msg, err := f.CreateShardChainMessenger(args)
	require.Nil(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "*broadcast.shardChainMessenger", fmt.Sprintf("%T", msg))
}
