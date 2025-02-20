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

func TestNewSCProcessorHelper(t *testing.T) {
	t.Parallel()

	t.Run("NilAccountsAdapter should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		args.Accounts = nil
		sch, err := scrCommon.NewSCProcessorHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilAccountsAdapter)
	})
	t.Run("NilShardCoordinator should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		args.ShardCoordinator = nil
		sch, err := scrCommon.NewSCProcessorHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilShardCoordinator)
	})
	t.Run("NilMarshalizer should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		args.Marshalizer = nil
		sch, err := scrCommon.NewSCProcessorHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilMarshalizer)
	})
	t.Run("NilHasher should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		args.Hasher = nil
		sch, err := scrCommon.NewSCProcessorHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilHasher)
	})
	t.Run("NilPubkeyConverter should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		args.PubkeyConverter = nil
		sch, err := scrCommon.NewSCProcessorHelper(args)
		require.Nil(t, sch)
		require.ErrorIs(t, err, process.ErrNilPubkeyConverter)
	})
	t.Run("ValidArgs should not err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		sch, err := scrCommon.NewSCProcessorHelper(args)
		require.NotNil(t, sch)
		require.NoError(t, err)
	})
}

func TestSCProcessorHelper_GetAccountFromAddress(t *testing.T) {
	t.Parallel()

	t.Run("not same shard should not err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
			ComputeIdCalled: func(address []byte) uint32 {
				return 1
			},
			SelfIDCalled: func() uint32 {
				return 2
			},
		}
		sch, _ := scrCommon.NewSCProcessorHelper(args)

		account, err := sch.GetAccountFromAddress([]byte{})
		require.Nil(t, account)
		require.Nil(t, err)
	})
	t.Run("load account with err should err", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected")

		args := getSCProcessorHelperArgs()
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return nil, expectedErr
			},
		}
		sch, _ := scrCommon.NewSCProcessorHelper(args)

		account, err := sch.GetAccountFromAddress([]byte{})
		require.Nil(t, account)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("load account with wrong type should err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &testState.BaseAccountMock{}, nil
			},
		}
		sch, _ := scrCommon.NewSCProcessorHelper(args)

		account, err := sch.GetAccountFromAddress([]byte{})
		require.Nil(t, account)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})
	t.Run("load account with success should not err", func(t *testing.T) {
		t.Parallel()

		args := getSCProcessorHelperArgs()
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &testState.AccountWrapMock{}, nil
			},
		}
		sch, _ := scrCommon.NewSCProcessorHelper(args)

		account, err := sch.GetAccountFromAddress([]byte{})
		require.NotNil(t, account)
		require.NoError(t, err)
	})
}

func TestSCProcessorHelper_CheckSCRBeforeProcessing(t *testing.T) {
	t.Parallel()

	t.Run("nil SCR should not err", func(t *testing.T) {
		args := getSCProcessorHelperArgs()
		sch, _ := scrCommon.NewSCProcessorHelper(args)

		scrData, err := sch.CheckSCRBeforeProcessing(nil)
		require.Nil(t, scrData)
		require.ErrorIs(t, err, process.ErrNilSmartContractResult)
	})
	t.Run("marshaller err should err", func(t *testing.T) {
		args := getSCProcessorHelperArgs()
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
		sch, _ := scrCommon.NewSCProcessorHelper(args)

		scr := &smartContractResult.SmartContractResult{
			Nonce:   0,
			RcvAddr: []byte{1},
			SndAddr: []byte{1},
		}
		scrData, err := sch.CheckSCRBeforeProcessing(scr)
		require.Nil(t, scrData)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("load sender error should err", func(t *testing.T) {
		args := getSCProcessorHelperArgs()
		expectedErr := errors.New("expected")
		senderAddr := []byte{1}
		receiverAddr := []byte{2}
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(address, senderAddr) {
					return nil, expectedErr
				}
				return &testState.AccountWrapMock{}, nil
			},
		}

		sch, _ := scrCommon.NewSCProcessorHelper(args)

		scr := &smartContractResult.SmartContractResult{
			Nonce:   0,
			RcvAddr: receiverAddr,
			SndAddr: senderAddr,
		}
		scrData, err := sch.CheckSCRBeforeProcessing(scr)
		require.Nil(t, scrData)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("load receiver error should err", func(t *testing.T) {
		args := getSCProcessorHelperArgs()
		expectedErr := errors.New("expected")
		senderAddr := []byte{1}
		receiverAddr := []byte{2}
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(address, receiverAddr) {
					return nil, expectedErr
				}
				return &testState.AccountWrapMock{}, nil
			},
		}

		sch, _ := scrCommon.NewSCProcessorHelper(args)

		scr := &smartContractResult.SmartContractResult{
			Nonce:   0,
			RcvAddr: receiverAddr,
			SndAddr: senderAddr,
		}
		scrData, err := sch.CheckSCRBeforeProcessing(scr)
		require.Nil(t, scrData)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("should work", func(t *testing.T) {
		args := getSCProcessorHelperArgs()
		expectedHash := []byte{123}
		expectedJournalLen := 7
		senderAddr := []byte{1}
		receiverAddr := []byte{2}
		expectedSender := &testState.AccountWrapMock{
			CodeHash: senderAddr, // for testing purposes
		}
		expectedReceiver := &testState.AccountWrapMock{
			CodeHash: receiverAddr, // for testing purposes
		}

		args.Hasher = &testscommon.HasherStub{
			ComputeCalled: func(s string) []byte {
				return expectedHash
			},
		}
		args.Accounts = &testState.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(address, receiverAddr) {
					return expectedReceiver, nil
				}
				if bytes.Equal(address, senderAddr) {
					return expectedSender, nil
				}
				return nil, errors.New("unexpected")
			},
			JournalLenCalled: func() int {
				return expectedJournalLen
			},
		}
		sch, _ := scrCommon.NewSCProcessorHelper(args)

		scr := &smartContractResult.SmartContractResult{
			Nonce:   0,
			RcvAddr: receiverAddr,
			SndAddr: senderAddr,
		}
		scrData, err := sch.CheckSCRBeforeProcessing(scr)
		require.NotNil(t, scrData)
		require.Equal(t, expectedHash, scrData.GetHash())
		require.Equal(t, expectedSender, scrData.GetSender())
		require.Equal(t, expectedReceiver, scrData.GetDestination())
		require.NotNil(t, scrData.GetDestination())
		require.Equal(t, expectedJournalLen, scrData.GetSnapshot())
		require.Nil(t, err)
	})

}

func TestSCProcessorHelper_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := getSCProcessorHelperArgs()
	sch, _ := scrCommon.NewSCProcessorHelper(args)

	require.False(t, sch.IsInterfaceNil())
}

func getSCProcessorHelperArgs() scrCommon.SCProcessorHelperArgs {
	return scrCommon.SCProcessorHelperArgs{
		Accounts:         &testState.AccountsStub{},
		ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
		Marshalizer:      &testscommon.MarshallerStub{},
		Hasher:           &testscommon.HasherStub{},
		PubkeyConverter:  &testscommon.PubkeyConverterStub{},
	}
}
