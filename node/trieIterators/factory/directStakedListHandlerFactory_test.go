package factory

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDirectStakedListHandlerHandler_Disabled(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: 0,
	}

	directStakedListHandler, err := CreateDirectStakedListHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*disabled.directStakedListProcessor", fmt.Sprintf("%T", directStakedListHandler))
}

func TestCreateDirectStakedListHandlerHandler_DirectStakedListProcessor(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: core.MetachainShardId,
		Accounts: &trieIterators.AccountsWrapper{
			Mutex:           &sync.Mutex{},
			AccountsAdapter: &testscommon.AccountsStub{},
		},
		PublicKeyConverter: &mock.PubkeyConverterMock{},
		BlockChain:         &mock.BlockChainMock{},
		QueryService:       &mock.SCQueryServiceStub{},
	}

	directStakedListHandler, err := CreateDirectStakedListHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*trieIterators.directStakedListProcessor", fmt.Sprintf("%T", directStakedListHandler))
}
