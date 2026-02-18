package factory_test

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

func getDefaultArgs() factory.ArgsAccountCreator {
	return factory.ArgsAccountCreator{
		Hasher:                 &hashingMocks.HasherMock{},
		Marshaller:             &marshallerMock.MarshalizerMock{},
		EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		StateAccessesCollector: &stateMock.StateAccessesCollectorStub{},
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
	assert.False(t, check.IfNil(accF))

	acc, err := accF.CreateAccount(nil)

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAddress)
}

func TestAccountCreator_CreateAccountOk(t *testing.T) {
	t.Parallel()

	accF, _ := factory.NewAccountCreator(getDefaultArgs())

	acc, err := accF.CreateAccount(make([]byte, 32))
	assert.Nil(t, err)
	assert.False(t, check.IfNil(acc))
}

func TestAccountCreator_UnmarshalLightAccount(t *testing.T) {
	t.Parallel()

	t.Run("correctly extracts nonce, balance, codeMetadata, rootHash", func(t *testing.T) {
		t.Parallel()

		accF, err := factory.NewAccountCreator(getDefaultArgs())
		assert.Nil(t, err)

		lau, ok := accF.(state.LightAccountUnmarshaller)
		assert.True(t, ok, "accountCreator should implement LightAccountUnmarshaller")

		// Create a real account, marshal it, then unmarshal as light
		addr := make([]byte, 32)
		addr[0] = 0xAA

		acc, err := accF.CreateAccount(addr)
		assert.Nil(t, err)

		userAcc, ok := acc.(state.UserAccountHandler)
		assert.True(t, ok)

		userAcc.IncreaseNonce(77)
		err = userAcc.AddToBalance(big.NewInt(12345))
		assert.Nil(t, err)
		userAcc.SetCodeMetadata([]byte{0x08, 0x00}) // guarded bit

		// Marshal with the same marshaller
		marshaller := &marshallerMock.MarshalizerMock{}
		data, err := marshaller.Marshal(acc)
		assert.Nil(t, err)

		lightAcc, err := lau.UnmarshalLightAccount(addr, data)
		assert.Nil(t, err)
		assert.NotNil(t, lightAcc)
		assert.Equal(t, uint64(77), lightAcc.GetNonce())
		assert.Equal(t, 0, big.NewInt(12345).Cmp(lightAcc.GetBalance()))
		assert.Equal(t, []byte{0x08, 0x00}, lightAcc.GetCodeMetadata())
		assert.False(t, lightAcc.IsInterfaceNil())
	})

	t.Run("nil balance in marshalled data returns zero", func(t *testing.T) {
		t.Parallel()

		accF, _ := factory.NewAccountCreator(getDefaultArgs())
		lau := accF.(state.LightAccountUnmarshaller)

		// Create account with zero balance (nil in protobuf)
		addr := make([]byte, 32)
		acc, _ := accF.CreateAccount(addr)

		marshaller := &marshallerMock.MarshalizerMock{}
		data, _ := marshaller.Marshal(acc)

		lightAcc, err := lau.UnmarshalLightAccount(addr, data)
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(0), lightAcc.GetBalance())
	})

	t.Run("invalid data returns error", func(t *testing.T) {
		t.Parallel()

		accF, _ := factory.NewAccountCreator(getDefaultArgs())
		lau := accF.(state.LightAccountUnmarshaller)

		lightAcc, err := lau.UnmarshalLightAccount([]byte("addr"), []byte("invalid data"))
		assert.NotNil(t, err)
		assert.Nil(t, lightAcc)
	})
}
