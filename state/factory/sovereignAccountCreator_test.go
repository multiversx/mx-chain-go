package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/require"
)

func createArgsSovAccountCreator() ArgsSovereignAccountCreator {
	return ArgsSovereignAccountCreator{
		ArgsAccountCreator: ArgsAccountCreator{
			Hasher:              &hashingMocks.HasherMock{},
			Marshaller:          &marshallerMock.MarshalizerMock{},
			EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		},
		BaseTokenID: "WEGLD",
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

	args := createArgsSovAccountCreator()
	sovAccCreator, _ := NewSovereignAccountCreator(args)

	sovAcc, err := sovAccCreator.CreateAccount([]byte("address"))
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("%T", sovAcc), "*accounts.sovereignAccount")
}
