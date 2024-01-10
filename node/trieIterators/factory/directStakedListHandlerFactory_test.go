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
	"github.com/stretchr/testify/require"
)

func TestNewDirectStakedListProcessorFactory(t *testing.T) {
	t.Parallel()

	directStakedListHandlerFactory := trieIteratorsFactory.NewDirectStakedListProcessorFactory()
	require.False(t, directStakedListHandlerFactory.IsInterfaceNil())
}

func TestDirectStakedListProcessorFactory_CreateDirectStakedListProcessor_Disabled(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: 0,
	}

	directStakedListHandler, err := trieIteratorsFactory.NewDirectStakedListProcessorFactory().CreateDirectStakedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*disabled.directStakedListProcessor", fmt.Sprintf("%T", directStakedListHandler))
}

func TestDirectStakedListProcessorFactory_CreateDirectStakedListProcessor_DirectStakedListProcessorFactory(t *testing.T) {
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

	directStakedListHandler, err := trieIteratorsFactory.NewDirectStakedListProcessorFactory().CreateDirectStakedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*trieIterators.directStakedListProcessor", fmt.Sprintf("%T", directStakedListHandler))
}
