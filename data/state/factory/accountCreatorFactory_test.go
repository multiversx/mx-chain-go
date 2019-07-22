package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewAccountFactoryCreator_BadType(t *testing.T) {
	t.Parallel()

	accF, err := factory.NewAccountFactoryCreator(factory.InvalidType)

	assert.Equal(t, err, state.ErrUnknownAccountType)
	assert.Nil(t, accF)
}

func TestNewAccountFactoryCreator_NormalAccount(t *testing.T) {
	t.Parallel()

	accF, err := factory.NewAccountFactoryCreator(factory.UserAccount)
	assert.Nil(t, err)

	accWrp, err := accF.CreateAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	_, ok := accWrp.(*state.Account)
	assert.Equal(t, true, ok)

	assert.Nil(t, err)
	assert.NotNil(t, accF)
}

func TestNewAccountFactoryCreator_MetaAccount(t *testing.T) {
	t.Parallel()

	accF, err := factory.NewAccountFactoryCreator(factory.ShardStatistics)
	assert.Nil(t, err)

	accWrp, err := accF.CreateAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	_, ok := accWrp.(*state.MetaAccount)
	assert.Equal(t, true, ok)

	assert.Nil(t, err)
	assert.NotNil(t, accF)
}

func TestNewAccountFactoryCreator_PeerAccount(t *testing.T) {
	t.Parallel()

	accF, err := factory.NewAccountFactoryCreator(factory.ValidatorAccount)
	assert.Nil(t, err)

	accWrp, err := accF.CreateAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	_, ok := accWrp.(*state.PeerAccount)
	assert.Equal(t, true, ok)

	assert.Nil(t, err)
	assert.NotNil(t, accF)
}

func TestNewAccountFactoryCreator_UnknownType(t *testing.T) {
	t.Parallel()

	accF, err := factory.NewAccountFactoryCreator(10)
	assert.Nil(t, accF)
	assert.Equal(t, state.ErrUnknownAccountType, err)
}
