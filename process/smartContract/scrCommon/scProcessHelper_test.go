package scrCommon_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/testscommon"
	testState "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewSCProcessHelper(t *testing.T) {
	t.Parallel()

	t.Run("NilAccountsAdapter should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		args.Accounts = nil
		sch, err := scrCommon.NewSCProcessHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilAccountsAdapter)
	})
	t.Run("NilShardCoordinator should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		args.ShardCoordinator = nil
		sch, err := scrCommon.NewSCProcessHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilShardCoordinator)
	})
	t.Run("NilMarshalizer should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		args.Marshalizer = nil
		sch, err := scrCommon.NewSCProcessHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilMarshalizer)
	})
	t.Run("NilHasher should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		args.Hasher = nil
		sch, err := scrCommon.NewSCProcessHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilHasher)
	})
	t.Run("NilPubkeyConverter should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		args.PubkeyConverter = nil
		sch, err := scrCommon.NewSCProcessHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilPubkeyConverter)
	})
	t.Run("ValidArgs should not err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		sch, err := scrCommon.NewSCProcessHelper(args)
		require.NotNil(t, sch)
		require.NoError(t, err)
	})
}

func TestSCProcessHelper_GetAccountFromAddress(t *testing.T) {
	t.Parallel()

	t.Run("not same shard should not err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
			ComputeIdCalled: func(address []byte) uint32 {
				return 1
			},
			SelfIDCalled: func() uint32 {
				return 2
			},
		}
		sch, _ := scrCommon.NewSCProcessHelper(args)

		account, err := sch.GetAccountFromAddress([]byte{})
		require.Nil(t, account)
		require.Nil(t, err)
	})
	t.Run("load account with err should err", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected")

		args := getSCProcessHelperArgs()
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return nil, expectedErr
			},
		}
		sch, _ := scrCommon.NewSCProcessHelper(args)

		account, err := sch.GetAccountFromAddress([]byte{})
		require.Nil(t, account)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("load account with wrong type should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &testState.BaseAccountMock{}, nil
			},
		}
		sch, _ := scrCommon.NewSCProcessHelper(args)

		account, err := sch.GetAccountFromAddress([]byte{})
		require.Nil(t, account)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})
	t.Run("load account with success should not err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessHelperArgs()
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &testState.AccountWrapMock{}, nil
			},
		}
		sch, _ := scrCommon.NewSCProcessHelper(args)

		account, err := sch.GetAccountFromAddress([]byte{})
		require.NotNil(t, account)
		require.NoError(t, err)
	})
}

func TestSCProcessHelper_CheckSCRBeforeProcessing(t *testing.T) {
	t.Parallel()

	t.Run("nil SCR should not err", func(t *testing.T) {
		args := getSCProcessHelperArgs()
		sch, _ := scrCommon.NewSCProcessHelper(args)

		account, err := sch.CheckSCRBeforeProcessing(nil)
		require.Nil(t, account)
		require.ErrorIs(t, err, process.ErrNilSmartContractResult)
	})
	t.Run("marshaller err should err", func(t *testing.T) {
		args := getSCProcessHelperArgs()
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &testState.AccountWrapMock{}, nil
			},
		}
		expectedErr := errors.New("expected")
		args.Marshalizer = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		sch, _ := scrCommon.NewSCProcessHelper(args)

		scr := &smartContractResult.SmartContractResult{
			Nonce:   0,
			RcvAddr: []byte{1},
			SndAddr: []byte{1},
		}
		account, err := sch.CheckSCRBeforeProcessing(scr)
		require.Nil(t, account)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("load sender error should err", func(t *testing.T) {
		args := getSCProcessHelperArgs()
		expectedErr := errors.New("expected")
		senderAddr := []byte{1}
		receiverAddr := []byte{2}
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				if bytes.Compare(address, senderAddr) == 0 {
					return nil, expectedErr
				}
				return &testState.AccountWrapMock{}, nil
			},
		}

		sch, _ := scrCommon.NewSCProcessHelper(args)

		scr := &smartContractResult.SmartContractResult{
			Nonce:   0,
			RcvAddr: receiverAddr,
			SndAddr: senderAddr,
		}
		account, err := sch.CheckSCRBeforeProcessing(scr)
		require.Nil(t, account)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("load receiver error should err", func(t *testing.T) {
		args := getSCProcessHelperArgs()
		expectedErr := errors.New("expected")
		senderAddr := []byte{1}
		receiverAddr := []byte{2}
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				if bytes.Compare(address, receiverAddr) == 0 {
					return nil, expectedErr
				}
				return &testState.AccountWrapMock{}, nil
			},
		}

		sch, _ := scrCommon.NewSCProcessHelper(args)

		scr := &smartContractResult.SmartContractResult{
			Nonce:   0,
			RcvAddr: receiverAddr,
			SndAddr: senderAddr,
		}
		account, err := sch.CheckSCRBeforeProcessing(scr)
		require.Nil(t, account)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("not nil SCR should not err", func(t *testing.T) {
		args := getSCProcessHelperArgs()
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &testState.AccountWrapMock{}, nil
			},
		}
		sch, _ := scrCommon.NewSCProcessHelper(args)

		scr := &smartContractResult.SmartContractResult{
			Nonce:   0,
			RcvAddr: []byte{1},
			SndAddr: []byte{1},
		}
		account, err := sch.CheckSCRBeforeProcessing(scr)
		require.NotNil(t, account)
		require.Nil(t, err)
	})

}

func TestSCProcessHelper_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := getSCProcessHelperArgs()
	sch, _ := scrCommon.NewSCProcessHelper(args)

	require.False(t, sch.IsInterfaceNil())
}

func getSCProcessHelperArgs() scrCommon.SCProcessHelperArgs {
	return scrCommon.SCProcessHelperArgs{
		Accounts:         &testState.AccountsStub{},
		ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
		Marshalizer:      &testscommon.MarshallerStub{},
		Hasher:           &testscommon.HasherStub{},
		PubkeyConverter:  &testscommon.PubkeyConverterStub{},
	}
}
