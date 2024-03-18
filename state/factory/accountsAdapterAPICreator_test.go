package factory

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	mockState "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	mockTrie "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func createMockAccountsArgs() state.ArgsAccountsDB {
	return state.ArgsAccountsDB{
		Trie: &mockTrie.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &storageManager.StorageManagerStub{}
			},
		},
		Hasher:                &testscommon.HasherStub{},
		Marshaller:            &marshallerMock.MarshalizerMock{},
		AccountFactory:        &mockState.AccountsFactoryStub{},
		StoragePruningManager: &mockState.StoragePruningManagerStub{},
		AddressConverter:      &testscommon.PubkeyConverterMock{},
		SnapshotsManager:      &mockState.SnapshotsManagerStub{},
		StateChangesCollector: state.NewStateChangesCollector(),
	}
}

func TestCreateAccountsAdapterAPIOnFinal(t *testing.T) {
	t.Parallel()

	t.Run("error while creating provider", func(t *testing.T) {
		t.Parallel()

		accountsAdapterApi, err := CreateAccountsAdapterAPIOnFinal(createMockAccountsArgs(), nil)

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "CreateAccountsAdapterAPIOnFinal"))
		assert.True(t, check.IfNil(accountsAdapterApi))
	})
	t.Run("error while creating accounts adapter", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsArgs()
		args.AccountFactory = nil
		accountsAdapterApi, err := CreateAccountsAdapterAPIOnFinal(args, &testscommon.ChainHandlerStub{})

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "CreateAccountsAdapterAPIOnFinal"))
		assert.True(t, check.IfNil(accountsAdapterApi))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsArgs()
		accountsAdapterApi, err := CreateAccountsAdapterAPIOnFinal(args, &testscommon.ChainHandlerStub{})

		assert.Nil(t, err)
		assert.False(t, check.IfNil(accountsAdapterApi))
	})
}

func TestCreateAccountsAdapterAPIOnCurrent(t *testing.T) {
	t.Parallel()

	t.Run("error while creating provider", func(t *testing.T) {
		t.Parallel()

		accountsAdapterApi, err := CreateAccountsAdapterAPIOnCurrent(createMockAccountsArgs(), nil)

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "CreateAccountsAdapterAPIOnCurrent"))
		assert.True(t, check.IfNil(accountsAdapterApi))
	})
	t.Run("error while creating accounts adapter", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsArgs()
		args.AccountFactory = nil
		accountsAdapterApi, err := CreateAccountsAdapterAPIOnCurrent(args, &testscommon.ChainHandlerStub{})

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "CreateAccountsAdapterAPIOnCurrent"))
		assert.True(t, check.IfNil(accountsAdapterApi))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsArgs()
		accountsAdapterApi, err := CreateAccountsAdapterAPIOnCurrent(args, &testscommon.ChainHandlerStub{})

		assert.Nil(t, err)
		assert.False(t, check.IfNil(accountsAdapterApi))
	})
}

func TestCreateAccountsAdapterAPIOnHistorical(t *testing.T) {
	t.Parallel()

	t.Run("error while creating accounts adapter", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsArgs()
		args.AccountFactory = nil
		accountsAdapterApi, err := CreateAccountsAdapterAPIOnHistorical(args)

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "CreateAccountsAdapterAPIOnHistorical"))
		assert.True(t, check.IfNil(accountsAdapterApi))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockAccountsArgs()
		accountsAdapterApi, err := CreateAccountsAdapterAPIOnHistorical(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(accountsAdapterApi))
	})
}
