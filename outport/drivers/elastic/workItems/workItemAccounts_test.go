package workItems_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport/drivers/elastic/workItems"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/stretchr/testify/require"
)

func TestItemAccounts_Save(t *testing.T) {
	called := false
	itemAccounts := workItems.NewItemAccounts(
		&mock.ElasticProcessorStub{
			SaveAccountsCalled: func(_ []state.UserAccountHandler) error {
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
		&mock.ElasticProcessorStub{
			SaveAccountsCalled: func(_ []state.UserAccountHandler) error {
				return localErr
			},
		},
		[]state.UserAccountHandler{},
	)
	require.False(t, itemAccounts.IsInterfaceNil())

	err := itemAccounts.Save()
	require.Equal(t, localErr, err)
}
