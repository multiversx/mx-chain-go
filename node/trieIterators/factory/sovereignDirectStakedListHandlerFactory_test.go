package factory_test

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/mock"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignDirectStakedListHandlerFactory(t *testing.T) {
	t.Parallel()

	directStakedListHandlerFactory := trieIteratorsFactory.NewDirectStakedListHandlerFactory()
	require.False(t, directStakedListHandlerFactory.IsInterfaceNil())
}

func TestSovereignDirectStakedListHandlerFactory_CreateDirectStakedListHandler_DirectStakedListProcessor(t *testing.T) {
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

	sovereignDirectStakedListHandler, err := trieIteratorsFactory.NewSovereignDirectStakedListHandlerFactory().CreateDirectStakedListHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*trieIterators.directStakedListProcessor", fmt.Sprintf("%T", sovereignDirectStakedListHandler))
}
