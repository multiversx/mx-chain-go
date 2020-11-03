package workItems_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/require"
)

func TestItemRating_Save(t *testing.T) {
	id := "0_1"
	called := false
	itemRating := workItems.NewItemRating(
		&mock.ElasticProcessorStub{
			SaveValidatorsRatingCalled: func(index string, validatorsRatingInfo []workItems.ValidatorRatingInfo) error {
				require.Equal(t, id, index)
				called = true
				return nil
			},
		},
		id,
		[]workItems.ValidatorRatingInfo{
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
		&mock.ElasticProcessorStub{
			SaveValidatorsRatingCalled: func(index string, validatorsRatingInfo []workItems.ValidatorRatingInfo) error {
				return localErr
			},
		},
		id,
		[]workItems.ValidatorRatingInfo{
			{PublicKey: "pub-key", Rating: 100},
		},
	)
	require.False(t, itemRating.IsInterfaceNil())

	err := itemRating.Save()
	require.Equal(t, localErr, err)
}
