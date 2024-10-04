package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestOneShardCoordinator_NumberOfShardsShouldWork(t *testing.T) {
	t.Parallel()

	oneShardCoordinator := OneShardCoordinator{}

	returnedVal := oneShardCoordinator.NumberOfShards()
	require.Equal(t, uint32(1), returnedVal)

	returnedVal = oneShardCoordinator.ComputeId([]byte{})
	require.Equal(t, uint32(0), returnedVal)

	returnedVal = oneShardCoordinator.SelfId()
	require.Equal(t, uint32(0), returnedVal)

	isShameShard := oneShardCoordinator.SameShard(nil, nil)
	require.True(t, isShameShard)

	communicationID := oneShardCoordinator.CommunicationIdentifier(0)
	require.Equal(t, core.CommunicationIdentifierBetweenShards(0, 0), communicationID)

	isInterfaceNil := oneShardCoordinator.IsInterfaceNil()
	require.False(t, isInterfaceNil)

}
