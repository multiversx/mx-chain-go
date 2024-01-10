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
	"github.com/stretchr/testify/require"
)

func TestNewSovereignDirectStakedListProcessorFactory(t *testing.T) {
	t.Parallel()

	directStakedListHandlerFactory := trieIteratorsFactory.NewSovereignDirectStakedListProcessorFactory()
	require.False(t, directStakedListHandlerFactory.IsInterfaceNil())
}

func TestSovereignDirectStakedListHandlerFactory_CreateDirectStakedListProcessorHandler_DirectStakedListHandlerFactory(t *testing.T) {
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

	sovereignDirectStakedListHandler, err := trieIteratorsFactory.NewSovereignDirectStakedListProcessorFactory().CreateDirectStakedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*trieIterators.directStakedListProcessor", fmt.Sprintf("%T", sovereignDirectStakedListHandler))
}
