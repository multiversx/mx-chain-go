package builtInFunctions

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewESDTNFTBurnFunc(t *testing.T) {
	t.Parallel()

	// nil marshalizer
	ebf, err := NewESDTNFTBurnFunc(10, nil, nil, nil)
	require.True(t, check.IfNil(ebf))
	require.Equal(t, process.ErrNilMarshalizer, err)

	// nil pause handler
	ebf, err = NewESDTNFTBurnFunc(10, &mock.MarshalizerMock{}, nil, nil)
	require.True(t, check.IfNil(ebf))
	require.Equal(t, process.ErrNilPauseHandler, err)

	// nil roles handler
	ebf, err = NewESDTNFTBurnFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, nil)
	require.True(t, check.IfNil(ebf))
	require.Equal(t, process.ErrNilRolesHandler, err)

	// should work
	ebf, err = NewESDTNFTBurnFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})
	require.False(t, check.IfNil(ebf))
	require.NoError(t, err)
}

func TestESDTNFTBurn_SetNewGasConfig_NilGasCost(t *testing.T) {
	t.Parallel()

	defaultGasCost := uint64(10)
	ebf, _ := NewESDTNFTBurnFunc(defaultGasCost, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	ebf.SetNewGasConfig(nil)
	require.Equal(t, defaultGasCost, ebf.funcGasCost)
}

func TestEsdtNFTBurnFunc_SetNewGasConfig_ShouldWork(t *testing.T) {
	t.Parallel()

	defaultGasCost := uint64(10)
	newGasCost := uint64(37)
	ebf, _ := NewESDTNFTBurnFunc(defaultGasCost, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	ebf.SetNewGasConfig(
		&process.GasCost{
			BuiltInCost: process.BuiltInCost{
				ESDTNFTBurn: newGasCost,
			},
		},
	)

	require.Equal(t, newGasCost, ebf.funcGasCost)
}

func TestEsdtNFTBurnFunc_ProcessBuiltinFunctionErrorOnCheckESDTNFTCreateBurnAddInput(t *testing.T) {
	t.Parallel()

	ebf, _ := NewESDTNFTBurnFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	// nil vm input
	output, err := ebf.ProcessBuiltinFunction(mock.NewAccountWrapMock([]byte("addr")), nil, nil)
	require.Nil(t, output)
	require.Equal(t, process.ErrNilVmInput, err)

	// vm input - value not zero
	output, err = ebf.ProcessBuiltinFunction(
		mock.NewAccountWrapMock([]byte("addr")),
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue: big.NewInt(37),
			},
		},
	)
	require.Nil(t, output)
	require.Equal(t, process.ErrBuiltInFunctionCalledWithValue, err)

	// vm input - invalid number of arguments
	output, err = ebf.ProcessBuiltinFunction(
		mock.NewAccountWrapMock([]byte("addr")),
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue: big.NewInt(0),
				Arguments: [][]byte{[]byte("single arg")},
			},
		},
	)
	require.Nil(t, output)
	require.Equal(t, process.ErrInvalidArguments, err)

	// vm input - invalid number of arguments
	output, err = ebf.ProcessBuiltinFunction(
		mock.NewAccountWrapMock([]byte("addr")),
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue: big.NewInt(0),
				Arguments: [][]byte{[]byte("arg0")},
			},
		},
	)
	require.Nil(t, output)
	require.Equal(t, process.ErrInvalidArguments, err)

	// vm input - invalid receiver
	output, err = ebf.ProcessBuiltinFunction(
		mock.NewAccountWrapMock([]byte("addr")),
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:  big.NewInt(0),
				Arguments:  [][]byte{[]byte("arg0"), []byte("arg1")},
				CallerAddr: []byte("address 1"),
			},
			RecipientAddr: []byte("address 2"),
		},
	)
	require.Nil(t, output)
	require.Equal(t, process.ErrInvalidRcvAddr, err)

	// nil user account
	output, err = ebf.ProcessBuiltinFunction(
		nil,
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:  big.NewInt(0),
				Arguments:  [][]byte{[]byte("arg0"), []byte("arg1")},
				CallerAddr: []byte("address 1"),
			},
			RecipientAddr: []byte("address 1"),
		},
	)
	require.Nil(t, output)
	require.Equal(t, process.ErrNilUserAccount, err)

	// not enough gas
	output, err = ebf.ProcessBuiltinFunction(
		mock.NewAccountWrapMock([]byte("addr")),
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:   big.NewInt(0),
				Arguments:   [][]byte{[]byte("arg0"), []byte("arg1")},
				CallerAddr:  []byte("address 1"),
				GasProvided: 1,
			},
			RecipientAddr: []byte("address 1"),
		},
	)
	require.Nil(t, output)
	require.Equal(t, process.ErrNotEnoughGas, err)
}

func TestEsdtNFTBurnFunc_ProcessBuiltinFunctionInvalidNumberOfArguments(t *testing.T) {
	t.Parallel()

	ebf, _ := NewESDTNFTBurnFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})
	output, err := ebf.ProcessBuiltinFunction(
		mock.NewAccountWrapMock([]byte("addr")),
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:   big.NewInt(0),
				Arguments:   [][]byte{[]byte("arg0"), []byte("arg1")},
				CallerAddr:  []byte("address 1"),
				GasProvided: 12,
			},
			RecipientAddr: []byte("address 1"),
		},
	)
	require.Nil(t, output)
	require.Equal(t, process.ErrInvalidArguments, err)
}

func TestEsdtNFTBurnFunc_ProcessBuiltinFunctionCheckAllowedToExecuteError(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	rolesHandler := &mock.ESDTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(_ state.UserAccountHandler, _ []byte, _ []byte) error {
			return localErr
		},
	}
	ebf, _ := NewESDTNFTBurnFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, rolesHandler)
	output, err := ebf.ProcessBuiltinFunction(
		mock.NewAccountWrapMock([]byte("addr")),
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:   big.NewInt(0),
				Arguments:   [][]byte{[]byte("arg0"), []byte("arg1"), []byte("arg2")},
				CallerAddr:  []byte("address 1"),
				GasProvided: 12,
			},
			RecipientAddr: []byte("address 1"),
		},
	)

	require.Nil(t, output)
	require.Equal(t, localErr, err)
}

func TestEsdtNFTBurnFunc_ProcessBuiltinFunctionNewSenderShouldErr(t *testing.T) {
	t.Parallel()

	ebf, _ := NewESDTNFTBurnFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})
	output, err := ebf.ProcessBuiltinFunction(
		mock.NewAccountWrapMock([]byte("addr")),
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:   big.NewInt(0),
				Arguments:   [][]byte{[]byte("arg0"), []byte("arg1"), []byte("arg2")},
				CallerAddr:  []byte("address 1"),
				GasProvided: 12,
			},
			RecipientAddr: []byte("address 1"),
		},
	)

	require.Nil(t, output)
	require.Error(t, err)
	require.Equal(t, process.ErrNewNFTDataOnSenderAddress, err)
}

func TestEsdtNFTBurnFunc_ProcessBuiltinFunctionMetaDataMissing(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	ebf, _ := NewESDTNFTBurnFunc(10, marshalizer, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	userAcc := mock.NewAccountWrapMock([]byte("addr"))
	esdtData := &esdt.ESDigitalToken{}
	esdtDataBytes, _ := marshalizer.Marshal(esdtData)
	tailLength := 28 // len(esdtKey) + "identifier"
	esdtDataBytes = append(esdtDataBytes, make([]byte, tailLength)...)
	userAcc.SetDataTrie(&testscommon.TrieStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return esdtDataBytes, nil
		},
	})
	output, err := ebf.ProcessBuiltinFunction(
		userAcc,
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:   big.NewInt(0),
				Arguments:   [][]byte{[]byte("arg0"), []byte("arg1"), []byte("arg2")},
				CallerAddr:  []byte("address 1"),
				GasProvided: 12,
			},
			RecipientAddr: []byte("address 1"),
		},
	)

	require.Nil(t, output)
	require.Equal(t, process.ErrNFTDoesNotHaveMetadata, err)
}

func TestEsdtNFTBurnFunc_ProcessBuiltinFunctionInvalidBurnQuantity(t *testing.T) {
	t.Parallel()

	initialQuantity := big.NewInt(55)
	quantityToBurn := big.NewInt(75)

	marshalizer := &mock.MarshalizerMock{}

	ebf, _ := NewESDTNFTBurnFunc(10, marshalizer, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	userAcc := mock.NewAccountWrapMock([]byte("addr"))
	esdtData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Name: []byte("test"),
		},
		Value: initialQuantity,
	}
	esdtDataBytes, _ := marshalizer.Marshal(esdtData)
	tailLength := 28 // len(esdtKey) + "identifier"
	esdtDataBytes = append(esdtDataBytes, make([]byte, tailLength)...)
	userAcc.SetDataTrie(&testscommon.TrieStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return esdtDataBytes, nil
		},
	})
	output, err := ebf.ProcessBuiltinFunction(
		userAcc,
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:   big.NewInt(0),
				Arguments:   [][]byte{[]byte("arg0"), []byte("arg1"), quantityToBurn.Bytes()},
				CallerAddr:  []byte("address 1"),
				GasProvided: 12,
			},
			RecipientAddr: []byte("address 1"),
		},
	)

	require.Nil(t, output)
	require.Equal(t, process.ErrInvalidNFTQuantity, err)
}

func TestEsdtNFTBurnFunc_ProcessBuiltinFunctionShouldErrOnSaveBecauseTokenIsPaused(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	pauseHandler := &mock.PauseHandlerStub{
		IsPausedCalled: func(_ []byte) bool {
			return true
		},
	}

	ebf, _ := NewESDTNFTBurnFunc(10, marshalizer, pauseHandler, &mock.ESDTRoleHandlerStub{})

	userAcc := mock.NewAccountWrapMock([]byte("addr"))
	esdtData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Name: []byte("test"),
		},
		Value: big.NewInt(10),
	}
	esdtDataBytes, _ := marshalizer.Marshal(esdtData)
	tailLength := 28 // len(esdtKey) + "identifier"
	esdtDataBytes = append(esdtDataBytes, make([]byte, tailLength)...)
	userAcc.SetDataTrie(&testscommon.TrieStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return esdtDataBytes, nil
		},
	})
	output, err := ebf.ProcessBuiltinFunction(
		userAcc,
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:   big.NewInt(0),
				Arguments:   [][]byte{[]byte("arg0"), []byte("arg1"), big.NewInt(5).Bytes()},
				CallerAddr:  []byte("address 1"),
				GasProvided: 12,
			},
			RecipientAddr: []byte("address 1"),
		},
	)

	require.Nil(t, output)
	require.Equal(t, process.ErrESDTTokenIsPaused, err)
}

func TestEsdtNFTBurnFunc_ProcessBuiltinFunctionShouldWork(t *testing.T) {
	t.Parallel()

	tokenIdentifier := "testTkn"
	key := core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + tokenIdentifier

	nonce := big.NewInt(33)
	initialQuantity := big.NewInt(50)
	quantityToBurn := big.NewInt(37)
	expectedQuantity := big.NewInt(0).Sub(initialQuantity, quantityToBurn)

	marshalizer := &mock.MarshalizerMock{}
	ebf, _ := NewESDTNFTBurnFunc(10, marshalizer, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	userAcc := mock.NewAccountWrapMock([]byte("addr"))
	esdtData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Name: []byte("test"),
		},
		Value: initialQuantity,
	}
	esdtDataBytes, _ := marshalizer.Marshal(esdtData)
	tokenKey := append([]byte(key), nonce.Bytes()...)
	tailLength := len(tokenKey) + len("identifier")
	esdtDataBytes = append(esdtDataBytes, make([]byte, tailLength)...)
	userAcc.SetDataTrie(&testscommon.TrieStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return esdtDataBytes, nil
		},
	})
	output, err := ebf.ProcessBuiltinFunction(
		userAcc,
		nil,
		&vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallValue:   big.NewInt(0),
				Arguments:   [][]byte{[]byte(tokenIdentifier), nonce.Bytes(), quantityToBurn.Bytes()},
				CallerAddr:  []byte("address 1"),
				GasProvided: 12,
			},
			RecipientAddr: []byte("address 1"),
		},
	)

	require.NotNil(t, output)
	require.NoError(t, err)
	require.Equal(t, vmcommon.Ok, output.ReturnCode)

	res, err := userAcc.DataTrieTracker().RetrieveValue([]byte(key))
	require.NoError(t, err)
	require.NotNil(t, res)

	finalTokenData := esdt.ESDigitalToken{}
	_ = marshalizer.Unmarshal(&finalTokenData, res)
	require.Equal(t, expectedQuantity.Bytes(), finalTokenData.Value.Bytes())
}
