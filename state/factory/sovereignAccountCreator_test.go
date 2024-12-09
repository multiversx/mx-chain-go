package factory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
)

func createArgsSovAccountCreator() ArgsSovereignAccountCreator {
	return ArgsSovereignAccountCreator{
		ArgsAccountCreator: ArgsAccountCreator{
			Hasher:              &hashingMocks.HasherMock{},
			Marshaller:          &marshallerMock.MarshalizerMock{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		},
		BaseTokenID: "WEGLD-bd4d79",
	}
}

func TestNewSovereignAccountCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createArgsSovAccountCreator()
		args.Hasher = nil
		sovAccCreator, err := NewSovereignAccountCreator(args)
		require.Nil(t, sovAccCreator)
		require.Error(t, errors.ErrNilHasher, err)
	})
	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createArgsSovAccountCreator()
		args.Marshaller = nil
		sovAccCreator, err := NewSovereignAccountCreator(args)
		require.Nil(t, sovAccCreator)
		require.Error(t, errors.ErrNilMarshalizer, err)
	})
	t.Run("nil enable epochs handler, should return error", func(t *testing.T) {
		args := createArgsSovAccountCreator()
		args.EnableEpochsHandler = nil
		sovAccCreator, err := NewSovereignAccountCreator(args)
		require.Nil(t, sovAccCreator)
		require.Error(t, errors.ErrNilEnableEpochsHandler, err)
	})
	t.Run("invalid base esdt token, should return error", func(t *testing.T) {
		args := createArgsSovAccountCreator()
		args.BaseTokenID = ""
		sovAccCreator, err := NewSovereignAccountCreator(args)
		require.Nil(t, sovAccCreator)
		require.Error(t, errors.ErrEmptyBaseToken, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createArgsSovAccountCreator()
		sovAccCreator, err := NewSovereignAccountCreator(args)
		require.Nil(t, err)
		require.False(t, sovAccCreator.IsInterfaceNil())
	})
}

func TestSovereignAccountCreator_CreateAccount(t *testing.T) {
	t.Parallel()

	testCreateAccount(t, "WEGLD-bd4d79")
	testCreateAccount(t, "sov-ESDT-a1b2c3")
}

func testCreateAccount(t *testing.T, baseTokenID string) {
	args := createArgsSovAccountCreator()
	args.BaseTokenID = baseTokenID
	sovAccCreator, _ := NewSovereignAccountCreator(args)

	sovAcc, err := sovAccCreator.CreateAccount([]byte("address"))
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%T", sovAcc), "*accounts.sovereignAccount")
}
