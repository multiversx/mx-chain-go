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

func TestCreateTotalStakedValueHandler_Disabled(t *testing.T) {
	t.Parallel()

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID: 0,
	}

	totalStakedValueHandler, err := CreateTotalStakedValueHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*disabled.stakeValuesProcessor", fmt.Sprintf("%T", totalStakedValueHandler))
}

func TestCreateTotalStakedValueHandler_TotalStakedValueProcessor(t *testing.T) {
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

	totalStakedValueHandler, err := CreateTotalStakedValueHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*trieIterators.stakedValuesProcessor", fmt.Sprintf("%T", totalStakedValueHandler))
}
