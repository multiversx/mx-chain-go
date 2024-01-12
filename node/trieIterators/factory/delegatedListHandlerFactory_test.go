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

func createMockArgs(shardId uint32) trieIterators.ArgTrieIteratorProcessor {
	return trieIterators.ArgTrieIteratorProcessor{
		ShardID: shardId,
		Accounts: &trieIterators.AccountsWrapper{
			Mutex:           &sync.Mutex{},
			AccountsAdapter: &stateMock.AccountsStub{},
		},
		PublicKeyConverter: &testscommon.PubkeyConverterMock{},
		QueryService:       &mock.SCQueryServiceStub{},
	}
}

func TestNewDelegatedListProcessorFactory(t *testing.T) {
	t.Parallel()

	delegatedListHandlerFactory := trieIteratorsFactory.NewDelegatedListProcessorFactory()
	require.False(t, delegatedListHandlerFactory.IsInterfaceNil())
}

func TestDelegatedListProcessorFactory_CreateDelegatedListProcessorHandlerDisabledProcessor(t *testing.T) {
	t.Parallel()

	args := createMockArgs(0)

	delegatedListHandler, err := trieIteratorsFactory.NewDelegatedListProcessorFactory().CreateDelegatedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*disabled.delegatedListProcessor", fmt.Sprintf("%T", delegatedListHandler))
}

func TestDelegatedListProcessorFactory_CreateDelegatedListProcessorHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgs(core.MetachainShardId)

	delegatedListHandler, err := trieIteratorsFactory.NewDelegatedListProcessorFactory().CreateDelegatedListProcessorHandler(args)
	require.Nil(t, err)
	require.Equal(t, "*trieIterators.delegatedListProcessor", fmt.Sprintf("%T", delegatedListHandler))
}
