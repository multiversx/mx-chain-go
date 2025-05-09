package graceperiod

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
)

func TestNewEpochChangeGracePeriod(t *testing.T) {
	t.Parallel()

	t.Run("should return error for empty config", func(t *testing.T) {
		t.Parallel()

		ecgp, err := NewEpochChangeGracePeriod(nil)
		require.Nil(t, ecgp)
		require.Equal(t, errEmptyGracePeriodByEpochConfig, err)
	})

	t.Run("should return error for duplicate epoch configs", func(t *testing.T) {
		t.Parallel()

		configs := []config.EpochChangeGracePeriodByEpoch{
			{EnableEpoch: 0, GracePeriodInRounds: 10},
			{EnableEpoch: 1, GracePeriodInRounds: 20},
			{EnableEpoch: 1, GracePeriodInRounds: 30},
		}
		ecgp, err := NewEpochChangeGracePeriod(configs)
		require.Nil(t, ecgp)
		require.Equal(t, errDuplicatedEpochConfig, err)
	})

	t.Run("should return error for missing epoch 0 config", func(t *testing.T) {
		t.Parallel()

		configs := []config.EpochChangeGracePeriodByEpoch{
			{EnableEpoch: 1, GracePeriodInRounds: 20},
			{EnableEpoch: 2, GracePeriodInRounds: 30},
		}
		ecgp, err := NewEpochChangeGracePeriod(configs)
		require.Nil(t, ecgp)
		require.Equal(t, errMissingEpochZeroConfig, err)
	})

	t.Run("should create epochChangeGracePeriod successfully", func(t *testing.T) {
		t.Parallel()

		configs := []config.EpochChangeGracePeriodByEpoch{
			{EnableEpoch: 0, GracePeriodInRounds: 10},
			{EnableEpoch: 2, GracePeriodInRounds: 30},
			{EnableEpoch: 1, GracePeriodInRounds: 20},
		}
		ecgp, err := NewEpochChangeGracePeriod(configs)
		require.NotNil(t, ecgp)
		require.NoError(t, err)
		require.Equal(t, uint32(10), ecgp.orderedConfigByEpoch[0].GracePeriodInRounds)
		require.Equal(t, uint32(20), ecgp.orderedConfigByEpoch[1].GracePeriodInRounds)
		require.Equal(t, uint32(30), ecgp.orderedConfigByEpoch[2].GracePeriodInRounds)
	})
}

func TestGetGracePeriodForEpoch(t *testing.T) {
	t.Parallel()

	configs := []config.EpochChangeGracePeriodByEpoch{
		{EnableEpoch: 0, GracePeriodInRounds: 10},
		{EnableEpoch: 2, GracePeriodInRounds: 30},
		{EnableEpoch: 5, GracePeriodInRounds: 50},
	}
	ecgp, err := NewEpochChangeGracePeriod(configs)
	require.NotNil(t, ecgp)
	require.NoError(t, err)

	t.Run("should return correct grace period for matching epoch", func(t *testing.T) {
		t.Parallel()

		gracePeriod, err := ecgp.GetGracePeriodForEpoch(2)
		require.NoError(t, err)
		require.Equal(t, uint32(30), gracePeriod)
	})

	t.Run("should return grace period for closest lower epoch", func(t *testing.T) {
		t.Parallel()

		gracePeriod, err := ecgp.GetGracePeriodForEpoch(4)
		require.NoError(t, err)
		require.Equal(t, uint32(30), gracePeriod)
	})

	t.Run("should return grace period for higher epochs than configured", func(t *testing.T) {
		t.Parallel()

		gracePeriod, err := ecgp.GetGracePeriodForEpoch(10)
		require.NoError(t, err)
		require.Equal(t, uint32(50), gracePeriod)
	})

	t.Run("should return error for empty config", func(t *testing.T) {
		t.Parallel()

		cfg := []config.EpochChangeGracePeriodByEpoch{
			{EnableEpoch: 0, GracePeriodInRounds: 10},
		}
		emptyECGP, _ := NewEpochChangeGracePeriod(cfg)
		// force the config to be empty, to simulate the error case
		emptyECGP.orderedConfigByEpoch = make([]config.EpochChangeGracePeriodByEpoch, 0)
		gracePeriod, err := emptyECGP.GetGracePeriodForEpoch(0)
		require.Equal(t, uint32(0), gracePeriod)
		require.Equal(t, errEmptyGracePeriodByEpochConfig, err)
	})
}

func TestIsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ecgp *epochChangeGracePeriod
	require.True(t, ecgp.IsInterfaceNil())

	ecgp = &epochChangeGracePeriod{}
	require.False(t, ecgp.IsInterfaceNil())
}
