package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNewAccountFactoryCreator_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	accF, err := factory.NewAccountFactoryCreator(nil)

	assert.Equal(t, err, state.ErrNilShardCoordinator)
	assert.Nil(t, accF)
}

func TestNewAccountFactoryCreator_NormalAccount(t *testing.T) {
	t.Parallel()

	shardC := mock.ShardCoordinatorMock{
		SelfID:     0,
		NrOfShards: 1,
	}
	accF, err := factory.NewAccountFactoryCreator(shardC)
	assert.Nil(t, err)

	accWrp, err := accF.CreateAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	_, ok := accWrp.(*state.Account)
	assert.Equal(t, true, ok)

	assert.Nil(t, err)
	assert.NotNil(t, accF)
}

func TestNewAccountFactoryCreator_MetaAccount(t *testing.T) {
	t.Parallel()

	shardC := mock.ShardCoordinatorMock{
		SelfID:     sharding.MetachainShardId,
		NrOfShards: 1,
	}
	accF, err := factory.NewAccountFactoryCreator(shardC)
	assert.Nil(t, err)

	accWrp, err := accF.CreateAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	_, ok := accWrp.(*state.MetaAccount)
	assert.Equal(t, true, ok)

	assert.Nil(t, err)
	assert.NotNil(t, accF)
}

func TestNewAccountFactoryCreator_BadShardID(t *testing.T) {
	t.Parallel()

	shardC := mock.ShardCoordinatorMock{
		SelfID:     10,
		NrOfShards: 5,
	}
	accF, err := factory.NewAccountFactoryCreator(shardC)
	assert.Nil(t, accF)
	assert.Equal(t, state.ErrUnknownShardId, err)
}
