package syncer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO add more tests

func TestValidatorAccountsSyncer_SyncAccounts(t *testing.T) {
	t.Parallel()

	args := ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
	}

	syncer, err := NewValidatorAccountsSyncer(args)
	assert.Nil(t, err)
	assert.NotNil(t, syncer)

	err = syncer.SyncAccounts([]byte("rootHash"), nil)
	assert.Equal(t, ErrNilStorageMarker, err)
}
