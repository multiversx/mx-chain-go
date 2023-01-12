package state_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	mockState "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

var testApiOptOnFinal = api.AccountQueryOptions{OnFinalBlock: true}
var testApiOptOnStartOfEpoch = api.AccountQueryOptions{OnStartOfEpoch: core.OptionalUint32{Value: 32, HasValue: true}}
var testApiOptOnCurrent = api.AccountQueryOptions{}
var testApiOptOnHistorical = api.AccountQueryOptions{BlockRootHash: []byte("abba"), HintEpoch: core.OptionalUint32{Value: 7, HasValue: true}}

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

	var optionsPassedToGetAccountWithBlockInfo common.RootHashHolder = holders.NewRootHashHolderAsEmpty()

	commonAccount := &mockState.UserAccountStub{}
	commonBlockInfo := holders.NewBlockInfo([]byte("hash"), 111, []byte("root hash"))
	commonAddress := []byte("address")

	t.Run("on final state should work", func(t *testing.T) {
		args := createMockArgsAccountsRepository()
		args.FinalStateAccountsWrapper = &mockState.AccountsStub{
			GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
				optionsPassedToGetAccountWithBlockInfo = options
				return commonAccount, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		account, bi, err := repository.GetAccountWithBlockInfo(commonAddress, testApiOptOnFinal)
		assert.Equal(t, commonAccount, account)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
		assert.Equal(t, holders.NewRootHashHolderAsEmpty(), optionsPassedToGetAccountWithBlockInfo)
	})
	t.Run("on current state should work", func(t *testing.T) {
		args := createMockArgsAccountsRepository()
		args.CurrentStateAccountsWrapper = &mockState.AccountsStub{
			GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
				optionsPassedToGetAccountWithBlockInfo = options
				return commonAccount, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		account, bi, err := repository.GetAccountWithBlockInfo(commonAddress, testApiOptOnCurrent)
		assert.Equal(t, commonAccount, account)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
		assert.Equal(t, holders.NewRootHashHolderAsEmpty(), optionsPassedToGetAccountWithBlockInfo)
	})
	t.Run("on historical state should work", func(t *testing.T) {
		args := createMockArgsAccountsRepository()
		args.HistoricalStateAccountsWrapper = &mockState.AccountsStub{
			GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
				optionsPassedToGetAccountWithBlockInfo = options
				return commonAccount, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		account, bi, err := repository.GetAccountWithBlockInfo(commonAddress, testApiOptOnHistorical)
		assert.Equal(t, commonAccount, account)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
		assert.Equal(t, testApiOptOnHistorical.BlockRootHash, optionsPassedToGetAccountWithBlockInfo.GetRootHash())
		assert.Equal(t, testApiOptOnHistorical.HintEpoch, optionsPassedToGetAccountWithBlockInfo.GetEpoch())
	})
	t.Run("on start of epoch should error", func(t *testing.T) {
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

	var optionsPassedToGetAccountWithBlockInfo common.RootHashHolder = holders.NewRootHashHolderAsEmpty()

	commonCode := []byte("code")
	commonBlockInfo := holders.NewBlockInfo([]byte("hash"), 111, []byte("root hash"))
	commonAddress := []byte("address")

	t.Run("on final state should work", func(t *testing.T) {
		args := createMockArgsAccountsRepository()
		args.FinalStateAccountsWrapper = &mockState.AccountsStub{
			GetCodeWithBlockInfoCalled: func(codeHash []byte, options common.RootHashHolder) ([]byte, common.BlockInfo, error) {
				optionsPassedToGetAccountWithBlockInfo = options
				return commonCode, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		code, bi, err := repository.GetCodeWithBlockInfo(commonAddress, testApiOptOnFinal)
		assert.Equal(t, commonCode, code)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
		assert.Equal(t, holders.NewRootHashHolderAsEmpty(), optionsPassedToGetAccountWithBlockInfo)
	})
	t.Run("on current state should work", func(t *testing.T) {
		args := createMockArgsAccountsRepository()
		args.CurrentStateAccountsWrapper = &mockState.AccountsStub{
			GetCodeWithBlockInfoCalled: func(codeHash []byte, options common.RootHashHolder) ([]byte, common.BlockInfo, error) {
				optionsPassedToGetAccountWithBlockInfo = options
				return commonCode, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		code, bi, err := repository.GetCodeWithBlockInfo(commonAddress, testApiOptOnCurrent)
		assert.Equal(t, commonCode, code)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
		assert.Equal(t, holders.NewRootHashHolderAsEmpty(), optionsPassedToGetAccountWithBlockInfo)
	})
	t.Run("on historical state should work", func(t *testing.T) {
		args := createMockArgsAccountsRepository()
		args.HistoricalStateAccountsWrapper = &mockState.AccountsStub{
			GetCodeWithBlockInfoCalled: func(codeHash []byte, options common.RootHashHolder) ([]byte, common.BlockInfo, error) {
				optionsPassedToGetAccountWithBlockInfo = options
				return commonCode, commonBlockInfo, nil
			},
		}
		repository, _ := state.NewAccountsRepository(args)

		code, bi, err := repository.GetCodeWithBlockInfo(commonAddress, testApiOptOnHistorical)
		assert.Equal(t, commonCode, code)
		assert.Equal(t, commonBlockInfo, bi)
		assert.Nil(t, err)
		assert.Equal(t, testApiOptOnHistorical.BlockRootHash, optionsPassedToGetAccountWithBlockInfo.GetRootHash())
		assert.Equal(t, testApiOptOnHistorical.HintEpoch, optionsPassedToGetAccountWithBlockInfo.GetEpoch())
	})
	t.Run("on start of epoch should error", func(t *testing.T) {
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
	historicalCloseCalled := false
	var errCurrent error
	var errFinal error
	var errHistorical error

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
	args.HistoricalStateAccountsWrapper = &mockState.AccountsStub{
		CloseCalled: func() error {
			historicalCloseCalled = true
			return errHistorical
		},
	}

	t.Run("all have close errors, but last one is returned", func(t *testing.T) {
		currentCloseCalled = false
		finalCloseCalled = false
		historicalCloseCalled = false

		errCurrent = errors.New("err current")
		errFinal = errors.New("err final")
		errHistorical = errors.New("err historical")

		repository, _ := state.NewAccountsRepository(args)
		err := repository.Close()

		assert.True(t, currentCloseCalled)
		assert.True(t, finalCloseCalled)
		assert.True(t, historicalCloseCalled)

		assert.Equal(t, errHistorical, err)
	})

	t.Run("historical has close error", func(t *testing.T) {
		currentCloseCalled = false
		finalCloseCalled = false
		historicalCloseCalled = false

		errCurrent = nil
		errFinal = nil
		errHistorical = errors.New("err historical")

		repository, _ := state.NewAccountsRepository(args)
		err := repository.Close()

		assert.True(t, currentCloseCalled)
		assert.True(t, finalCloseCalled)
		assert.True(t, historicalCloseCalled)

		assert.Equal(t, errHistorical, err)
	})

	t.Run("current has close error", func(t *testing.T) {
		currentCloseCalled = false
		finalCloseCalled = false
		historicalCloseCalled = false

		errCurrent = errors.New("err current")
		errFinal = nil
		errHistorical = nil

		repository, _ := state.NewAccountsRepository(args)
		err := repository.Close()

		assert.True(t, currentCloseCalled)
		assert.True(t, finalCloseCalled)
		assert.True(t, historicalCloseCalled)

		assert.Equal(t, errCurrent, err)
	})
}
