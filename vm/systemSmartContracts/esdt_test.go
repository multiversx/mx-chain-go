package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArgumentsForESDT() ArgsNewESDTSmartContract {
	return ArgsNewESDTSmartContract{
		Eei:     &mock.SystemEIStub{},
		GasCost: vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{ESDTIssue: 10}},
		ESDTSCConfig: config.ESDTSystemSCConfig{
			BaseIssuingCost: "1000",
		},
		ESDTSCAddress: []byte("address"),
		Marshalizer:   &mock.MarshalizerMock{},
		Hasher:        &mock.HasherMock{},
		EpochNotifier: &mock.EpochNotifierStub{},
	}
}

func TestNewESDTSmartContract(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	e, err := NewESDTSmartContract(args)
	ky := hex.EncodeToString([]byte("ELRONDesdttxgenESDTtkn"))
	fmt.Println(ky)

	assert.Nil(t, err)
	assert.NotNil(t, e)
}

func TestNewESDTSmartContract_NilEEIShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.Eei = nil

	e, err := NewESDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewESDTSmartContract_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.Marshalizer = nil

	e, err := NewESDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilMarshalizer, err)
}

func TestNewESDTSmartContract_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.Hasher = nil

	e, err := NewESDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilHasher, err)
}

func TestNewESDTSmartContract_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.EpochNotifier = nil

	e, err := NewESDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilEpochNotifier, err)
}

func TestNewESDTSmartContract_BaseIssuingCostLessThanZeroShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.ESDTSCConfig.BaseIssuingCost = "-1"

	e, err := NewESDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrInvalidBaseIssuingCost, err)
}

func TestNewESDTSmartContract_InvalidBaseIssuingCostShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.ESDTSCConfig.BaseIssuingCost = "invalid cost"

	e, err := NewESDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrInvalidBaseIssuingCost, err)
}

func TestEsdt_ExecuteIssue(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("addr"),
			CallValue:   big.NewInt(0),
			GasProvided: 100000,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	eei.gasRemaining = vmInput.GasProvided
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TICKER")}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(100).Bytes())
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	tokenID := eei.output[0]

	vmInput.Arguments[0] = []byte("01234567891&*@")
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("getAllESDTTokens", [][]byte{})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, tokenID, eei.output[0])
}

func TestEsdt_ExecuteNilArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	e, _ := NewESDTSmartContract(args)

	output := e.Execute(nil)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteInit(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	e, _ := NewESDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      nil,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "_init",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestEsdt_ExecuteWrongFunctionCall(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.ESDTSCConfig.OwnerAddress = "owner"
	e, _ := NewESDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      nil,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "wrong function",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionNotFound, output)
}

func TestEsdt_ExecuteBurnWrongNumOfArgsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{[]byte("esdtToken"), {100}})
	vmInput.Arguments = [][]byte{[]byte("wrong_token_name")}

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "number of arguments must be equal with 2"))
}

func TestEsdt_ExecuteBurnWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{[]byte("esdtToken"), {100}})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecuteBurnWrongValueToBurnShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{[]byte("esdtToken"), {100}})
	vmInput.Arguments[1] = []byte{0}

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "negative or 0 value to burn"))
}

func TestEsdt_ExecuteBurnOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{[]byte("esdtToken"), {100}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteBurnOnNonBurnableTokenShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		Burnable: false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{tokenName, {100}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "token is not burnable"))
}

func TestEsdt_ExecuteBurn(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:  tokenName,
		Burnable:   true,
		BurntValue: big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{[]byte("esdtToken"), {100}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	esdtData := &ESDTData{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, big.NewInt(200), esdtData.BurntValue)
}

func TestEsdt_ExecuteMintTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "accepted arguments number 2/3"))
}

func TestEsdt_ExecuteMintTooManyArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("esdtToken"), {200}, []byte("dest"), []byte("arg")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "accepted arguments number 2/3"))
}

func TestEsdt_ExecuteMintWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("esdtToken"), {200}})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecuteMintNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("esdtToken"), {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_ExecuteMintOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("esdtToken"), {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteMintNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestEsdt_ExecuteMintWrongMintValueShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("owner"),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {0}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "negative or zero mint value"))
}

func TestEsdt_ExecuteMintNonMintableTokenShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("owner"),
		Mintable:     false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "token is not mintable"))
}

func TestEsdt_ExecuteMintSavesTokenWithMintedTokensAdded(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    []byte("esdtToken"),
		OwnerAddress: []byte("owner"),
		Mintable:     true,
		MintedValue:  big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {200}})

	_ = e.Execute(vmInput)

	esdtData := &ESDTData{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, big.NewInt(300), esdtData.MintedValue)
}

func TestEsdt_ExecuteMintInvalidDestinationAddressShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: []byte("owner"),
		Mintable:     true,
		MintedValue:  big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, {200}, []byte("dest")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "destination address of invalid length"))
}

func TestEsdt_ExecuteMintTransferFailsShouldErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForESDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
			OwnerAddress: []byte("owner"),
			Mintable:     true,
			MintedValue:  big.NewInt(100),
		})
		return marshalizedData
	}
	args.Eei.(*mock.SystemEIStub).TransferCalled = func(destination []byte, sender []byte, value *big.Int, input []byte) error {
		return err
	}
	args.Eei.(*mock.SystemEIStub).AddReturnMessageCalled = func(msg string) {
		assert.Equal(t, err.Error(), msg)
	}

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{[]byte("esdtToken"), {200}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteMintWithTwoArgsShouldSetOwnerAsDestination(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	mintValue := []byte{200}
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: owner,
		Mintable:     true,
		MintedValue:  big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, mintValue})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(owner)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(mintValue)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmcommon.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteMintWithThreeArgsShouldSetThirdArgAsDestination(t *testing.T) {
	t.Parallel()

	dest := []byte("_dest")
	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	mintValue := []byte{200}
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: owner,
		Mintable:     true,
		MintedValue:  big.NewInt(100),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("mint", [][]byte{tokenName, mintValue, dest})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(dest)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(mintValue)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmcommon.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteIssueDisabled(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.ESDTSCConfig.EnabledEpoch = 1
	e, _ := NewESDTSmartContract(args)

	callValue, _ := big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      [][]byte{[]byte("01234567891")},
			CallValue:      callValue,
			CallType:       0,
			GasPrice:       0,
			GasProvided:    args.GasCost.MetaChainSystemSCsCost.ESDTIssue,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteToggleFreezeTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 2"))
}

func TestEsdt_ExecuteToggleFreezeWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken"), []byte("owner")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecuteToggleFreezeNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_ExecuteToggleFreezeOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteToggleFreezeNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := "esdtToken"
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[tokenName] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte(tokenName), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestEsdt_ExecuteToggleFreezeNonFreezableTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: owner,
		CanFreeze:    false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, owner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot freeze"))
}

func TestEsdt_ExecuteToggleFreezeTransferFailsShouldErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForESDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
			OwnerAddress: []byte("owner"),
			CanFreeze:    true,
		})
		return marshalizedData
	}
	args.Eei.(*mock.SystemEIStub).TransferCalled = func(destination []byte, sender []byte, value *big.Int, input []byte) error {
		return err
	}
	args.Eei.(*mock.SystemEIStub).AddReturnMessageCalled = func(msg string) {
		assert.Equal(t, err.Error(), msg)
	}

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteToggleFreezeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, owner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(owner)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTFreeze + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmcommon.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteToggleUnFreezeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unFreeze", [][]byte{tokenName, owner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(owner)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTUnFreeze + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmcommon.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteWipeTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 2"))
}

func TestEsdt_ExecuteWipeWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken"), []byte("owner")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecuteWipeNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_ExecuteWipeOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteWipeNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := "esdtToken"
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[tokenName] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte(tokenName), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestEsdt_ExecuteWipeNonWipeableTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: owner,
		CanWipe:      false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{tokenName, owner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot wipe"))
}

func TestEsdt_ExecuteWipeInvalidDestShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: owner,
		CanWipe:      true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{tokenName, []byte("dest")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid arguments"))
}

func TestEsdt_ExecuteWipeTransferFailsShouldErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForESDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
			OwnerAddress: []byte("owner"),
			CanWipe:      true,
		})
		return marshalizedData
	}
	args.Eei.(*mock.SystemEIStub).TransferCalled = func(destination []byte, sender []byte, value *big.Int, input []byte) error {
		return err
	}
	args.Eei.(*mock.SystemEIStub).AddReturnMessageCalled = func(msg string) {
		assert.Equal(t, err.Error(), msg)
	}

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteWipeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanWipe:      true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{tokenName, owner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(owner)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTWipe + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmcommon.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecutePauseTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 1"))
}

func TestEsdt_ExecutePauseWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{[]byte("esdtToken")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecutePauseNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_ExecutePauseOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecutePauseNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := "esdtToken"
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[tokenName] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{[]byte(tokenName)})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestEsdt_ExecutePauseNonPauseableTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: owner,
		CanPause:     false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot pause/un-pause"))
}

func TestEsdt_ExecutePauseOnAPausedTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot pause an already paused contract"))
}

func TestEsdt_ExecuteTogglePauseSavesTokenWithPausedFlagSet(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: []byte("owner"),
		CanPause:     true,
		IsPaused:     false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{tokenName})

	_ = e.Execute(vmInput)

	esdtData := &ESDTData{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, true, esdtData.IsPaused)
}

func TestEsdt_ExecuteTogglePauseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanPause:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("pause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()

	systemAddress := make([]byte, len(core.SystemAccountAddress))
	copy(systemAddress, core.SystemAccountAddress)
	systemAddress[len(core.SystemAccountAddress)-1] = 0

	createdAcc, accCreated := vmOutput.OutputAccounts[string(systemAddress)]
	assert.True(t, accCreated)

	assert.True(t, len(createdAcc.OutputTransfers) == 1)
	outputTransfer := createdAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	expectedInput := core.BuiltInFunctionESDTPause + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmcommon.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteUnPauseOnAnUnPausedTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unPause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot unPause an already un-paused contract"))
}

func TestEsdt_ExecuteUnPauseSavesTokenWithPausedFlagSetToFalse(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: []byte("owner"),
		CanPause:     true,
		IsPaused:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unPause", [][]byte{tokenName})

	_ = e.Execute(vmInput)

	esdtData := &ESDTData{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, false, esdtData.IsPaused)
}

func TestEsdt_ExecuteUnPauseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unPause", [][]byte{tokenName})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()

	systemAddress := make([]byte, len(core.SystemAccountAddress))
	copy(systemAddress, core.SystemAccountAddress)
	systemAddress[len(core.SystemAccountAddress)-1] = 0

	createdAcc, accCreated := vmOutput.OutputAccounts[string(systemAddress)]
	assert.True(t, accCreated)

	assert.True(t, len(createdAcc.OutputTransfers) == 1)
	outputTransfer := createdAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	expectedInput := core.BuiltInFunctionESDTUnPause + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmcommon.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteTransferOwnershipWrongNumOfArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "expected num of arguments 2"))
}

func TestEsdt_ExecuteTransferOwnershipWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken"), []byte("newOwner")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecuteTransferOwnershipNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_ExecuteTransferOwnershipOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteTransferOwnershipNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestEsdt_ExecuteTransferOwnershipNonTransferableTokenShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress:   []byte("owner"),
		CanChangeOwner: false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot change owner of the token"))
}

func TestEsdt_ExecuteTransferOwnershipInvalidDestinationAddressShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:      tokenName,
		OwnerAddress:   []byte("owner"),
		CanChangeOwner: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken"), []byte("newOwner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "destination address of invalid length"))
}

func TestEsdt_ExecuteTransferOwnershipSavesTokenWithNewOwnerAddressSet(t *testing.T) {
	t.Parallel()

	newOwner := []byte("12345")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:      []byte("esdtToken"),
		OwnerAddress:   []byte("owner"),
		CanChangeOwner: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken"), newOwner})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	esdtData := &ESDTData{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, newOwner, esdtData.OwnerAddress)
}

func TestEsdt_ExecuteEsdtControlChangesWrongNumOfArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))
}

func TestEsdt_ExecuteEsdtControlChangesWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken"), []byte("burnable")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecuteEsdtControlChangesNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken"), []byte("burnable")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_ExecuteEsdtControlChangesOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken"), []byte("burnable")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteEsdtControlChangesNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("random address"),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken"), []byte("burnable")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestEsdt_ExecuteEsdtControlChangesNonUpgradableTokenShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		OwnerAddress: []byte("owner"),
		Upgradable:   false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken"), []byte("burnable")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "token is not upgradable"))
}

func TestEsdt_ExecuteEsdtControlChangesSavesTokenWithUpgradedPropreties(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTData{
		TokenName:    []byte("esdtToken"),
		OwnerAddress: []byte("owner"),
		Upgradable:   true,
		BurntValue:   big.NewInt(100),
		MintedValue:  big.NewInt(1000),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken"), []byte(burnable)})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput = getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken"),
		[]byte(burnable), []byte("true"),
		[]byte(mintable), []byte("true"),
		[]byte(canPause), []byte("true"),
		[]byte(canFreeze), []byte("true"),
		[]byte(canWipe), []byte("true"),
		[]byte(upgradable), []byte("false"),
		[]byte(canChangeOwner), []byte("true"),
	})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	esdtData := &ESDTData{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.True(t, esdtData.Burnable)
	assert.True(t, esdtData.Mintable)
	assert.True(t, esdtData.CanPause)
	assert.True(t, esdtData.CanFreeze)
	assert.True(t, esdtData.CanWipe)
	assert.False(t, esdtData.Upgradable)
	assert.True(t, esdtData.CanChangeOwner)

	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("getTokenProperties", [][]byte{[]byte("esdtToken")})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 12, len(eei.output))
	assert.Equal(t, []byte("esdtToken"), eei.output[0])
	assert.Equal(t, vmInput.CallerAddr, eei.output[1])
}

func TestEsdt_ExecuteConfigChange(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("configChange", [][]byte{[]byte("esdtToken"), []byte(burnable)})
	vmInput.CallerAddr = e.ownerAddress
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	newOwner := vmInput.RecipientAddr
	vmInput = getDefaultVmInputForFunc("configChange",
		[][]byte{newOwner, big.NewInt(100).Bytes(), big.NewInt(5).Bytes(), big.NewInt(20).Bytes()})
	vmInput.CallerAddr = e.ownerAddress
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	esdtData := &ESDTConfig{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage([]byte(configKeyPrefix)))
	assert.True(t, esdtData.BaseIssuingCost.Cmp(big.NewInt(100)) == 0)
	assert.Equal(t, esdtData.MaxTokenNameLength, uint32(20))
	assert.Equal(t, esdtData.MinTokenNameLength, uint32(5))
	assert.True(t, bytes.Equal(newOwner, esdtData.OwnerAddress))
}

func TestEsdt_ExecuteClaim(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("claim", [][]byte{})
	vmInput.CallerAddr = e.ownerAddress

	eei.outputAccounts[string(vmInput.RecipientAddr)] = &vmcommon.OutputAccount{
		Address:      vmInput.RecipientAddr,
		Nonce:        0,
		BalanceDelta: big.NewInt(0),
		Balance:      big.NewInt(100),
	}

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	scOutAcc := eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, scOutAcc.BalanceDelta.Cmp(big.NewInt(-100)) == 0)

	receiver := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.True(t, receiver.BalanceDelta.Cmp(big.NewInt(100)) == 0)
}
