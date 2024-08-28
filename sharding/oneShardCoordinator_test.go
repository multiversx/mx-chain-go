package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestOneShardCoordinator_NumberOfShardsShouldWork(t *testing.T) {
	t.Parallel()

	oneShardCoordinator := OneShardCoordinator{}

	t.Run("NumberOfShardsShouldWork", func(t *testing.T) {
		t.Parallel()

		returnedVal := oneShardCoordinator.NumberOfShards()
		require.Equal(t, uint32(1), returnedVal)
	})

	t.Run("ComputeIdShouldWork", func(t *testing.T) {
		t.Parallel()

		returnedVal := oneShardCoordinator.ComputeId([]byte{})
		require.Equal(t, uint32(0), returnedVal)
	})

	t.Run("SelfIdShouldWork", func(t *testing.T) {
		t.Parallel()

		returnedVal := oneShardCoordinator.SelfId()
		require.Equal(t, uint32(0), returnedVal)
	})

	t.Run("SameShardShouldWork", func(t *testing.T) {
		t.Parallel()

		returnedVal := oneShardCoordinator.SameShard(nil, nil)
		require.True(t, returnedVal)
	})

	t.Run("CommunicationIdentifierShouldWork", func(t *testing.T) {
		t.Parallel()

		returnedVal := oneShardCoordinator.CommunicationIdentifier(0)
		require.Equal(t, core.CommunicationIdentifierBetweenShards(0, 0), returnedVal)
	})

	t.Run("IsInterfaceNilShouldWork", func(t *testing.T) {
		t.Parallel()

		returnedVal := oneShardCoordinator.IsInterfaceNil()
		require.False(t, returnedVal)
	})

}
