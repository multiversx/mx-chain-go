package builtInFunctions

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func TestNewESDTLocalMintFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() (c uint64, m marshal.Marshalizer, p process.ESDTPauseHandler, r process.ESDTRoleHandler)
		exError  error
	}{
		{
			name: "NilMarshalizer",
			argsFunc: func() (c uint64, m marshal.Marshalizer, p process.ESDTPauseHandler, r process.ESDTRoleHandler) {
				return 0, nil, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{}
			},
			exError: process.ErrNilMarshalizer,
		},
		{
			name: "NilPauseHandler",
			argsFunc: func() (c uint64, m marshal.Marshalizer, p process.ESDTPauseHandler, r process.ESDTRoleHandler) {
				return 0, &mock.MarshalizerMock{}, nil, &mock.ESDTRoleHandlerStub{}
			},
			exError: process.ErrNilPauseHandler,
		},
		{
			name: "NilRolesHandler",
			argsFunc: func() (c uint64, m marshal.Marshalizer, p process.ESDTPauseHandler, r process.ESDTRoleHandler) {
				return 0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, nil
			},
			exError: process.ErrNilRolesHandler,
		},
		{
			name: "Ok",
			argsFunc: func() (c uint64, m marshal.Marshalizer, p process.ESDTPauseHandler, r process.ESDTRoleHandler) {
				return 0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{}
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewESDTLocalMintFunc(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestEsdtLocalMint_SetNewGasConfig(t *testing.T) {
	t.Parallel()

	esdtLocalMintF, _ := NewESDTLocalMintFunc(0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	esdtLocalMintF.SetNewGasConfig(&process.GasCost{BuiltInCost: process.BuiltInCost{
		ESDTTransfer: 500},
	})

	require.Equal(t, uint64(500), esdtLocalMintF.funcGasCost)
}

func TestEsdtLocalMint_ProcessBuiltinFunction_CalledWithValueShouldErr(t *testing.T) {
	t.Parallel()

	esdtLocalMintF, _ := NewESDTLocalMintFunc(0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	_, err := esdtLocalMintF.ProcessBuiltinFunction(&mock.AccountWrapMock{}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(1),
		},
	})
	require.Equal(t, process.ErrBuiltInFunctionCalledWithValue, err)
}

func TestEsdtLocalMint_ProcessBuiltinFunction_CheckAllowToExecuteShouldErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local err")
	esdtLocalMintF, _ := NewESDTLocalMintFunc(0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account state.UserAccountHandler, tokenID []byte, action []byte) error {
			return localErr
		},
	})

	_, err := esdtLocalMintF.ProcessBuiltinFunction(&mock.AccountWrapMock{}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
			Arguments: [][]byte{[]byte("arg1"), []byte("arg2")},
		},
	})
	require.Equal(t, localErr, err)
}

func TestEsdtLocalMint_ProcessBuiltinFunction_CannotAddToEsdtBalanceShouldErr(t *testing.T) {
	t.Parallel()

	esdtLocalMintF, _ := NewESDTLocalMintFunc(0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account state.UserAccountHandler, tokenID []byte, action []byte) error {
			return nil
		},
	})

	localErr := errors.New("local err")
	_, err := esdtLocalMintF.ProcessBuiltinFunction(&mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					return nil, localErr
				},
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					return localErr
				},
			}
		},
	}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
			Arguments: [][]byte{[]byte("arg1"), big.NewInt(1).Bytes()},
		},
	})
	require.Equal(t, localErr, err)
}

func TestEsdtLocalMint_ProcessBuiltinFunction_ShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtLocalMintF, _ := NewESDTLocalMintFunc(50, marshalizer, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account state.UserAccountHandler, tokenID []byte, action []byte) error {
			return nil
		},
	})

	sndAccout := &mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					esdtData := &esdt.ESDigitalToken{Value: big.NewInt(100)}
					return marshalizer.Marshal(esdtData)
				},
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					esdtData := &esdt.ESDigitalToken{}
					_ = marshalizer.Unmarshal(esdtData, value)
					require.Equal(t, big.NewInt(101), esdtData.Value)
					return nil
				},
			}
		},
	}
	vmOutput, err := esdtLocalMintF.ProcessBuiltinFunction(sndAccout, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("arg1"), big.NewInt(1).Bytes()},
			GasProvided: 500,
		},
	})
	require.Equal(t, nil, err)

	expectedVMOutput := &vmcommon.VMOutput{
		ReturnCode:   vmcommon.Ok,
		GasRemaining: 450,
	}
	require.Equal(t, expectedVMOutput, vmOutput)
}
