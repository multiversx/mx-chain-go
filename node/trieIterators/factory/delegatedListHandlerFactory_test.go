package factory

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDelegatedListHandlerHandler_Disabled(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: 0,
	}

	delegatedListHandler, err := CreateDelegatedListHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*disabled.delegatedListProcessor", fmt.Sprintf("%T", delegatedListHandler))
}

func TestCreateDelegatedListHandlerHandler_DelegatedListProcessor(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: core.MetachainShardId,
		Accounts: &trieIterators.AccountsWrapper{
			Mutex:           &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{},
		},
		PublicKeyConverter: &mock.PubkeyConverterMock{},
		BlockChain:         &mock.BlockChainMock{},
		QueryService:       &mock.SCQueryServiceStub{},
	}

	delegatedListHandler, err := CreateDelegatedListHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*trieIterators.delegatedListProcessor", fmt.Sprintf("%T", delegatedListHandler))
}
