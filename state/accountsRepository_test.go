package state_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
	"github.com/ElrondNetwork/elrond-go/state"
	mockState "github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

var testApiOptOnFinal = api.AccountQueryOptions{OnFinalBlock: true}
var testApiOptOnStartOfEpoch = api.AccountQueryOptions{OnStartOfEpoch: 23}
var testApiOptOnCurrent = api.AccountQueryOptions{}

func createMockArgsAccountsRepository() state.ArgsAccountsRepository {
	return state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      &mockState.AccountsStub{},
		CurrentStateAccountsWrapper:    &mockState.AccountsStub{},
		HistoricalStateAccountsWrapper: &mockState.AccountsStub{},
	}
}

func TestNewAccountsRepository(t *testing.T) {
	t.Parallel()

	t.Run("nil final accounts adapter", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		args.FinalStateAccountsWrapper = nil
		repository, err := state.NewAccountsRepository(args)

		assert.True(t, errors.Is(err, state.ErrNilAccountsAdapter))
		assert.True(t, strings.Contains(err.Error(), "FinalStateAccountsWrapper"))
		assert.True(t, check.IfNil(repository))
	})
	t.Run("nil current accounts adapter", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		args.CurrentStateAccountsWrapper = nil
		repository, err := state.NewAccountsRepository(args)

		assert.True(t, errors.Is(err, state.ErrNilAccountsAdapter))
		assert.True(t, strings.Contains(err.Error(), "CurrentStateAccountsWrapper"))
		assert.True(t, check.IfNil(repository))
	})
	t.Run("nil historical accounts adapter", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		args.HistoricalStateAccountsWrapper = nil
		repository, err := state.NewAccountsRepository(args)

		assert.True(t, errors.Is(err, state.ErrNilAccountsAdapter))
		assert.True(t, strings.Contains(err.Error(), "HistoricalStateAccountsWrapper"))
		assert.True(t, check.IfNil(repository))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		repository, err := state.NewAccountsRepository(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(repository))
	})
}

func TestAccountsRepository_GetAccountWithBlockInfo(t *testing.T) {
	t.Parallel()

	commonAccount := &mockState.UserAccountStub{}
	commonBlockInfo := holders.NewBlockInfo([]byte("hash"), 111, []byte("root hash"))
	commonAddress := []byte("address")

	t.Run("on final state should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		args.FinalStateAccountsWrapper = &mockState.AccountsStub{
			GetAccountWithBlockInfoCalled: func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return commonAccount, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		account, bi, err := repository.GetAccountWithBlockInfo(commonAddress, testApiOptOnFinal)
		assert.Equal(t, commonAccount, account)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
	})
	t.Run("on current state should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		args.CurrentStateAccountsWrapper = &mockState.AccountsStub{
			GetAccountWithBlockInfoCalled: func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return commonAccount, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		account, bi, err := repository.GetAccountWithBlockInfo(commonAddress, testApiOptOnCurrent)
		assert.Equal(t, commonAccount, account)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
	})
	t.Run("on start of epoch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		repository, _ := state.NewAccountsRepository(args)

		account, bi, err := repository.GetAccountWithBlockInfo(commonAddress, testApiOptOnStartOfEpoch)
		assert.True(t, check.IfNil(account))
		assert.True(t, check.IfNil(bi))
		assert.Equal(t, state.ErrFunctionalityNotImplemented, err)
	})
}

func TestAccountsRepository_GetCodeWithBlockInfo(t *testing.T) {
	t.Parallel()

	commonCode := []byte("code")
	commonBlockInfo := holders.NewBlockInfo([]byte("hash"), 111, []byte("root hash"))
	commonAddress := []byte("address")

	t.Run("on final state should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		args.FinalStateAccountsWrapper = &mockState.AccountsStub{
			GetCodeWithBlockInfoCalled: func(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error) {
				return commonCode, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		code, bi, err := repository.GetCodeWithBlockInfo(commonAddress, testApiOptOnFinal)
		assert.Equal(t, commonCode, code)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
	})
	t.Run("on current state should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		args.CurrentStateAccountsWrapper = &mockState.AccountsStub{
			GetCodeWithBlockInfoCalled: func(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error) {
				return commonCode, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		code, bi, err := repository.GetCodeWithBlockInfo(commonAddress, testApiOptOnCurrent)
		assert.Equal(t, commonCode, code)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
	})
	t.Run("on start of epoch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsAccountsRepository()
		repository, _ := state.NewAccountsRepository(args)

		code, bi, err := repository.GetCodeWithBlockInfo(commonAddress, testApiOptOnStartOfEpoch)
		assert.Nil(t, code)
		assert.True(t, check.IfNil(bi))
		assert.Equal(t, state.ErrFunctionalityNotImplemented, err)
	})
}

func TestAccountsRepository_GetCurrentStateAccountsWrapper(t *testing.T) {
	t.Parallel()

	args := createMockArgsAccountsRepository()
	repository, _ := state.NewAccountsRepository(args)

	assert.True(t, args.CurrentStateAccountsWrapper == repository.GetCurrentStateAccountsWrapper()) // pointer testing
}

func TestAccountsRepository_Close(t *testing.T) {
	t.Parallel()

	currentCloseCalled := false
	finalCloseCalled := false
	errCurrent := errors.New("err current")
	errFinal := errors.New("err final")

	args := createMockArgsAccountsRepository()
	args.CurrentStateAccountsWrapper = &mockState.AccountsStub{
		CloseCalled: func() error {
			currentCloseCalled = true
			return errCurrent
		},
	}
	args.FinalStateAccountsWrapper = &mockState.AccountsStub{
		CloseCalled: func() error {
			finalCloseCalled = true
			return errFinal
		},
	}

	t.Run("current and final close errors", func(t *testing.T) {

		repository, _ := state.NewAccountsRepository(args)

		err := repository.Close()

		assert.True(t, currentCloseCalled)
		assert.True(t, finalCloseCalled)
		assert.Equal(t, errFinal, err)
	})
	t.Run("current close errors", func(t *testing.T) {
		args.FinalStateAccountsWrapper = &mockState.AccountsStub{
			CloseCalled: func() error {
				finalCloseCalled = true
				return nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		err := repository.Close()

		assert.True(t, currentCloseCalled)
		assert.True(t, finalCloseCalled)
		assert.Equal(t, errCurrent, err)
	})

}
