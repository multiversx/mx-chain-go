package smartContract

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func createAccounts(tx *transaction.Transaction) (state.AccountHandler, state.AccountHandler) {
	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	acntSrc, _ := state.NewAccount(mock.NewAddressMock(tx.SndAddr), tracker)
	acntDst, _ := state.NewAccount(mock.NewAddressMock(tx.RcvAddr), tracker)

	return acntSrc, acntDst
}

func TestNewSmartContractProcessorNilVM(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		nil,
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNoVM, err)
}

func TestNewSmartContractProcessorNilArgsParser(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		nil,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewSmartContractProcessorNilHasher(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		nil,
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewSmartContractProcessorNilMarshalizer(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		nil,
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewSmartContractProcessorNilAccountsDB(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewSmartContractProcessorNilAdrConv(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		nil,
		mock.NewMultiShardsCoordinatorMock(5))

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewSmartContractProcessorNilShardCoordinator(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		nil)

	assert.Nil(t, sc)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewSmartContractProcessor(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.NotNil(t, sc)
	assert.Nil(t, err)
}

func TestScProcessor_ComputeTransactionTypeNil(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	_, err = sc.ComputeTransactionType(nil, nil, nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_ComputeTransactionTypeNilTx(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	acntSrc, acntDst := createAccounts(tx)

	tx = nil
	_, err = sc.ComputeTransactionType(tx, acntSrc, acntDst)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_ComputeTransactionTypeErrWrongTransaction(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = nil
	tx.Value = big.NewInt(45)

	acntSrc, acntDst := createAccounts(tx)

	_, err = sc.ComputeTransactionType(tx, acntSrc, acntDst)
	assert.Equal(t, process.ErrWrongTransaction, err)
}

func TestScProcessor_ComputeTransactionTypeScDeployment(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = nil
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	acntSrc, acntDst := createAccounts(tx)

	txType, err := sc.ComputeTransactionType(tx, acntSrc, acntDst)
	assert.Nil(t, err)
	assert.Equal(t, process.SCDeployment, txType)
}

func TestScProcessor_ComputeTransactionTypeScInvoking(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

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

	txType, err := sc.ComputeTransactionType(tx, acntSrc, acntDst)
	assert.Nil(t, err)
	assert.Equal(t, process.SCInvoking, txType)
}

func TestScProcessor_ComputeTransactionTypeMoveBalance(t *testing.T) {
	t.Parallel()

	sc, err := NewSmartContractProcessor(
		&mock.VMExecutionHandlerStub{},
		&mock.AtArgumentParserMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))

	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	acntSrc, acntDst := createAccounts(tx)

	txType, err := sc.ComputeTransactionType(tx, acntSrc, acntDst)
	assert.Nil(t, err)
	assert.Equal(t, process.MoveBalance, txType)
}

func TestScProcessor_DeploySmartContractBadParse(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = nil
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	tmpError := errors.New("error")
	argParser.ParseDataCalled = func(data []byte) error {
		return tmpError
	}
	err = sc.DeploySmartContract(tx, acntSrc, acntDst)
	assert.Equal(t, tmpError, err)
}

func TestScProcessor_DeploySmartContractRunError(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = nil
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	tmpError := errors.New("error")
	vm.RunSmartContractCreateCalled = func(input *vmcommon.ContractCreateInput) (output *vmcommon.VMOutput, e error) {
		return nil, tmpError
	}
	err = sc.DeploySmartContract(tx, acntSrc, acntDst)
	assert.Equal(t, tmpError, err)
}

func TestScProcessor_DeploySmartContractWrongTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	err = sc.DeploySmartContract(tx, acntSrc, acntDst)
	assert.Equal(t, process.ErrWrongTransaction, err)
}

func TestScProcessor_DeploySmartContract(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = nil
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	err = sc.DeploySmartContract(tx, acntSrc, acntDst)
	assert.Equal(t, nil, err)
}

func TestScProcessor_ExecuteSmartContractTransactionNilTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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
	assert.Equal(t, process.ErrNilSCDestAccount, err)

	acntDst = nil
	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.Equal(t, process.ErrNilSCDestAccount, err)
}

func TestScProcessor_ExecuteSmartContractTransactionBadParser(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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
	argParser.ParseDataCalled = func(data []byte) error {
		return tmpError
	}
	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.Equal(t, tmpError, err)
}

func TestScProcessor_ExecuteSmartContractTransactionVMRunError(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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
	vm.RunSmartContractCallCalled = func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
		return nil, tmpError
	}
	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.Equal(t, tmpError, err)
}

func TestScProcessor_ExecuteSmartContractTransaction(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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
	err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	assert.Equal(t, nil, err)
}

func TestScProcessor_CreateVMCallInputWrongCode(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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

	vmInput, err := sc.CreateVMDeployInput(tx)
	assert.Nil(t, vmInput)
	assert.Equal(t, tmpError, err)
}

func TestScProcessor_CreateVMDeployInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	vmInput, err := sc.CreateVMDeployInput(tx)
	assert.NotNil(t, vmInput)
	assert.Nil(t, err)
}

func TestScProcessor_CreateVMInputWrongArgument(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	tmpError := errors.New("error")
	argParser.GetArgumentsCalled = func() (ints []*big.Int, e error) {
		return nil, tmpError
	}
	vmInput, err := sc.CreateVMInput(tx)
	assert.Nil(t, vmInput)
	assert.Equal(t, tmpError, err)
}

func TestScProcessor_CreateVMInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
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

//------- moveBalances
func TestScProcessor_MoveBalancesShouldNotFailWhenAcntSrcIsNotInNodeShard(t *testing.T) {
	adrDst := mock.NewAddressMock([]byte{67})
	journalizeCalled := false
	saveAccountCalled := false
	acntDst, _ := state.NewAccount(adrDst, &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled = true
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled = true
			return nil
		},
	})

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	err = sc.MoveBalances(nil, acntDst, big.NewInt(0))

	assert.True(t, journalizeCalled && saveAccountCalled)
	assert.Nil(t, err)
}

func TestScProcessor_MoveBalancesShouldNotFailWhenAcntDstIsNotInNodeShard(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	journalizeCalled := false
	saveAccountCalled := false
	acntSrc, _ := state.NewAccount(adrSrc, &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled = true
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled = true
			return nil
		},
	})

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)
	err = sc.MoveBalances(acntSrc, nil, big.NewInt(0))

	assert.True(t, journalizeCalled && saveAccountCalled)
	assert.Nil(t, err)
}

func TestTxProcessor_MoveBalancesOkValsShouldWork(t *testing.T) {
	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc, err := state.NewAccount(adrSrc, tracker)
	assert.Nil(t, err)

	adrDst := mock.NewAddressMock([]byte{67})
	acntDst, err := state.NewAccount(adrDst, tracker)
	assert.Nil(t, err)

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSrc.Balance = big.NewInt(64)
	acntDst.Balance = big.NewInt(31)
	err = sc.MoveBalances(acntSrc, acntDst, big.NewInt(14))

	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(50), acntSrc.Balance)
	assert.Equal(t, big.NewInt(45), acntDst.Balance)
	assert.Equal(t, 2, journalizeCalled)
	assert.Equal(t, 2, saveAccountCalled)
}

func TestScProcessor_MoveBalancesToSelfOkValsShouldWork(t *testing.T) {
	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc, err := state.NewAccount(adrSrc, tracker)
	assert.Nil(t, err)

	acntDst := acntSrc

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSrc.Balance = big.NewInt(64)

	err = sc.MoveBalances(acntSrc, acntDst, big.NewInt(1))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(64), acntSrc.Balance)
	assert.Equal(t, big.NewInt(64), acntDst.Balance)
	assert.Equal(t, 2, journalizeCalled)
	assert.Equal(t, 2, saveAccountCalled)
}

//------- increaseNonce

func TestTxProcessor_IncreaseNonceOkValsShouldWork(t *testing.T) {
	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc, err := state.NewAccount(adrSrc, tracker)
	assert.Nil(t, err)

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSrc.Nonce = 45

	err = sc.IncreaseNonce(acntSrc)
	assert.Nil(t, err)
	assert.Equal(t, uint64(46), acntSrc.Nonce)
	assert.Equal(t, 1, journalizeCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func createAccountsAndTransaction() (*state.Account, *state.Account, *transaction.Transaction) {
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	adrSrc := mock.NewAddressMock([]byte("SRC"))
	acntSrc, _ := state.NewAccount(adrSrc, tracker)

	adrDst := mock.NewAddressMock([]byte("DST"))
	acntDst, _ := state.NewAccount(adrDst, tracker)

	return acntSrc, acntDst, tx
}

func TestScProcessor_processVMOutputNilVMOutput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSrc, acntDst, tx := createAccountsAndTransaction()

	err = sc.processVMOutput(nil, tx, acntSrc, acntDst)
	assert.Equal(t, process.ErrNilVMOutput, err)
}

func TestScProcessor_processVMOutputNilTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSrc, acntDst, _ := createAccountsAndTransaction()

	vmOutput := &vmcommon.VMOutput{}
	err = sc.processVMOutput(vmOutput, nil, acntSrc, acntDst)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_processVMOutputNilSndAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	_, acntDst, tx := createAccountsAndTransaction()

	vmOutput := &vmcommon.VMOutput{}
	err = sc.processVMOutput(vmOutput, tx, nil, acntDst)
	assert.Nil(t, err)
}

func TestScProcessor_processVMOutputNilDstAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(5))
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acntSnd, _, tx := createAccountsAndTransaction()

	vmOutput := &vmcommon.VMOutput{}
	err = sc.processVMOutput(vmOutput, tx, acntSnd, nil)
	assert.Nil(t, err)
}

func TestScProcessor_GetAccountFromAddress(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	accountsDB.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return nil, state.ErrAccNotFound
	}

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		accountsDB,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock())
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("SRC"))
	assert.Nil(t, acc)
	assert.Equal(t, state.ErrAccNotFound, err)
}

func TestScProcessor_GetAccountFromAddrFaildAddressConv(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}
	addrConv.Fail = true

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		removeCalled++
		return nil
	}

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		accountsDB,
		addrConv,
		mock.NewOneShardCoordinatorMock())
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	assert.Nil(t, acc)
	assert.NotNil(t, err)
	assert.Equal(t, 0, removeCalled)
}

func TestScProcessor_GetAccountFromAddrFaildGetExistingAccount(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}
	addrConv.Fail = true

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		removeCalled++
		return nil
	}

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		accountsDB,
		addrConv,
		mock.NewOneShardCoordinatorMock())
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	assert.Nil(t, acc)
	assert.NotNil(t, err)
	assert.Equal(t, 0, removeCalled)
}

func TestScProcessor_GetAccountFromAddrAccNotInShard(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}
	addrConv.Fail = true

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		removeCalled++
		return nil
	}

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		accountsDB,
		addrConv,
		mock.NewOneShardCoordinatorMock())
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	assert.Nil(t, acc)
	assert.NotNil(t, err)
	assert.Equal(t, 0, removeCalled)
}

func TestScProcessor_DeleteAccountsFaildAtRemove(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}

	accountsDB := &mock.AccountsStub{}
	accountsDB.RemoveAccountCalled = func(addressContainer state.AddressContainer) error {
		return
	}

	vm := &mock.VMExecutionHandlerStub{}
	argParser := &mock.AtArgumentParserMock{}
	sc, err := NewSmartContractProcessor(
		vm,
		argParser,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		accountsDB,
		addrConv,
		mock.NewOneShardCoordinatorMock())
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	assert.Nil(t, acc)
	assert.NotNil(t, err)
	assert.Equal(t, 0, removeCalled)
}
