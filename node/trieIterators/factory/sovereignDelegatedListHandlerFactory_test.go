package factory_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignDelegatedListHandlerFactory(t *testing.T) {
	t.Parallel()

	sovereignDelegatedListHandlerFactory := trieIteratorsFactory.NewSovereignDelegatedListHandlerFactory()
	require.False(t, sovereignDelegatedListHandlerFactory.IsInterfaceNil())
}

func TestSovereignDelegatedListHandlerFactory_CreateDelegatedListHandler_DelegatedListProcessor(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: core.MetachainShardId,
		Accounts: &trieIterators.AccountsWrapper{
			Mutex:           &sync.Mutex{},
			AccountsAdapter: &stateMock.AccountsStub{},
		},
		PublicKeyConverter: &testscommon.PubkeyConverterMock{},
		QueryService:       &mock.SCQueryServiceStub{},
	}

	sovereignDelegatedListHandler, err := trieIteratorsFactory.NewSovereignDelegatedListHandlerFactory().CreateDelegatedListHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*trieIterators.delegatedListProcessor", fmt.Sprintf("%T", sovereignDelegatedListHandler))
}
