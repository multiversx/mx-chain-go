package factory_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
)

func getDefaultArgs() factory.ArgsAccountCreator {
	return factory.ArgsAccountCreator{
		Hasher:              &hashingMocks.HasherMock{},
		Marshaller:          &marshallerMock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
}

func TestNewAccountCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.Hasher = nil
		accF, err := factory.NewAccountCreator(args)
		assert.True(t, check.IfNil(accF))
		assert.Equal(t, errors.ErrNilHasher, err)
	})
	t.Run("nil marshalizer", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.Marshaller = nil
		accF, err := factory.NewAccountCreator(args)
		assert.True(t, check.IfNil(accF))
		assert.Equal(t, errors.ErrNilMarshalizer, err)
	})
	t.Run("nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.EnableEpochsHandler = nil
		accF, err := factory.NewAccountCreator(args)
		assert.True(t, check.IfNil(accF))
		assert.Equal(t, errors.ErrNilEnableEpochsHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		accF, err := factory.NewAccountCreator(getDefaultArgs())
		assert.False(t, check.IfNil(accF))
		assert.Nil(t, err)
	})
}

func TestAccountCreator_CreateAccountNilAddress(t *testing.T) {
	t.Parallel()

	accF, _ := factory.NewAccountCreator(getDefaultArgs())

	_, ok := accF.(*factory.AccountCreator)
	assert.Equal(t, true, ok)
	assert.False(t, check.IfNil(accF))

	acc, err := accF.CreateAccount(nil)

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAddress)
}

func TestAccountCreator_CreateAccountOk(t *testing.T) {
	t.Parallel()

	accF, _ := factory.NewAccountCreator(getDefaultArgs())

	_, ok := accF.(*factory.AccountCreator)
	assert.Equal(t, true, ok)

	acc, err := accF.CreateAccount(make([]byte, 32))

	assert.Nil(t, err)
	assert.False(t, check.IfNil(acc))
}
