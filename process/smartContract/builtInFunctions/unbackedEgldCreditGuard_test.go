package builtInFunctions

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/vmcommonMocks"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewUnbackedEGLDCreditGuard(t *testing.T) {
	t.Parallel()

	t.Run("nil inner should error", func(t *testing.T) {
		t.Parallel()

		guard, err := newUnbackedEGLDCreditGuard(
			nil,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&testscommon.ShardsCoordinatorMock{},
		)
		require.Nil(t, guard)
		require.Equal(t, process.ErrNilBuiltInFunction, err)
	})
	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		guard, err := newUnbackedEGLDCreditGuard(
			&vmcommonMocks.BuiltInFunctionExecutableStub{},
			nil,
			&testscommon.ShardsCoordinatorMock{},
		)
		require.Nil(t, guard)
		require.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		guard, err := newUnbackedEGLDCreditGuard(
			&vmcommonMocks.BuiltInFunctionExecutableStub{},
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			nil,
		)
		require.Nil(t, guard)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})
}

func TestSumEGLDFromMultiTransferArgs(t *testing.T) {
	t.Parallel()

	egldAmount := big.NewInt(1000)
	otherToken := []byte("OTHER-abcdef")

	t.Run("dest format sums EGLD-000000 only", func(t *testing.T) {
		t.Parallel()

		args := [][]byte{
			big.NewInt(2).Bytes(),
			[]byte(vmcommon.EGLDIdentifier),
			{},
			egldAmount.Bytes(),
			otherToken,
			{},
			big.NewInt(5).Bytes(),
		}

		require.Equal(t, egldAmount, sumEGLDFromMultiTransferArgs(args, false))
	})
	t.Run("sender format skips destination address", func(t *testing.T) {
		t.Parallel()

		args := [][]byte{
			bytes.Repeat([]byte{2}, 32),
			big.NewInt(1).Bytes(),
			[]byte(vmcommon.EGLDIdentifier),
			{},
			egldAmount.Bytes(),
		}

		require.Equal(t, egldAmount, sumEGLDFromMultiTransferArgs(args, true))
		require.Equal(t, int64(0), sumEGLDFromMultiTransferArgs(args, false).Int64())
	})
	t.Run("no EGLD token yields zero", func(t *testing.T) {
		t.Parallel()

		args := [][]byte{
			big.NewInt(1).Bytes(),
			otherToken,
			{},
			egldAmount.Bytes(),
		}

		require.Equal(t, int64(0), sumEGLDFromMultiTransferArgs(args, false).Int64())
	})
}

func TestUnbackedEGLDCreditGuard_DestPathRequiresCallValue(t *testing.T) {
	t.Parallel()

	sender := bytes.Repeat([]byte{1}, 32)
	dest := bytes.Repeat([]byte{2}, 32)
	egldAmount := big.NewInt(1000)
	destArgs := [][]byte{
		big.NewInt(1).Bytes(),
		[]byte(vmcommon.EGLDIdentifier),
		{},
		egldAmount.Bytes(),
	}

	t.Run("flag off allows dest path without CallValue", func(t *testing.T) {
		t.Parallel()

		innerCalled := false
		inner := &vmcommonMocks.BuiltInFunctionExecutableStub{
			ProcessBuiltinFunctionCalled: func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				innerCalled = true
				return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
			},
		}
		guard, err := newUnbackedEGLDCreditGuard(
			inner,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&testscommon.ShardsCoordinatorMock{},
		)
		require.Nil(t, err)

		vmOutput, err := guard.ProcessBuiltinFunction(
			nil,
			&vmcommonMocks.UserAccountStub{},
			&vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr:  sender,
					CallValue:   big.NewInt(0),
					Arguments:   destArgs,
					GasProvided: 100000,
				},
				RecipientAddr: dest,
				Function:      core.BuiltInFunctionMultiESDTNFTTransfer,
			},
		)
		require.Nil(t, err)
		require.NotNil(t, vmOutput)
		require.True(t, innerCalled)
	})
	t.Run("flag on rejects dest path when CallValue does not cover EGLD", func(t *testing.T) {
		t.Parallel()

		innerCalled := false
		inner := &vmcommonMocks.BuiltInFunctionExecutableStub{
			ProcessBuiltinFunctionCalled: func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				innerCalled = true
				return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
			},
		}
		guard, err := newUnbackedEGLDCreditGuard(
			inner,
			enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GuardUnbackedEGLDInMultiTransferFlag),
			&testscommon.ShardsCoordinatorMock{},
		)
		require.Nil(t, err)

		vmOutput, err := guard.ProcessBuiltinFunction(
			nil,
			&vmcommonMocks.UserAccountStub{},
			&vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr:  sender,
					CallValue:   big.NewInt(0),
					Arguments:   destArgs,
					GasProvided: 100000,
				},
				RecipientAddr: dest,
				Function:      core.BuiltInFunctionMultiESDTNFTTransfer,
			},
		)
		require.Equal(t, process.ErrUnbackedEGLDInMultiTransfer, err)
		require.Nil(t, vmOutput)
		require.False(t, innerCalled)
	})
	t.Run("flag on allows dest path when CallValue covers EGLD", func(t *testing.T) {
		t.Parallel()

		innerCalled := false
		inner := &vmcommonMocks.BuiltInFunctionExecutableStub{
			ProcessBuiltinFunctionCalled: func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				innerCalled = true
				return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
			},
		}
		guard, err := newUnbackedEGLDCreditGuard(
			inner,
			enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GuardUnbackedEGLDInMultiTransferFlag),
			&testscommon.ShardsCoordinatorMock{},
		)
		require.Nil(t, err)

		vmOutput, err := guard.ProcessBuiltinFunction(
			nil,
			&vmcommonMocks.UserAccountStub{},
			&vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr:  sender,
					CallValue:   big.NewInt(0).Set(egldAmount),
					Arguments:   destArgs,
					GasProvided: 100000,
				},
				RecipientAddr: dest,
				Function:      core.BuiltInFunctionMultiESDTNFTTransfer,
			},
		)
		require.Nil(t, err)
		require.NotNil(t, vmOutput)
		require.True(t, innerCalled)
	})
	t.Run("non-EGLD dest path is unchanged", func(t *testing.T) {
		t.Parallel()

		innerCalled := false
		inner := &vmcommonMocks.BuiltInFunctionExecutableStub{
			ProcessBuiltinFunctionCalled: func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				innerCalled = true
				return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
			},
		}
		guard, err := newUnbackedEGLDCreditGuard(
			inner,
			enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GuardUnbackedEGLDInMultiTransferFlag),
			&testscommon.ShardsCoordinatorMock{},
		)
		require.Nil(t, err)

		vmOutput, err := guard.ProcessBuiltinFunction(
			nil,
			&vmcommonMocks.UserAccountStub{},
			&vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr: sender,
					CallValue:  big.NewInt(0),
					Arguments: [][]byte{
						big.NewInt(1).Bytes(),
						[]byte("TOKEN-abcdef"),
						{},
						egldAmount.Bytes(),
					},
					GasProvided: 100000,
				},
				RecipientAddr: dest,
				Function:      core.BuiltInFunctionMultiESDTNFTTransfer,
			},
		)
		require.Nil(t, err)
		require.NotNil(t, vmOutput)
		require.True(t, innerCalled)
	})
}

func TestUnbackedEGLDCreditGuard_SenderPathAttachesCallValue(t *testing.T) {
	t.Parallel()

	sender := bytes.Repeat([]byte{1}, 32)
	dest := bytes.Repeat([]byte{2}, 32)
	egldAmount := big.NewInt(1000)
	senderArgs := [][]byte{
		dest,
		big.NewInt(1).Bytes(),
		[]byte(vmcommon.EGLDIdentifier),
		{},
		egldAmount.Bytes(),
	}
	destFormatData := core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) +
		"@" + hex.EncodeToString([]byte(vmcommon.EGLDIdentifier)) +
		"@" +
		"@" + hex.EncodeToString(egldAmount.Bytes())

	createGuard := func(flagOn bool, shardOfDest uint32) *unbackedEGLDCreditGuard {
		inner := &vmcommonMocks.BuiltInFunctionExecutableStub{
			ProcessBuiltinFunctionCalled: func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.Ok,
					OutputAccounts: map[string]*vmcommon.OutputAccount{
						string(dest): {
							Address: dest,
							OutputTransfers: []vmcommon.OutputTransfer{
								{
									Value: big.NewInt(0),
									Data:  []byte(destFormatData),
								},
							},
						},
					},
				}, nil
			},
		}

		enableEpochs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
		if flagOn {
			enableEpochs = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GuardUnbackedEGLDInMultiTransferFlag)
		}

		guard, err := newUnbackedEGLDCreditGuard(
			inner,
			enableEpochs,
			&testscommon.ShardsCoordinatorMock{
				CurrentShard: 0,
				ComputeIdCalled: func(address []byte) uint32 {
					if bytes.Equal(address, dest) {
						return shardOfDest
					}
					return 0
				},
			},
		)
		require.Nil(t, err)
		return guard
	}

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			CallValue:   big.NewInt(0),
			Arguments:   senderArgs,
			GasProvided: 100000,
		},
		RecipientAddr: sender,
		Function:      core.BuiltInFunctionMultiESDTNFTTransfer,
	}

	t.Run("cross-shard output transfer receives EGLD amount as Value", func(t *testing.T) {
		t.Parallel()

		guard := createGuard(true, 1)
		vmOutput, err := guard.ProcessBuiltinFunction(&vmcommonMocks.UserAccountStub{}, nil, input)
		require.Nil(t, err)
		require.Equal(t, egldAmount, vmOutput.OutputAccounts[string(dest)].OutputTransfers[0].Value)
	})
	t.Run("same-shard output transfer is left unchanged", func(t *testing.T) {
		t.Parallel()

		guard := createGuard(true, 0)
		vmOutput, err := guard.ProcessBuiltinFunction(&vmcommonMocks.UserAccountStub{}, nil, input)
		require.Nil(t, err)
		require.Equal(t, int64(0), vmOutput.OutputAccounts[string(dest)].OutputTransfers[0].Value.Int64())
	})
	t.Run("flag off leaves output transfer Value unchanged", func(t *testing.T) {
		t.Parallel()

		guard := createGuard(false, 1)
		vmOutput, err := guard.ProcessBuiltinFunction(&vmcommonMocks.UserAccountStub{}, nil, input)
		require.Nil(t, err)
		require.Equal(t, int64(0), vmOutput.OutputAccounts[string(dest)].OutputTransfers[0].Value.Int64())
	})
}

func TestUnbackedEGLDCreditGuard_ForwardsInnerBehavior(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("inner error")
	gasSet := false
	inner := &vmcommonMocks.BuiltInFunctionExecutableStub{
		ProcessBuiltinFunctionCalled: func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			return nil, expectedErr
		},
		SetNewGasConfigCalled: func(gasCost *vmcommon.GasCost) {
			gasSet = true
		},
		IsActiveCalled: func() bool {
			return false
		},
		CheckIsExecutableCalled: func(senderAddr []byte, value *big.Int, receiverAddr []byte, gasProvidedForCall uint64, arguments [][]byte) error {
			return expectedErr
		},
	}
	guard, err := newUnbackedEGLDCreditGuard(
		inner,
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.GuardUnbackedEGLDInMultiTransferFlag),
		&testscommon.ShardsCoordinatorMock{},
	)
	require.Nil(t, err)

	vmOutput, err := guard.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{})
	require.Nil(t, vmOutput)
	require.Equal(t, expectedErr, err)

	guard.SetNewGasConfig(&vmcommon.GasCost{})
	require.True(t, gasSet)
	require.False(t, guard.IsActive())
	require.Equal(t, expectedErr, guard.CheckIsExecutable(nil, big.NewInt(0), nil, 0, nil))
	require.False(t, guard.IsInterfaceNil())
	require.True(t, (*unbackedEGLDCreditGuard)(nil).IsInterfaceNil())
}

func TestWrapMultiESDTNFTTransferWithUnbackedEGLDGuard(t *testing.T) {
	t.Parallel()

	t.Run("nil container should error", func(t *testing.T) {
		t.Parallel()

		err := wrapMultiESDTNFTTransferWithUnbackedEGLDGuard(
			nil,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			&testscommon.ShardsCoordinatorMock{},
		)
		require.Equal(t, process.ErrNilBuiltInFunction, err)
	})
	t.Run("already wrapped is a no-op", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		factory, err := CreateBuiltInFunctionsFactory(args)
		require.Nil(t, err)

		container := factory.BuiltInFunctionContainer()
		first, err := container.Get(core.BuiltInFunctionMultiESDTNFTTransfer)
		require.Nil(t, err)
		_, ok := first.(*unbackedEGLDCreditGuard)
		require.True(t, ok)

		err = wrapMultiESDTNFTTransferWithUnbackedEGLDGuard(
			container,
			args.EnableEpochsHandler,
			args.ShardCoordinator,
		)
		require.Nil(t, err)

		second, err := container.Get(core.BuiltInFunctionMultiESDTNFTTransfer)
		require.Nil(t, err)
		require.Equal(t, first, second)
	})
}
