package workItems_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/outport/drivers/elastic/workItems"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/stretchr/testify/require"
)

func TestItemValidators_Save(t *testing.T) {
	called := false
	validators := map[uint32][][]byte{
		0: {[]byte("val1"), []byte("val2")},
	}
	itemValidators := workItems.NewItemValidators(
		&mock.ElasticProcessorStub{
			SaveShardValidatorsPubKeysCalled: func(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
				called = true
				return nil
			},
		},
		1,
		validators,
	)
	require.False(t, itemValidators.IsInterfaceNil())

	err := itemValidators.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemValidators_SaveValidatorsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	validators := map[uint32][][]byte{
		0: {[]byte("val1"), []byte("val2")},
	}
	itemValidators := workItems.NewItemValidators(
		&mock.ElasticProcessorStub{
			SaveShardValidatorsPubKeysCalled: func(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
				return localErr
			},
		},
		1,
		validators,
	)
	require.False(t, itemValidators.IsInterfaceNil())

	err := itemValidators.Save()
	require.Equal(t, localErr, err)
}
