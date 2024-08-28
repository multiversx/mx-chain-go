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
}

func TestOneShardCoordinator_ComputeIdShouldWork(t *testing.T) {
	t.Parallel()

	oneShardCoordinator := OneShardCoordinator{}
	returnedVal := oneShardCoordinator.ComputeId([]byte{})
	require.Equal(t, uint32(0), returnedVal)
}

func TestOneShardCoordinator_SelfIdShouldWork(t *testing.T) {
	t.Parallel()

	oneShardCoordinator := OneShardCoordinator{}
	returnedVal := oneShardCoordinator.SelfId()
	require.Equal(t, uint32(0), returnedVal)
}

func TestOneShardCoordinator_SameShardShouldWork(t *testing.T) {
	t.Parallel()

	oneShardCoordinator := OneShardCoordinator{}
	returnedVal := oneShardCoordinator.SameShard(nil, nil)
	require.True(t, returnedVal)
}

func TestOneShardCoordinator_CommunicationIdentifierShouldWork(t *testing.T) {
	t.Parallel()

	oneShardCoordinator := OneShardCoordinator{}
	returnedVal := oneShardCoordinator.CommunicationIdentifier(0)
	require.Equal(t, core.CommunicationIdentifierBetweenShards(0, 0), returnedVal)
}

func TestOneShardCoordinator_IsInterfaceNilShouldWork(t *testing.T) {
	t.Parallel()

	oneShardCoordinator := OneShardCoordinator{}
	returnedVal := oneShardCoordinator.IsInterfaceNil()
	require.False(t, returnedVal)
}
