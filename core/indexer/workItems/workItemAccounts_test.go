package workItems_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestItemAccounts_Save(t *testing.T) {
	called := false
	itemAccounts := workItems.NewItemAccounts(
		&testscommon.ElasticProcessorStub{
			SaveAccountsCalled: func(_ []*types.AccountEGLD) error {
				called = true
				return nil
			},
		},
		[]state.UserAccountHandler{},
	)
	require.False(t, itemAccounts.IsInterfaceNil())

	err := itemAccounts.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemAccounts_SaveAccountsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemAccounts := workItems.NewItemAccounts(
		&testscommon.ElasticProcessorStub{
			SaveAccountsCalled: func(_ []*types.AccountEGLD) error {
				return localErr
			},
		},
		[]state.UserAccountHandler{},
	)
	require.False(t, itemAccounts.IsInterfaceNil())

	err := itemAccounts.Save()
	require.Equal(t, localErr, err)
}
