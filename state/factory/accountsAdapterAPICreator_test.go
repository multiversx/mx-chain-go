package factory

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	mockState "github.com/ElrondNetwork/elrond-go/testscommon/state"
	mockTrie "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func createMockAccountsArgs() state.ArgsAccountsDB {
	return state.ArgsAccountsDB{
		Trie: &mockTrie.TrieStub{
			GetStorageManagerCalled: func() common.StorageManager {
				return &testscommon.StorageManagerStub{}
			},
		},
		Hasher:                &testscommon.HasherStub{},
		Marshaller:            &testscommon.MarshalizerMock{},
		AccountFactory:        &mockState.AccountsFactoryStub{},
		StoragePruningManager: &mockState.StoragePruningManagerStub{},
		ProcessingMode:        0,
		ProcessStatusHandler:  &testscommon.ProcessStatusHandlerStub{},
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
