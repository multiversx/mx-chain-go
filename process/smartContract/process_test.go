package smartContract

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state/accounts"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateEmptyByteSlice(size int) []byte {
	buff := make([]byte, size)

	return buff
}

func createAccounts(tx *transaction.Transaction) (state.UserAccountHandler, state.UserAccountHandler) {
	acntSrc, _ := accounts.NewUserAccount(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Balance = acntSrc.Balance.Add(acntSrc.Balance, tx.Value)
	totalFee := big.NewInt(0)
	totalFee = totalFee.Mul(big.NewInt(int64(tx.GasLimit)), big.NewInt(int64(tx.GasPrice)))
	acntSrc.Balance = acntSrc.Balance.Add(acntSrc.Balance, totalFee)

	acntDst, _ := accounts.NewUserAccount(mock.NewAddressMock(tx.RcvAddr))

	return acntSrc, acntDst
}

func FillGasMapInternal(gasMap map[string]map[string]uint64, value uint64) map[string]map[string]uint64 {
	gasMap[core.BaseOperationCost] = FillGasMapBaseOperationCosts(value)
	gasMap[core.BuiltInCost] = FillGasMapBuiltInCosts(value)

	return gasMap
}

func FillGasMapBaseOperationCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["StorePerByte"] = value
	gasMap["DataCopyPerByte"] = value
	gasMap["ReleasePerByte"] = value
	gasMap["PersistPerByte"] = value
	gasMap["CompilePerByte"] = value

	return gasMap
}

func FillGasMapBuiltInCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["ChangeOwnerAddress"] = value

	return gasMap
}

func createMockSmartContractProcessorArguments() ArgsNewSmartContractProcessor {
	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule = FillGasMapInternal(gasSchedule, 1)
	return ArgsNewSmartContractProcessor{
		VmContainer:  &mock.VMContainerMock{},
		ArgsParser:   &mock.ArgumentParserMock{},
		Hasher:       &mock.HasherMock{},
		Marshalizer:  &mock.MarshalizerMock{},
		AccountsDB:   &mock.AccountsStub{},
		TempAccounts: &mock.TemporaryAccountsHandlerMock{},
		AdrConv:      &mock.AddressConverterMock{},
		Coordinator:  mock.NewMultiShardsCoordinatorMock(5),
		ScrForwarder: &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler: &mock.FeeAccumulatorStub{},
		EconomicsFee: &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.0
			},
		},
		TxTypeHandler: &mock.TxTypeHandlerMock{},
		GasHandler: &mock.GasHandlerMock{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasMap: gasSchedule,
	}
}

func TestNewSmartContractProcessorNilVM(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNoVM, err)
}

func TestNewSmartContractProcessorNilArgsParser(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewSmartContractProcessorNilHasher(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.Hasher = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewSmartContractProcessorNilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.Marshalizer = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewSmartContractProcessorNilAccountsDB(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewSmartContractProcessorNilAdrConv(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.AdrConv = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewSmartContractProcessorNilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.Coordinator = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewSmartContractProcessorNilFakeAccountsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.TempAccounts = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilTemporaryAccountsHandler, err)
}

func TestNewSmartContractProcessor_NilIntermediateMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ScrForwarder = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilIntermediateTransactionHandler, err)
}

func TestNewSmartContractProcessor_ErrNilUnsignedTxHandlerMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.TxFeeHandler = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilUnsignedTxHandler, err)
}

func TestNewSmartContractProcessor_ErrErrNilGasHandlerMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.GasHandler = nil
	sc, err := NewSmartContractProcessor(arguments)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilGasHandler, err)
}

func TestNewSmartContractProcessor(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	assert.NotNil(t, sc)
	assert.Nil(t, err)
	assert.False(t, sc.IsInterfaceNil())
}

func TestScProcessor_DeploySmartContractBadParse(t *testing.T) {
	t.Parallel()

	addrConverter := &mock.AddressConverterMock{}
	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.AdrConv = addrConverter
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(addrConverter.AddressLen())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	called := false
	tmpError := errors.New("error")
	argParser.ParseDataCalled = func(data string) error {
		called = true
		return tmpError
	}
	_ = sc.DeploySmartContract(tx, acntSrc)
	assert.True(t, called)
}

func TestScProcessor_DeploySmartContractRunError(t *testing.T) {
	t.Parallel()

	addrConverter := &mock.AddressConverterMock{}
	vmContainer := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.AdrConv = addrConverter
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(addrConverter.AddressLen())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	tmpError := errors.New("error")
	vm := &mock.VMExecutionHandlerStub{}
	called := false
	vm.RunSmartContractCreateCalled = func(input *vmcommon.ContractCreateInput) (output *vmcommon.VMOutput, e error) {
		called = true
		return nil, tmpError
	}

	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	vmArg := []byte("00")
	argParser.GetArgumentsCalled = func() ([][]byte, error) {
		return [][]byte{vmArg}, nil
	}

	_ = sc.DeploySmartContract(tx, acntSrc)
	assert.True(t, called)
}

func TestScProcessor_DeploySmartContractWrongTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	err = sc.DeploySmartContract(tx, acntSrc)
	assert.Equal(t, process.ErrWrongTransaction, err)
}

func TestScProcessor_DeploySmartContract(t *testing.T) {
	t.Parallel()

	addrConverter := &mock.AddressConverterMock{}
	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.AdrConv = addrConverter
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(addrConverter.AddressLen())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	vmArg := []byte("00")
	argParser.GetArgumentsCalled = func() ([][]byte, error) {
		return [][]byte{vmArg}, nil
	}

	err = sc.DeploySmartContract(tx, acntSrc)
	assert.Equal(t, nil, err)
}

func TestScProcessor_ExecuteSmartContractTransactionNilTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	err = sc.ExecuteSmartContractTransaction(nil, acntSrc, acntDst)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_ExecuteSmartContractTransactionNilAccount(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, nil)
	assert.Equal(t, process.ErrNilSCDestAccount, err)

	acntDst.SetCode(nil)
	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.Nil(t, err)

	acntDst = nil
	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.Equal(t, process.ErrNilSCDestAccount, err)
}

func TestScProcessor_ExecuteSmartContractTransactionBadParser(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	acntDst.SetCode([]byte("code"))
	tmpError := errors.New("error")
	called := false
	argParser.ParseDataCalled = func(data string) error {
		called = true
		return tmpError
	}
	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.True(t, called)
	assert.Nil(t, err)
}

func TestScProcessor_ExecuteSmartContractTransactionVMRunError(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	acntDst.SetCode([]byte("code"))
	tmpError := errors.New("error")
	vm := &mock.VMExecutionHandlerStub{}
	called := false
	vm.RunSmartContractCallCalled = func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
		called = true
		return nil, tmpError
	}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.True(t, called)
	assert.Nil(t, err)
}

func TestScProcessor_ExecuteSmartContractTransaction(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(0)
	acntSrc, acntDst := createAccounts(tx)

	accntState.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	acntDst.SetCode([]byte("code"))
	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.Nil(t, err)
}

func TestScProcessor_CreateVMCallInputWrongCode(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	tmpError := errors.New("error")
	argParser.GetFunctionCalled = func() (s string, e error) {
		return "", tmpError
	}
	vmInput, err := sc.CreateVMCallInput(tx)
	assert.Nil(t, vmInput)
	assert.Equal(t, tmpError, err)
}

func TestScProcessor_CreateVMCallInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	vmInput, err := sc.CreateVMCallInput(tx)
	assert.NotNil(t, vmInput)
	assert.Nil(t, err)
}

func TestScProcessor_CreateVMDeployInputBadFunction(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	tmpError := errors.New("error")
	argParser.GetCodeCalled = func() (code []byte, e error) {
		return nil, tmpError
	}
	vmArg := []byte("00")
	argParser.GetArgumentsCalled = func() ([][]byte, error) {
		return [][]byte{vmArg}, nil
	}

	vmInput, vmType, err := sc.CreateVMDeployInput(tx)
	assert.Nil(t, vmInput)
	assert.Equal(t, tmpError, err)
	assert.Nil(t, vmType)
}

func TestScProcessor_CreateVMDeployInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data@0000")
	tx.Value = big.NewInt(45)

	vmArg := []byte("00")
	argParser.GetArgumentsCalled = func() ([][]byte, error) {
		return [][]byte{vmArg}, nil
	}

	vmInput, vmType, err := sc.CreateVMDeployInput(tx)
	require.NotNil(t, vmInput)
	assert.Equal(t, vmcommon.DirectCall, vmInput.CallType)
	assert.True(t, bytes.Equal(vmArg, vmType))
	assert.Nil(t, err)
}

func TestScProcessor_CreateVMDeployInputNotEnoughArguments(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data@0000")
	tx.Value = big.NewInt(45)

	vmInput, vmType, err := sc.CreateVMDeployInput(tx)
	assert.Nil(t, vmInput)
	assert.Nil(t, vmType)
	assert.Equal(t, process.ErrNotEnoughArgumentsToDeploy, err)
}

func TestScProcessor_CreateVMInputWrongArgument(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	tmpError := errors.New("error")
	argParser.GetArgumentsCalled = func() (ints [][]byte, e error) {
		return nil, tmpError
	}
	vmInput, err := sc.CreateVMInput(tx)
	assert.Nil(t, vmInput)
	assert.Equal(t, tmpError, err)
}

func TestScProcessor_CreateVMInputNotEnoughGas(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.EconomicsFee = &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return 1000
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	tx.GasLimit = 100

	vmInput, err := sc.CreateVMInput(tx)
	assert.Nil(t, vmInput)
	assert.Equal(t, process.ErrNotEnoughGas, err)
}

func TestScProcessor_CreateVMInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	vmInput, err := sc.CreateVMInput(tx)
	assert.NotNil(t, vmInput)
	assert.Equal(t, nil, err)
}

func createAccountsAndTransaction() (state.UserAccountHandler, state.UserAccountHandler, *transaction.Transaction) {
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	acntSrc, acntDst := createAccounts(tx)

	return acntSrc.(state.UserAccountHandler), acntDst.(state.UserAccountHandler), tx
}

func TestScProcessor_processVMOutputNilVMOutput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSrc, _, tx := createAccountsAndTransaction()

	_, _, err = sc.processVMOutput(nil, tx, acntSrc, vmcommon.DirectCall)
	assert.Equal(t, process.ErrNilVMOutput, err)
}

func TestScProcessor_processVMOutputNilTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSrc, _, _ := createAccountsAndTransaction()

	vmOutput := &vmcommon.VMOutput{}
	_, _, err = sc.processVMOutput(vmOutput, nil, acntSrc, vmcommon.DirectCall)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_processVMOutputNilSndAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}
	_, _, err = sc.processVMOutput(vmOutput, tx, nil, vmcommon.DirectCall)
	assert.Nil(t, err)
}

func TestScProcessor_processVMOutputNilDstAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSnd, _, tx := createAccountsAndTransaction()

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}

	accntState.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return acntSnd, nil
	}

	tx.Value = big.NewInt(0)
	_, _, err = sc.processVMOutput(vmOutput, tx, acntSnd, vmcommon.DirectCall)
	assert.Nil(t, err)
}

func TestScProcessor_GetAccountFromAddressAccNotFound(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	accountsDB.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return nil, state.ErrAccNotFound
	}

	addrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.AdrConv = addrConv
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.GetAccountFromAddress([]byte("SRC"))
	assert.Nil(t, acc)
	assert.Equal(t, state.ErrAccNotFound, err)
}

func TestScProcessor_GetAccountFromAddrFaildAddressConv(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}
	addrConv.Fail = true

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		getCalled++
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.AdrConv = addrConv
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.GetAccountFromAddress([]byte("DST"))
	assert.Nil(t, acc)
	assert.NotNil(t, err)
	assert.Equal(t, 0, getCalled)
}

func TestScProcessor_GetAccountFromAddrFailedGetExistingAccount(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		getCalled++
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.AdrConv = addrConv
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.GetAccountFromAddress([]byte("DST"))
	assert.Nil(t, acc)
	assert.Equal(t, state.ErrAccNotFound, err)
	assert.Equal(t, 1, getCalled)
}

func TestScProcessor_GetAccountFromAddrAccNotInShard(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		getCalled++
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardCoordinator.SelfId() + 1
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.AdrConv = addrConv
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.GetAccountFromAddress([]byte("DST"))
	assert.Nil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, 0, getCalled)
}

func TestScProcessor_GetAccountFromAddr(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		getCalled++
		acc, _ := accounts.NewUserAccount(addressContainer)
		return acc, nil
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.AdrConv = addrConv
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.GetAccountFromAddress([]byte("DST"))
	assert.NotNil(t, acc)
	assert.Nil(t, err)
	assert.Equal(t, 1, getCalled)
}

func TestScProcessor_DeleteAccountsFailedAtRemove(t *testing.T) {
	t.Parallel()
	addrConv := &mock.AddressConverterMock{}

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return nil, state.ErrAccNotFound
	}
	accountsDB.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		removeCalled++
		return state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.AdrConv = addrConv
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.DeleteAccounts(deletedAccounts)
	assert.Equal(t, state.ErrAccNotFound, err)
	assert.Equal(t, 0, removeCalled)
}

func TestScProcessor_DeleteAccountsNotInShard(t *testing.T) {
	t.Parallel()
	addrConv := &mock.AddressConverterMock{}

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		removeCalled++
		return state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	computeIdCalled := 0
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		computeIdCalled++
		return shardCoordinator.SelfId() + 1
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.AdrConv = addrConv
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.DeleteAccounts(deletedAccounts)
	assert.Nil(t, err)
	assert.Equal(t, 0, removeCalled)
	assert.Equal(t, len(deletedAccounts), computeIdCalled)
}

func TestScProcessor_DeleteAccountsInShard(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}
	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		acc, _ := accounts.NewUserAccount(addressContainer)
		return acc, nil
	}
	accountsDB.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		removeCalled++
		return nil
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	computeIdCalled := 0
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		computeIdCalled++
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.AdrConv = addrConv
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.DeleteAccounts(deletedAccounts)
	assert.Nil(t, err)
	assert.Equal(t, len(deletedAccounts), removeCalled)
	assert.Equal(t, len(deletedAccounts), computeIdCalled)
}

func TestScProcessor_ProcessSCPaymentAccNotInShardShouldNotReturnError(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	err = sc.ProcessSCPayment(tx, nil)
	assert.Nil(t, err)
}

func TestScProcessor_ProcessSCPaymentNotEnoughBalance(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 15

	acntSrc, _ := createAccounts(tx)
	stAcc, _ := acntSrc.(state.UserAccountHandler)
	stAcc.SetBalance(big.NewInt(45))

	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()

	err = sc.ProcessSCPayment(tx, acntSrc)
	assert.Equal(t, process.ErrInsufficientFunds, err)
	assert.Equal(t, currBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
}

func TestScProcessor_ProcessSCPayment(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()
	modifiedBalance := currBalance - tx.Value.Uint64() - tx.GasLimit*tx.GasLimit

	err = sc.ProcessSCPayment(tx, acntSrc)
	assert.Nil(t, err)
	assert.Equal(t, modifiedBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
}

func TestScProcessor_RefundGasToSenderNilAndZeroRefund(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	txHash := []byte("txHash")

	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(0)}
	_, _, err = sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		tx,
		txHash,
		acntSrc,
		vmcommon.DirectCall,
	)
	assert.Nil(t, err)
	assert.Equal(t, currBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
}

func TestScProcessor_RefundGasToSenderAccNotInShard(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10
	txHash := []byte("txHash")
	acntSrc, _ := createAccounts(tx)
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(10)}
	sctx, consumed, err := sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		tx,
		txHash,
		nil,
		vmcommon.DirectCall,
	)
	assert.Nil(t, err)
	assert.NotNil(t, sctx)
	assert.Equal(t, 0, consumed.Cmp(big.NewInt(0).SetUint64(tx.GasPrice*tx.GasLimit)))

	acntSrc = nil
	vmOutput = &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(10)}
	sctx, consumed, err = sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		tx,
		txHash,
		acntSrc,
		vmcommon.DirectCall,
	)
	assert.Nil(t, err)
	assert.NotNil(t, sctx)
	assert.Equal(t, 0, consumed.Cmp(big.NewInt(0).SetUint64(tx.GasPrice*tx.GasLimit)))
}

func TestScProcessor_RefundGasToSender(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(10)
	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &mock.FeeHandlerStub{MinGasPriceCalled: func() uint64 {
		return minGasPrice
	}}
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 15
	tx.GasLimit = 15
	txHash := []byte("txHash")
	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()

	refundGas := big.NewInt(10)
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: refundGas}
	_, _, err = sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		tx,
		txHash,
		acntSrc,
		vmcommon.DirectCall,
	)
	assert.Nil(t, err)

	totalRefund := refundGas.Uint64() * minGasPrice
	assert.Equal(t, currBalance+totalRefund, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
}

func TestScProcessor_processVMOutputNilOutput(t *testing.T) {
	t.Parallel()

	acntSrc, _, tx := createAccountsAndTransaction()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	_, _, err = sc.ProcessVMOutput(nil, tx, acntSrc)

	assert.Equal(t, process.ErrNilVMOutput, err)
}

func TestScProcessor_processVMOutputNilTransaction(t *testing.T) {
	t.Parallel()

	acntSrc, _, _ := createAccountsAndTransaction()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	vmOutput := &vmcommon.VMOutput{}
	_, _, err = sc.ProcessVMOutput(vmOutput, nil, acntSrc)

	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_processVMOutput(t *testing.T) {
	t.Parallel()

	acntSrc, _, tx := createAccountsAndTransaction()

	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}

	accntState.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	tx.Value = big.NewInt(0)
	_, _, err = sc.ProcessVMOutput(vmOutput, tx, acntSrc)
	assert.Nil(t, err)
}

func TestScProcessor_processSCOutputAccounts(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}

	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}
	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	_, err = sc.ProcessSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	assert.Nil(t, err)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Code = []byte("contract-code")
	outacc1.Nonce = 5
	outacc1.BalanceDelta = big.NewInt(int64(5))
	outputAccounts = append(outputAccounts, outacc1)

	testAddr := mock.NewAddressMock(outaddress)
	testAcc, _ := accounts.NewUserAccount(testAddr)

	accountsDB.LoadAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		if bytes.Equal(addressContainer.Bytes(), testAddr.Bytes()) {
			return testAcc, nil
		}
		return nil, state.ErrAccNotFound
	}
	accountsDB.PutCodeCalled = func(accountHandler state.AccountHandler, code []byte) error {
		return nil
	}

	accountsDB.SaveAccountCalled = func(accountHandler state.AccountHandler) error {
		return nil
	}

	tx.Value = big.NewInt(int64(5))
	_, err = sc.ProcessSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	assert.Nil(t, err)

	outacc1.BalanceDelta = nil
	outacc1.Nonce++
	tx.Value = big.NewInt(0)
	_, err = sc.ProcessSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	assert.Nil(t, err)

	outacc1.Nonce++
	outacc1.BalanceDelta = big.NewInt(int64(10))
	tx.Value = big.NewInt(int64(10))
	fakeAccountsHandler.TempAccountCalled = func(address []byte) state.AccountHandler {
		fakeAcc, _ := accounts.NewUserAccount(mock.NewAddressMock(address))
		fakeAcc.Balance = big.NewInt(int64(5))
		return fakeAcc
	}

	currentBalance := testAcc.Balance.Uint64()
	vmOutBalance := outacc1.BalanceDelta.Uint64()
	_, err = sc.ProcessSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	assert.Nil(t, err)
	assert.Equal(t, currentBalance+vmOutBalance, testAcc.Balance.Uint64())
}

func TestScProcessor_processSCOutputAccountsNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}
	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	_, err = sc.ProcessSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	assert.Nil(t, err)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Code = []byte("contract-code")
	outacc1.Nonce = 5
	outputAccounts = append(outputAccounts, outacc1)

	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardCoordinator.SelfId() + 1
	}

	_, err = sc.ProcessSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	assert.Nil(t, err)
}

func TestScProcessor_CreateCrossShardTransactions(t *testing.T) {
	t.Parallel()

	testAccounts, _ := accounts.NewUserAccount(state.NewAddress([]byte("address")))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, err error) {
			return testAccounts, nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Nonce = 0
	outacc1.Balance = big.NewInt(int64(5))
	outacc1.BalanceDelta = big.NewInt(int64(15))
	outputAccounts = append(outputAccounts, outacc1, outacc1, outacc1)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 15
	txHash := []byte("txHash")

	scTxs, err := sc.CreateSCRTransactions(outputAccounts, tx, txHash)
	assert.Nil(t, err)
	assert.Equal(t, len(outputAccounts), len(scTxs))
}

func TestScProcessor_ProcessSmartContractResultNilScr(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	err = sc.ProcessSmartContractResult(nil)
	assert.Equal(t, process.ErrNilSmartContractResult, err)
}

func TestScProcessor_ProcessSmartContractResultErrGetAccount(t *testing.T) {
	t.Parallel()

	accError := errors.New("account get error")
	called := false
	accountsDB := &mock.AccountsStub{LoadAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		called = true
		return nil, accError
	}}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address")}
	err = sc.ProcessSmartContractResult(&scr)
	assert.True(t, called)
}

func TestScProcessor_ProcessSmartContractResultAccNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardCoordinator.CurrentShard + 1
	}
	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address")}
	err = sc.ProcessSmartContractResult(&scr)
	assert.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultBadAccType(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{LoadAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.AccountWrapMock{}, nil
	}}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address")}
	err = sc.ProcessSmartContractResult(&scr)
	assert.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultOutputBalanceNil(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accounts.NewUserAccount(addressContainer)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: []byte("recv address")}
	err = sc.ProcessSmartContractResult(&scr)
	assert.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultWithCode(t *testing.T) {
	t.Parallel()

	putCodeCalled := 0
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accounts.NewUserAccount(addressContainer)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			putCodeCalled++
			return nil
		},
	}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: []byte("recv address"),
		Code:    []byte("code"),
		Value:   big.NewInt(15),
	}
	err = sc.ProcessSmartContractResult(&scr)
	assert.Nil(t, err)
	assert.Equal(t, 1, putCodeCalled)
}

func TestScProcessor_ProcessSmartContractResultWithData(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accounts.NewUserAccount(addressContainer)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	test := "test"
	result := ""
	sep := "@"
	for i := 0; i < 6; i++ {
		result += test
		result += sep
	}

	scr := smartContractResult.SmartContractResult{
		RcvAddr: []byte("recv address"),
		Data:    []byte(result),
		Value:   big.NewInt(15),
	}
	err = sc.ProcessSmartContractResult(&scr)
	assert.Nil(t, err)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestScProcessor_ProcessSmartContractResultDeploySCShouldError(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accounts.NewUserAccount(addressContainer)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	arguments.TxTypeHandler = &mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, e error) {
			return process.SCDeployment, nil
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: []byte("recv address"),
		Data:    []byte("code@06"),
		Value:   big.NewInt(15),
	}
	err = sc.ProcessSmartContractResult(&scr)
	assert.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultExecuteSC(t *testing.T) {
	t.Parallel()

	scAddress := []byte("000000000001234567890123456789012")
	dstScAddress, _ := accounts.NewUserAccount(mock.NewAddressMock(scAddress))
	dstScAddress.SetCode([]byte("code"))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return dstScAddress, nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	}
	fakeAccountsHandler := &mock.TemporaryAccountsHandlerMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	executeCalled := false
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.TempAccounts = fakeAccountsHandler
	arguments.Coordinator = shardCoordinator
	arguments.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return &mock.VMExecutionHandlerStub{
				RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
					executeCalled = true
					return nil, nil
				},
			}, nil
		},
	}
	arguments.TxTypeHandler = &mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, e error) {
			return process.SCInvoking, nil
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: []byte("recv address"),
		Data:    []byte("code@06"),
		Value:   big.NewInt(15),
	}
	err = sc.ProcessSmartContractResult(&scr)
	assert.Nil(t, err)
	assert.True(t, executeCalled)
}
