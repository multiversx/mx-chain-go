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

func TestNewESDTLocalBurnFunc(t *testing.T) {
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
			name: "NilPouseHandler",
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
			_, err := NewESDTLocalBurnFunc(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestEsdtLocalBurn_ProcessBuiltinFunction_CalledWithValueShouldErr(t *testing.T) {
	t.Parallel()

	esdtLocalBurnF, _ := NewESDTLocalBurnFunc(0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	_, err := esdtLocalBurnF.ProcessBuiltinFunction(&mock.AccountWrapMock{}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(1),
		},
	})
	require.Equal(t, process.ErrBuiltInFunctionCalledWithValue, err)
}

func TestEsdtLocalBurn_ProcessBuiltinFunction_CheckAllowToExecuteShouldErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local err")
	esdtLocalBurnF, _ := NewESDTLocalBurnFunc(0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account state.UserAccountHandler, tokenID []byte, action []byte) error {
			return localErr
		},
	})

	_, err := esdtLocalBurnF.ProcessBuiltinFunction(&mock.AccountWrapMock{}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
			Arguments: [][]byte{[]byte("arg1"), []byte("arg2")},
		},
	})
	require.Equal(t, localErr, err)
}

func TestEsdtLocalBurn_ProcessBuiltinFunction_CannotAddToEsdtBalanceShouldErr(t *testing.T) {
	t.Parallel()

	esdtLocalBurnF, _ := NewESDTLocalBurnFunc(0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account state.UserAccountHandler, tokenID []byte, action []byte) error {
			return nil
		},
	})

	localErr := errors.New("local err")
	_, err := esdtLocalBurnF.ProcessBuiltinFunction(&mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					return nil, localErr
				},
			}
		},
	}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
			Arguments: [][]byte{[]byte("arg1"), []byte("arg2")},
		},
	})
	require.Equal(t, process.ErrInsufficientFunds, err)
}

func TestEsdtLocalBurn_ProcessBuiltinFunction_ShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdtLocalBurnF, _ := NewESDTLocalBurnFunc(50, marshalizer, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{
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
					require.Equal(t, big.NewInt(99), esdtData.Value)
					return nil
				},
			}
		},
	}
	vmOutput, err := esdtLocalBurnF.ProcessBuiltinFunction(sndAccout, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
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

func TestEsdtLocalBurn_SetNewGasConfig(t *testing.T) {
	t.Parallel()

	esdtLocalBurnF, _ := NewESDTLocalBurnFunc(0, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	esdtLocalBurnF.SetNewGasConfig(&process.GasCost{BuiltInCost: process.BuiltInCost{
		ESDTBurn: 500},
	})

	require.Equal(t, uint64(500), esdtLocalBurnF.funcGasCost)
}

func TestCheckInputArgumentsForLocalAction_InvalidRecipientAddr(t *testing.T) {
	t.Parallel()

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{[]byte("arg1"), big.NewInt(1).Bytes()},
			CallerAddr: []byte("caller"),
		},
		RecipientAddr: []byte("rec"),
	}

	err := checkInputArgumentsForLocalAction(&mock.UserAccountStub{}, vmInput, 0)
	require.Equal(t, process.ErrInvalidRcvAddr, err)
}

func TestCheckInputArgumentsForLocalAction_NilUserAccount(t *testing.T) {
	t.Parallel()

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{[]byte("arg1"), big.NewInt(1).Bytes()},
			CallerAddr: []byte("caller"),
		},
		RecipientAddr: []byte("caller"),
	}

	err := checkInputArgumentsForLocalAction(nil, vmInput, 0)
	require.Equal(t, process.ErrNilUserAccount, err)
}

func TestCheckInputArgumentsForLocalAction_NotEnoughGas(t *testing.T) {
	t.Parallel()

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("arg1"), big.NewInt(10).Bytes()},
			CallerAddr:  []byte("caller"),
			GasProvided: 1,
		},
		RecipientAddr: []byte("caller"),
	}

	err := checkInputArgumentsForLocalAction(&mock.UserAccountStub{}, vmInput, 500)
	require.Equal(t, process.ErrNotEnoughGas, err)
}
