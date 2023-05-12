package syncer_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidatorAccountsSyncer(t *testing.T) {
	t.Parallel()

	t.Run("invalid base args (nil hasher) should fail", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		args.Hasher = nil

		syncer, err := syncer.NewValidatorAccountsSyncer(args)
		assert.Nil(t, syncer)
		assert.Equal(t, state.ErrNilHasher, err)
	})

	t.Run("invalid timeout, should fail", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		args.Timeout = 0

		s, err := syncer.NewValidatorAccountsSyncer(args)
		assert.Nil(t, s)
		assert.True(t, errors.Is(err, common.ErrInvalidTimeout))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		v, err := syncer.NewValidatorAccountsSyncer(args)
		require.Nil(t, err)
		require.NotNil(t, v)
	})

}
