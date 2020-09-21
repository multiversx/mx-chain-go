package workItems_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/require"
)

func TestItemRounds_Save(t *testing.T) {
	called := false
	itemRounds := workItems.NewItemRounds(
		&mock.ElasticProcessorStub{
			SaveRoundsInfoCalled: func(infos []workItems.RoundInfo) error {
				called = true
				return nil
			},
		},
		[]workItems.RoundInfo{
			{},
		},
	)
	require.False(t, itemRounds.IsInterfaceNil())

	err := itemRounds.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemRounds_SaveRoundsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemRounds := workItems.NewItemRounds(
		&mock.ElasticProcessorStub{
			SaveRoundsInfoCalled: func(infos []workItems.RoundInfo) error {
				return localErr
			},
		},
		[]workItems.RoundInfo{
			{},
		},
	)
	require.False(t, itemRounds.IsInterfaceNil())

	err := itemRounds.Save()
	require.Equal(t, localErr, err)
}
