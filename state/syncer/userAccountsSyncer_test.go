package syncer_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func getDefaultUserAccountsSyncerArgs() syncer.ArgsNewUserAccountsSyncer {
	return syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		ShardId:                   1,
		Throttler:                 &mock.ThrottlerStub{},
		AddressPubKeyConverter:    &testscommon.PubkeyConverterStub{},
	}
}

func TestNewUserAccountsSyncer(t *testing.T) {
	t.Parallel()

	t.Run("invalid base args (nil hasher) should fail", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Hasher = nil

		syncer, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, syncer)
		assert.Equal(t, state.ErrNilHasher, err)
	})

	t.Run("nil throttler", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Throttler = nil

		syncer, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, syncer)
		assert.Equal(t, data.ErrNilThrottler, err)
	})

	t.Run("nil address pubkey converter", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.AddressPubKeyConverter = nil

		s, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, s)
		assert.Equal(t, syncer.ErrNilPubkeyConverter, err)
	})

	t.Run("invalid timeout, should fail", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Timeout = 0

		s, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, s)
		assert.True(t, errors.Is(err, common.ErrInvalidTimeout))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		syncer, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, err)
		assert.NotNil(t, syncer)
	})
}
