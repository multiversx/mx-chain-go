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

func TestCreateTotalStakedValueHandler_DisabledTotalStakedValueProcessor(t *testing.T) {
	t.Parallel()

	args := &ArgStakeProcessors{
		ShardID: 0,
	}

	totalStakedValueHandler, err := CreateTotalStakedValueHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*disabled.stakeValuesProcessor", fmt.Sprintf("%T", totalStakedValueHandler))
}

func TestCreateTotalStakedValueHandler_TotalStakedValueProcessor(t *testing.T) {
	t.Parallel()

	args := &ArgStakeProcessors{
		ShardID: core.MetachainShardId,
		Accounts: &trieIterators.AccountsWrapper{
			Mutex:           &sync.Mutex{},
			AccountsAdapter: &mock.AccountsStub{},
		},
		PublicKeyConverter: &mock.PubkeyConverterMock{},
		BlockChain:         &mock.BlockChainMock{},
		QueryService:       &mock.SCQueryServiceStub{},
	}

	totalStakedValueHandler, err := CreateTotalStakedValueHandler(args)
	require.Nil(t, err)
	assert.Equal(t, "*trieIterators.stakedValuesProc", fmt.Sprintf("%T", totalStakedValueHandler))
}
