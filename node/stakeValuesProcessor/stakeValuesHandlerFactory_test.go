package stakeValuesProcessor

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTotalStakedValueHandler_DisabledTotalStakedValueProcessor(t *testing.T) {
	t.Parallel()

	args := &ArgsTotalStakedValueHandler{
		ShardID: 0,
	}

	totalStakedValueHandler, err := CreateTotalStakedValueHandler(args)
	require.Nil(t, err)

	_, ok := totalStakedValueHandler.(*stakedValuesProc)
	require.False(t, ok)
}

func TestCreateTotalStakedValueHandler_TotalStakedValueProcessor(t *testing.T) {
	t.Parallel()

	args := &ArgsTotalStakedValueHandler{
		ShardID:            core.MetachainShardId,
		Accounts:           &mock.AccountsStub{},
		PublicKeyConverter: &mock.PubkeyConverterMock{},
		BlockChain:         &mock.BlockChainMock{},
		QueryService:       &mock.SCQueryServiceStub{},
	}

	totalStakedValueHandler, err := CreateTotalStakedValueHandler(args)
	require.Nil(t, err)

	totalStakedValueProc, ok := totalStakedValueHandler.(*stakedValuesProc)
	require.True(t, ok)
	require.NotNil(t, totalStakedValueProc)
}
