package workItems_test

import (
	"errors"
	mock2 "github.com/ElrondNetwork/elrond-go/outport/mock"
	"testing"

	"github.com/ElrondNetwork/elrond-go/outport/drivers/elastic/workItems"
	"github.com/ElrondNetwork/elrond-go/outport/types"
	"github.com/stretchr/testify/require"
)

func TestItemRating_Save(t *testing.T) {
	id := "0_1"
	called := false
	itemRating := workItems.NewItemRating(
		&mock2.ElasticProcessorStub{
			SaveValidatorsRatingCalled: func(index string, validatorsRatingInfo []types.ValidatorRatingInfo) error {
				require.Equal(t, id, index)
				called = true
				return nil
			},
		},
		id,
		[]types.ValidatorRatingInfo{
			{PublicKey: "pub-key", Rating: 100},
		},
	)
	require.False(t, itemRating.IsInterfaceNil())

	err := itemRating.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemRating_SaveShouldErr(t *testing.T) {
	id := "0_1"
	localErr := errors.New("local err")
	itemRating := workItems.NewItemRating(
		&mock2.ElasticProcessorStub{
			SaveValidatorsRatingCalled: func(index string, validatorsRatingInfo []types.ValidatorRatingInfo) error {
				return localErr
			},
		},
		id,
		[]types.ValidatorRatingInfo{
			{PublicKey: "pub-key", Rating: 100},
		},
	)
	require.False(t, itemRating.IsInterfaceNil())

	err := itemRating.Save()
	require.Equal(t, localErr, err)
}
