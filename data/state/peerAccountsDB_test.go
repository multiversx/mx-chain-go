package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerAccountsDB_WithNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		nil,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilTrie, err)
}

func TestNewPeerAccountsDB_WithNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilHasher, err)
}

func TestNewPeerAccountsDB_WithNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{},
		&mock.HasherMock{},
		nil,
		&mock.AccountsFactoryStub{},
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilMarshalizer, err)
}

func TestNewPeerAccountsDB_WithNilAddressFactoryShouldErr(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
	)

	assert.True(t, check.IfNil(adb))
	assert.Equal(t, state.ErrNilAccountFactory, err)
}

func TestNewPeerAccountsDB_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	adb, err := state.NewPeerAccountsDB(
		&testscommon.TrieStub{
			GetStorageManagerCalled: func() data.StorageManager {
				return &mock.StorageManagerStub{
					DatabaseCalled: func() data.DBWriteCacher {
						return mock.NewMemDbMock()
					},
				}
			},
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsFactoryStub{},
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(adb))
}
