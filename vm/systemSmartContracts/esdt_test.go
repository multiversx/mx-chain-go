package systemSmartContracts

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgumentsForESDT() ArgsNewESDTSmartContract {
	return ArgsNewESDTSmartContract{
		Eei:     &mock.SystemEIStub{},
		GasCost: vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{ESDTIssue: 10}},
		ESDTSCConfig: config.ESDTSystemSCConfig{
			BaseIssuingCost: "1000",
		},
		ESDTSCAddress:          []byte("address"),
		Marshalizer:            &mock.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		AddressPubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		EndOfEpochSCAddress:    vm.EndOfEpochAddress,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTFlagEnabledField:                          true,
			IsGlobalMintBurnFlagEnabledField:                true,
			IsMetaESDTSetFlagEnabledField:                   true,
			IsESDTRegisterAndSetAllRolesFlagEnabledField:    true,
			IsESDTNFTCreateOnMultiShardFlagEnabledField:     true,
			IsESDTTransferRoleFlagEnabledField:              true,
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
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

func TestNewESDTSmartContract_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.EnableEpochsHandler = nil

	e, err := NewESDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilEnableEpochsHandler, err)
}

func TestNewESDTSmartContract_NilPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	args.AddressPubKeyConverter = nil

	e, err := NewESDTSmartContract(args)
	assert.Nil(t, e)
	assert.Equal(t, vm.ErrNilAddressPubKeyConverter, err)
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

func TestEsdt_ExecuteIssueAlways6charactersForRandom(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("addr"),
			CallValue:   big.NewInt(0),
			GasProvided: 100000,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issueNonFungible",
	}
	eei.gasRemaining = vmInput.GasProvided
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue
	ticker := []byte("TICKER")
	vmInput.Arguments = [][]byte{[]byte("name"), ticker}

	randomWithPreprendedZeros := make([]byte, 32)
	randomWithPreprendedZeros[2] = 1
	e.hasher = &mock.HasherStub{
		ComputeCalled: func(s string) []byte {
			return randomWithPreprendedZeros
		},
	}

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	lastOutput := eei.output[len(eei.output)-1]
	assert.Equal(t, len(lastOutput), len(ticker)+1+6)

	vmInput.Function = "issueSemiFungible"
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	lastOutput = eei.output[len(eei.output)-1]
	assert.Equal(t, len(lastOutput), len(ticker)+1+6)

	vmInput.Arguments = nil
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteIssueWithMultiNFTCreate(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
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
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue
	ticker := []byte("TICKER")
	vmInput.Arguments = [][]byte{[]byte("name"), ticker, []byte(canCreateMultiShard), []byte("true")}

	enableEpochsHandler.IsESDTNFTCreateOnMultiShardFlagEnabledField = false
	returnCode := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)

	enableEpochsHandler.IsESDTNFTCreateOnMultiShardFlagEnabledField = true
	returnCode = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)

	vmInput.Function = "issueSemiFungible"
	returnCode = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)

	lastOutput := eei.output[len(eei.output)-1]
	token, _ := e.getExistingToken(lastOutput)
	assert.True(t, token.CanCreateMultiShard)
}

func TestEsdt_ExecuteIssue(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
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
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(10).Bytes())
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmInput.Arguments[0] = []byte("01234567891&*@")
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteIssueWithZero(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
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
	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TICKER")}
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(0).Bytes())
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(10).Bytes())
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue

	enableEpochsHandler.IsGlobalMintBurnFlagEnabledField = false
	enableEpochsHandler.IsESDTNFTCreateOnMultiShardFlagEnabledField = false
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestEsdt_ExecuteIssueTooMuchSupply(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
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

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TICKER")}
	tooMuchToIssue := make([]byte, 101)
	tooMuchToIssue[0] = 1
	vmInput.Arguments = append(vmInput.Arguments, tooMuchToIssue)
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(10).Bytes())
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_IssueInvalidNumberOfDecimals(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
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
	vmInput.Arguments = append(vmInput.Arguments, big.NewInt(25).Bytes())
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{[]byte("esdtToken"), {100}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteBurnAndMintDisabled(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.IsGlobalMintBurnFlagEnabledField = false
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{[]byte("esdtToken"), {100}})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "global burn is no more enabled, use local burn"))

	vmInput = getDefaultVmInputForFunc("mint", [][]byte{[]byte("esdtToken"), {100}})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "global mint is no more enabled, use local mint"))
}

func TestEsdt_ExecuteBurnOnNonBurnableTokenShouldWorkAndReturnBurntTokens(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		Burnable: false,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	burnValue := []byte{100}
	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc(core.BuiltInFunctionESDTBurn, [][]byte{tokenName, burnValue})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.True(t, strings.Contains(eei.returnMessage, "token is not burnable"))

	outputTransfer := eei.outputAccounts["owner"].OutputTransfers[0]
	expectedReturnData := []byte(core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(burnValue))
	assert.Equal(t, expectedReturnData, outputTransfer.Data)
}

func TestEsdt_ExecuteBurn(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	esdtData := &ESDTDataV2{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, big.NewInt(200), esdtData.BurntValue)
}

func TestEsdt_ExecuteMintTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	esdtData := &ESDTDataV2{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, big.NewInt(300), esdtData.MintedValue)

	vmInput.Arguments[1] = make([]byte, 101)
	returnCode := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestEsdt_ExecuteMintInvalidDestinationAddressShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
		marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteMintWithThreeArgsShouldSetThirdArgAsDestination(t *testing.T) {
	t.Parallel()

	dest := []byte("_dest")
	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	mintValue := []byte{200}
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteIssueDisabled(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.IsESDTFlagEnabledField = false
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
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 2"))

	vmInput.Function = "freezeSingleNFT"
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 3"))
}

func TestEsdt_ExecuteToggleFreezeWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken"), []byte("owner")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("owner"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecuteToggleFreezeNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("owner"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_ExecuteToggleFreezeOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("owner"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteToggleFreezeNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := "esdtToken"
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("owner"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestEsdt_ExecuteToggleFreezeNonFreezableTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, owner)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot freeze"))
}

func TestEsdt_ExecuteToggleFreezeTransferFailsShouldErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForESDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{[]byte("esdtToken"), getAddress()})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteToggleFreezeSingleNFTTransferFailsShouldErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForESDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
			OwnerAddress: []byte("owner"),
			CanFreeze:    true,
			TokenType:    []byte(core.NonFungibleESDT),
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
	vmInput := getDefaultVmInputForFunc("freezeSingleNFT", [][]byte{[]byte("esdtToken"), big.NewInt(10).Bytes(), getAddress()})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteToggleFreezeShouldWorkWithRealBech32Address(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToFreezeBech32 := "erd158tgst07d6rt93td6nh5cd2mmpfhtp7hr24l4wfgtlggqpnp6kjsnpvdqj"
	addressToFreeze, err := args.AddressPubKeyConverter.Decode(addressToFreezeBech32)
	assert.NoError(t, err)

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, addressToFreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToFreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTFreeze + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteToggleFreezeShouldFailWithBech32Converter(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToFreeze := []byte("not a bech32 address")

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, addressToFreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid address to freeze/unfreeze"))

	vmInput.Function = "freezeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, addressToFreeze)
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid address to freeze/unfreeze"))
}

func TestEsdt_ExecuteToggleFreezeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToFreeze := getAddress()

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freeze", [][]byte{tokenName, addressToFreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToFreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTFreeze + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteToggleUnFreezeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToUnfreeze := getAddress()

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unFreeze", [][]byte{tokenName, addressToUnfreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToUnfreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTUnFreeze + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteToggleFreezeSingleNFTShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
		TokenType:    []byte(core.NonFungibleESDT),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToFreeze := getAddress()

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("freezeSingleNFT", [][]byte{tokenName, big.NewInt(10).Bytes(), addressToFreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToFreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTFreeze + "@" + hex.EncodeToString(append(tokenName, big.NewInt(10).Bytes()...))
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteToggleUnFreezeSingleNFTShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanFreeze:    true,
		TokenType:    []byte(core.NonFungibleESDT),
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	addressToUnfreeze := getAddress()

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unFreezeSingleNFT", [][]byte{tokenName, big.NewInt(10).Bytes(), addressToUnfreeze})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToUnfreeze)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTUnFreeze + "@" + hex.EncodeToString(append(tokenName, big.NewInt(10).Bytes()...))
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteWipeTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 2"))

	vmInput.Function = "wipeSingleNFT"
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 3"))
}

func TestEsdt_ExecuteWipeWrongCallValueShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken"), []byte("owner")})
	vmInput.CallValue = big.NewInt(1)

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfFunds, output)
	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_ExecuteWipeNotEnoughGasShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_ExecuteWipeOnNonExistentTokenShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken"), []byte("owner")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNoTickerWithGivenName.Error()))
}

func TestEsdt_ExecuteWipeNotByOwnerShouldFail(t *testing.T) {
	t.Parallel()

	tokenName := "esdtToken"
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "can be called by owner only"))
}

func TestEsdt_ExecuteWipeNonWipeableTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot wipe"))
}

func TestEsdt_ExecuteWipeInvalidDestShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	assert.True(t, strings.Contains(eei.returnMessage, "invalid"))

	vmInput.Function = "wipeSingleNFT"
	vmInput.Arguments = append(vmInput.Arguments, []byte("one"))
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid"))
}

func TestEsdt_ExecuteWipeTransferFailsShouldErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForESDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
			OwnerAddress: []byte("owner"),
			CanWipe:      true,
			TokenType:    []byte(core.FungibleESDT),
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
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{[]byte("esdtToken"), getAddress()})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteWipeSingleNFTTransferFailsShouldErr(t *testing.T) {
	t.Parallel()

	err := errors.New("transfer error")
	args := createMockArgumentsForESDT()
	args.Eei.(*mock.SystemEIStub).GetStorageCalled = func(key []byte) []byte {
		marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
			OwnerAddress: []byte("owner"),
			CanWipe:      true,
			TokenType:    []byte(core.NonFungibleESDT),
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
	vmInput := getDefaultVmInputForFunc("wipeSingleNFT", [][]byte{[]byte("esdtToken"), big.NewInt(10).Bytes(), getAddress()})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_ExecuteWipeShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	addressToWipe := getAddress()
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		TokenType:    []byte(core.FungibleESDT),
		OwnerAddress: owner,
		CanWipe:      true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipe", [][]byte{tokenName, addressToWipe})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToWipe)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTWipe + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteWipeSingleNFTShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	addressToWipe := getAddress()
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		TokenType:    []byte(core.NonFungibleESDT),
		OwnerAddress: owner,
		CanWipe:      true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("wipeSingleNFT", [][]byte{tokenName, big.NewInt(10).Bytes(), addressToWipe})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	vmOutput := eei.CreateVMOutput()
	_, accCreated := vmOutput.OutputAccounts[string(args.ESDTSCAddress)]
	assert.True(t, accCreated)

	destAcc, accCreated := vmOutput.OutputAccounts[string(addressToWipe)]
	assert.True(t, accCreated)

	assert.True(t, len(destAcc.OutputTransfers) == 1)
	outputTransfer := destAcc.OutputTransfers[0]

	assert.Equal(t, big.NewInt(0), outputTransfer.Value)
	assert.Equal(t, uint64(0), outputTransfer.GasLimit)
	expectedInput := core.BuiltInFunctionESDTWipe + "@" + hex.EncodeToString(append(tokenName, big.NewInt(10).Bytes()...))
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecutePauseTooFewArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	esdtData := &ESDTDataV2{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, true, esdtData.IsPaused)

	require.Equal(t, &vmcommon.LogEntry{
		Identifier: []byte(core.BuiltInFunctionESDTPause),
		Topics:     [][]byte{[]byte("esdtToken")},
		Address:    []byte("owner"),
	}, eei.logs[0])
}

func TestEsdt_ExecuteTogglePauseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteUnPauseOnAnUnPausedTokenShouldFail(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	esdtData := &ESDTDataV2{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, false, esdtData.IsPaused)

	require.Equal(t, &vmcommon.LogEntry{
		Identifier: []byte(core.BuiltInFunctionESDTUnPause),
		Topics:     [][]byte{[]byte("esdtToken")},
		Address:    []byte("owner"),
	}, eei.logs[0])
}

func TestEsdt_ExecuteUnPauseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)
}

func TestEsdt_ExecuteTransferOwnershipWrongNumOfArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:      tokenName,
		OwnerAddress:   []byte("owner"),
		CanChangeOwner: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("transferOwnership", [][]byte{[]byte("esdtToken"), []byte("invalid address")})

	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid"))
}

func TestEsdt_ExecuteTransferOwnershipSavesTokenWithNewOwnerAddressSet(t *testing.T) {
	t.Parallel()

	newOwner := getAddress()
	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

	esdtData := &ESDTDataV2{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage(tokenName))
	assert.Equal(t, newOwner, esdtData.OwnerAddress)
}

func TestEsdt_ExecuteEsdtControlChangesWrongNumOfArgumentsShouldFail(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
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

func TestEsdt_ExecuteEsdtControlChangesSavesTokenWithUpgradedProperties(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:        []byte("esdtToken"),
		TokenType:        []byte(core.FungibleESDT),
		OwnerAddress:     []byte("owner"),
		Upgradable:       true,
		BurntValue:       big.NewInt(100),
		MintedValue:      big.NewInt(1000),
		NumWiped:         37,
		NFTCreateStopped: true,
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
		[]byte(canTransferNFTCreateRole), []byte("true"),
	})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	esdtData := &ESDTDataV2{}
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

	assert.Equal(t, 18, len(eei.output))
	assert.Equal(t, []byte("esdtToken"), eei.output[0])
	assert.Equal(t, []byte(core.FungibleESDT), eei.output[1])
	assert.Equal(t, vmInput.CallerAddr, eei.output[2])
	assert.Equal(t, "1000", string(eei.output[3]))
	assert.Equal(t, "100", string(eei.output[4]))
	assert.Equal(t, []byte("NumDecimals-0"), eei.output[5])
	assert.Equal(t, []byte("IsPaused-false"), eei.output[6])
	assert.Equal(t, []byte("CanUpgrade-false"), eei.output[7])
	assert.Equal(t, []byte("CanMint-true"), eei.output[8])
	assert.Equal(t, []byte("CanBurn-true"), eei.output[9])
	assert.Equal(t, []byte("CanChangeOwner-true"), eei.output[10])
	assert.Equal(t, []byte("CanPause-true"), eei.output[11])
	assert.Equal(t, []byte("CanFreeze-true"), eei.output[12])
	assert.Equal(t, []byte("CanWipe-true"), eei.output[13])
	assert.Equal(t, []byte("CanAddSpecialRoles-false"), eei.output[14])
	assert.Equal(t, []byte("CanTransferNFTCreateRole-true"), eei.output[15])
	assert.Equal(t, []byte("NFTCreateStopped-true"), eei.output[16])
	assert.Equal(t, []byte("NumWiped-37"), eei.output[17])
}

func TestEsdt_ExecuteEsdtControlChangesForMultiNFTTransferShouldFaild(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:        []byte("esdtToken"),
		TokenType:        []byte(core.NonFungibleESDT),
		OwnerAddress:     []byte("owner"),
		Upgradable:       true,
		BurntValue:       big.NewInt(0),
		MintedValue:      big.NewInt(0),
		NumWiped:         37,
		NFTCreateStopped: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("controlChanges", [][]byte{[]byte("esdtToken"),
		[]byte(canCreateMultiShard), []byte("true"),
	})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestEsdt_GetSpecialRolesValueNotZeroShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("esdtToken")})
	vmInput.CallValue = big.NewInt(37)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	assert.True(t, strings.Contains(eei.returnMessage, "callValue must be 0"))
}

func TestEsdt_GetSpecialRolesInvalidNumOfArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("esdtToken"), []byte("additional arg")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))
}

func TestEsdt_GetSpecialRolesNotEnoughGasShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ESDTOperations = 10

	e, _ := NewESDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("esdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)

	assert.True(t, strings.Contains(eei.returnMessage, "not enough gas"))
}

func TestEsdt_GetSpecialRolesInvalidTokenShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("invalid esdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	assert.True(t, strings.Contains(eei.returnMessage, "no ticker with given name"))
}

func TestEsdt_GetSpecialRolesNoSpecialRoles(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("esdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 0, len(eei.output))
}

func TestEsdt_GetSpecialRolesShouldWork(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	addr1 := "erd1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s8kszxk"
	addr1Bytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr1)

	addr2 := "erd1e7n8rzxdtl2n2fl6mrsg4l7stp2elxhfy6l9p7eeafspjhhrjq7qk05usw"
	addr2Bytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr2)

	specialRoles := []*ESDTRoles{
		{
			Address: addr1Bytes,
			Roles: [][]byte{
				[]byte(core.ESDTRoleLocalMint),
				[]byte(core.ESDTRoleLocalBurn),
			},
		},
		{
			Address: addr2Bytes,
			Roles: [][]byte{
				[]byte(core.ESDTRoleNFTAddQuantity),
				[]byte(core.ESDTRoleNFTCreate),
				[]byte(core.ESDTRoleNFTBurn),
			},
		},
	}
	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		SpecialRoles: specialRoles,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	e, _ := NewESDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("esdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 2, len(eei.output))
	assert.Equal(t, []byte("erd1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s8kszxk:ESDTRoleLocalMint,ESDTRoleLocalBurn"), eei.output[0])
	assert.Equal(t, []byte("erd1e7n8rzxdtl2n2fl6mrsg4l7stp2elxhfy6l9p7eeafspjhhrjq7qk05usw:ESDTRoleNFTAddQuantity,ESDTRoleNFTCreate,ESDTRoleNFTBurn"), eei.output[1])
}

func TestEsdt_UnsetSpecialRoleWithRemoveEntryFromSpecialRoles(t *testing.T) {
	t.Parallel()

	tokenName := []byte("esdtToken")
	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	owner := "erd1e7n8rzxdtl2n2fl6mrsg4l7stp2elxhfy6l9p7eeafspjhhrjq7qk05usw"
	ownerBytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(owner)

	addr1 := "erd1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s8kszxk"
	addr1Bytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr1)

	addr2 := "erd1rsq30t33aqeg8cuc3q4kfnx0jukzsx52yfua92r233zhhmndl3uszcs5qj"
	addr2Bytes, _ := testscommon.RealWorldBech32PubkeyConverter.Decode(addr2)

	specialRoles := []*ESDTRoles{
		{
			Address: addr1Bytes,
			Roles: [][]byte{
				[]byte(core.ESDTRoleLocalMint),
			},
		},
		{
			Address: addr2Bytes,
			Roles: [][]byte{
				[]byte(core.ESDTRoleNFTAddQuantity),
				[]byte(core.ESDTRoleNFTCreate),
				[]byte(core.ESDTRoleNFTBurn),
			},
		},
	}
	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		OwnerAddress:       ownerBytes,
		SpecialRoles:       specialRoles,
		CanAddSpecialRoles: true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap
	args.Eei = eei

	args.AddressPubKeyConverter = testscommon.RealWorldBech32PubkeyConverter

	e, _ := NewESDTSmartContract(args)

	eei.output = make([][]byte, 0)
	vmInput := getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("esdtToken")})
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 2, len(eei.output))
	assert.Equal(t, []byte("erd1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s8kszxk:ESDTRoleLocalMint"), eei.output[0])
	assert.Equal(t, []byte("erd1rsq30t33aqeg8cuc3q4kfnx0jukzsx52yfua92r233zhhmndl3uszcs5qj:ESDTRoleNFTAddQuantity,ESDTRoleNFTCreate,ESDTRoleNFTBurn"), eei.output[1])

	// unset the role for the address
	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.CallerAddr = ownerBytes
	vmInput.Arguments = [][]byte{[]byte("esdtToken"), addr1Bytes, []byte(core.ESDTRoleLocalMint)}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	// get roles again
	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("esdtToken")})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))

	// set the role for the address
	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.CallerAddr = ownerBytes
	vmInput.Arguments = [][]byte{[]byte("esdtToken"), addr1Bytes, []byte(core.ESDTRoleLocalMint)}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	// get roles again
	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFunc("getSpecialRoles", [][]byte{[]byte("esdtToken")})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 2, len(eei.output))
	assert.Equal(t, []byte("erd1kzzv2uw97q5k9mt458qk3q9u3cwhwqykvyk598q2f6wwx7gvrd9s8kszxk:ESDTRoleLocalMint"), eei.output[1])
}

func TestEsdt_ExecuteConfigChangeGetContractConfig(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("configChange", [][]byte{[]byte("esdtToken"), []byte(burnable)})
	vmInput.CallerAddr = e.ownerAddress
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	newBaseIssingCost := big.NewInt(100)
	newMinTokenNameLength := int64(5)
	newMaxTokenNameLength := int64(20)
	newOwner := vmInput.RecipientAddr
	vmInput = getDefaultVmInputForFunc("configChange",
		[][]byte{newOwner, newBaseIssingCost.Bytes(), big.NewInt(newMinTokenNameLength).Bytes(),
			big.NewInt(newMaxTokenNameLength).Bytes()})
	vmInput.CallerAddr = e.ownerAddress
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	esdtData := &ESDTConfig{}
	_ = args.Marshalizer.Unmarshal(esdtData, eei.GetStorage([]byte(configKeyPrefix)))
	assert.True(t, esdtData.BaseIssuingCost.Cmp(newBaseIssingCost) == 0)
	assert.Equal(t, uint32(newMaxTokenNameLength), esdtData.MaxTokenNameLength)
	assert.Equal(t, uint32(newMinTokenNameLength), esdtData.MinTokenNameLength)
	assert.Equal(t, newOwner, esdtData.OwnerAddress)

	vmInput = getDefaultVmInputForFunc("getContractConfig", make([][]byte, 0))
	vmInput.CallerAddr = []byte("any address")
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	require.Equal(t, 4, len(eei.output))
	assert.Equal(t, newOwner, eei.output[0])
	assert.Equal(t, newBaseIssingCost.Bytes(), eei.output[1])
	assert.Equal(t, big.NewInt(newMinTokenNameLength).Bytes(), eei.output[2])
	assert.Equal(t, big.NewInt(newMaxTokenNameLength).Bytes(), eei.output[3])

}

func TestEsdt_ExecuteClaim(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
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

func getAddress() []byte {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return key
}

func TestEsdt_SetSpecialRoleCheckArgumentsErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestEsdt_SetSpecialRoleCheckBasicOwnershipErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
}

func TestEsdt_SetSpecialRoleNewSendRoleChangeDataErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local err")
	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller"),
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTSetRole@6d79546f6b656e@45534454526f6c654c6f63616c4275726e"), input)
			return localErr
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_SetSpecialRoleAlreadyExists(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalBurn)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTSetRole@6d79546f6b656e@45534454526f6c654c6f63616c4275726e"), input)
			return nil
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_SetSpecialRoleCannotSaveToken(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
				TokenType:          []byte(core.FungibleESDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTSetRole@6d79546f6b656e@45534454526f6c654c6f63616c4275726e"), input)
			castedMarshalizer := args.Marshalizer.(*mock.MarshalizerMock)
			castedMarshalizer.Fail = true
			return nil
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_SetSpecialRoleShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
				TokenType:          []byte(core.FungibleESDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTSetRole@6d79546f6b656e@45534454526f6c654c6f63616c4275726e"), input)
			return nil
		},
		SetStorageCalled: func(key []byte, value []byte) {
			token := &ESDTDataV2{}
			_ = args.Marshalizer.Unmarshal(token, value)
			require.Equal(t, [][]byte{[]byte(core.ESDTRoleLocalMint), []byte(core.ESDTRoleLocalBurn)}, token.SpecialRoles[0].Roles)
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_SetSpecialRoleNFTShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
				TokenType:          []byte(core.NonFungibleESDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTSetRole@6d79546f6b656e@45534454526f6c654e4654437265617465"), input)
			return nil
		},
		SetStorageCalled: func(key []byte, value []byte) {
			token := &ESDTDataV2{}
			_ = args.Marshalizer.Unmarshal(token, value)
			require.Equal(t, [][]byte{[]byte(core.ESDTRoleLocalMint), []byte(core.ESDTRoleNFTCreate)}, token.SpecialRoles[0].Roles)
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.Arguments[2] = []byte(core.ESDTRoleNFTAddQuantity)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.Arguments[2] = []byte(core.ESDTRoleNFTCreate)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_SetSpecialRoleTransferNotEnabledShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.IsESDTTransferRoleFlagEnabledField = false

	token := &ESDTDataV2{
		OwnerAddress: []byte("caller123"),
		SpecialRoles: []*ESDTRoles{
			{
				Address: []byte("myAddress"),
				Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
			},
		},
		TokenType:          []byte(core.NonFungibleESDT),
		CanAddSpecialRoles: true,
	}
	esdtTransferData := core.BuiltInFunctionESDTSetLimitedTransfer + "@" + hex.EncodeToString([]byte("myToken"))
	called := false
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		SendGlobalSettingToAllCalled: func(sender []byte, input []byte) {
			assert.Equal(t, input, []byte(esdtTransferData))
			called = true
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabledField = false
	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleTransfer)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	token.TokenType = []byte(core.NonFungibleESDT)
	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.FungibleESDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.SemiFungibleESDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	enableEpochsHandler.IsESDTTransferRoleFlagEnabledField = true
	called = false
	token.TokenType = []byte(core.NonFungibleESDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.True(t, called)

	token.TokenType = []byte(core.FungibleESDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)

	called = false
	newAddressRole := &ESDTRoles{
		Address: []byte("address"),
		Roles:   [][]byte{[]byte(core.ESDTRoleTransfer)},
	}
	token.SpecialRoles = append(token.SpecialRoles, newAddressRole)
	token.TokenType = []byte(core.SemiFungibleESDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.False(t, called)

	token.SpecialRoles[0].Roles = append(token.SpecialRoles[0].Roles, []byte(core.ESDTRoleTransfer))
	token.TokenType = []byte(core.SemiFungibleESDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.False(t, called)

	vmInput.Function = "unSetSpecialRole"
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.False(t, called)

	esdtTransferData = core.BuiltInFunctionESDTUnSetLimitedTransfer + "@" + hex.EncodeToString([]byte("myToken"))
	token.SpecialRoles = token.SpecialRoles[:1]
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.True(t, called)
}

func TestEsdt_SetSpecialRoleTransferWithTransferRoleEnhancement(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.IsESDTTransferRoleFlagEnabledField = false

	token := &ESDTDataV2{
		OwnerAddress: []byte("caller123"),
		SpecialRoles: []*ESDTRoles{
			{
				Address: []byte("myAddress"),
				Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
			},
		},
		TokenType:          []byte(core.NonFungibleESDT),
		CanAddSpecialRoles: true,
	}
	called := 0
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleTransfer)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	enableEpochsHandler.IsESDTTransferRoleFlagEnabledField = true
	called = 0
	token.TokenType = []byte(core.NonFungibleESDT)
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		if called == 0 {
			assert.Equal(t, core.BuiltInFunctionESDTSetLimitedTransfer+"@"+hex.EncodeToString([]byte("myToken")), string(input))
		} else {
			assert.Equal(t, vmcommon.BuiltInFunctionESDTTransferRoleAddAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress")), string(input))
		}
		called++
	}

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 2)

	called = 0
	newAddressRole := &ESDTRoles{
		Address: []byte("address"),
		Roles:   [][]byte{[]byte(core.ESDTRoleTransfer)},
	}
	token.SpecialRoles = append(token.SpecialRoles, newAddressRole)
	token.TokenType = []byte(core.SemiFungibleESDT)
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		assert.Equal(t, vmcommon.BuiltInFunctionESDTTransferRoleAddAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress")), string(input))
		called++
	}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 1)

	token.SpecialRoles[0].Roles = append(token.SpecialRoles[0].Roles, []byte(core.ESDTRoleTransfer))
	vmInput.Function = "unSetSpecialRole"
	called = 0
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		assert.Equal(t, vmcommon.BuiltInFunctionESDTTransferRoleDeleteAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress")), string(input))
		called++
	}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 1)

	called = 0
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		if called == 0 {
			assert.Equal(t, core.BuiltInFunctionESDTUnSetLimitedTransfer+"@"+hex.EncodeToString([]byte("myToken")), string(input))
		} else {
			assert.Equal(t, vmcommon.BuiltInFunctionESDTTransferRoleDeleteAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress")), string(input))
		}

		called++
	}
	token.SpecialRoles = token.SpecialRoles[:1]
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 2)
}

func TestEsdt_SendAllTransferRoleAddresses(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabledField = false

	token := &ESDTDataV2{
		OwnerAddress: []byte("caller1234"),
		SpecialRoles: []*ESDTRoles{
			{
				Address: []byte("myAddress1"),
				Roles:   [][]byte{[]byte(core.ESDTRoleTransfer)},
			},
			{
				Address: []byte("myAddress2"),
				Roles:   [][]byte{[]byte(core.ESDTRoleTransfer)},
			},
			{
				Address: []byte("myAddress3"),
				Roles:   [][]byte{[]byte(core.ESDTRoleTransfer)},
			},
		},
		TokenType:          []byte(core.NonFungibleESDT),
		CanAddSpecialRoles: true,
	}
	called := 0
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("sendAllTransferRoleAddresses", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress")}
	vmInput.CallerAddr = []byte("caller1234")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionNotFound, retCode)

	enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabledField = true
	eei.ReturnMessage = ""
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, "wrong number of arguments, expected 1", eei.ReturnMessage)

	called = 0
	token.TokenType = []byte(core.NonFungibleESDT)
	eei.SendGlobalSettingToAllCalled = func(sender []byte, input []byte) {
		assert.Equal(t, vmcommon.BuiltInFunctionESDTTransferRoleAddAddress+"@"+hex.EncodeToString([]byte("myToken"))+"@"+hex.EncodeToString([]byte("myAddress1"))+"@"+hex.EncodeToString([]byte("myAddress2"))+"@"+hex.EncodeToString([]byte("myAddress3")), string(input))
		called++
	}
	vmInput.Arguments = [][]byte{[]byte("myToken")}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Equal(t, called, 1)

	called = 0
	token.SpecialRoles = make([]*ESDTRoles, 0)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, "no address with transfer role", eei.ReturnMessage)
}

func TestEsdt_SetSpecialRoleSFTShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
				TokenType:          []byte(core.SemiFungibleESDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTSetRole@6d79546f6b656e@45534454526f6c654e46544164645175616e74697479"), input)
			return nil
		},
		SetStorageCalled: func(key []byte, value []byte) {
			token := &ESDTDataV2{}
			_ = args.Marshalizer.Unmarshal(token, value)
			require.Equal(t, [][]byte{[]byte(core.ESDTRoleLocalMint), []byte(core.ESDTRoleNFTAddQuantity)}, token.SpecialRoles[0].Roles)
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.Arguments[2] = []byte(core.ESDTRoleNFTAddQuantity)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_SetSpecialRoleCreateNFTTwoTimesShouldError(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleNFTCreate)},
					},
				},
				TokenType:          []byte(core.NonFungibleESDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller234"), []byte(core.ESDTRoleNFTCreate)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_SetSpecialRoleCreateNFTTwoTimesMultiShardShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddres4"),
						Roles:   [][]byte{[]byte(core.ESDTRoleNFTCreate)},
					},
				},
				TokenType:           []byte(core.NonFungibleESDT),
				CanAddSpecialRoles:  true,
				CanCreateMultiShard: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("setSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller234"), []byte(core.ESDTRoleNFTCreate)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, eei.ReturnMessage, vm.ErrInvalidAddress.Error())

	vmInput.Arguments[1] = []byte("caller23X")
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_UnSetSpecialRoleCreateNFTShouldError(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleNFTCreate)},
					},
				},
				TokenType:          []byte(core.NonFungibleESDT),
				CanAddSpecialRoles: true,
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller234"), []byte(core.ESDTRoleNFTCreate)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_UnsetSpecialRoleCheckArgumentsErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)
}

func TestEsdt_UnsetSpecialRoleCheckArgumentsInvalidRoleErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte("mirage")}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
}

func TestEsdt_UnsetSpecialRoleCheckArgumentsDuplicatedRoleInArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte(core.ESDTRoleLocalBurn), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_UnsetSpecialRoleCheckBasicOwnershipErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("1"), []byte("caller"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.OutOfFunds, retCode)
}

func TestEsdt_UnsetSpecialRoleNewShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller"),
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("caller"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_UnsetSpecialRoleCannotRemoveRoleNotExistsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalBurn)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_UnsetSpecialRoleRemoveRoleTransferErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local err")
	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTUnSetRole@6d79546f6b656e@45534454526f6c654c6f63616c4d696e74"), input)
			return localErr
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalMint)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_UnsetSpecialRoleRemoveRoleSaveTokenErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTUnSetRole@6d79546f6b656e@45534454526f6c654c6f63616c4d696e74"), input)
			castedMarshalizer := args.Marshalizer.(*mock.MarshalizerMock)
			castedMarshalizer.Fail = true
			return nil
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalMint)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_UnsetSpecialRoleRemoveRoleShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTUnSetRole@6d79546f6b656e@45534454526f6c654c6f63616c4d696e74"), input)
			return nil
		},
		SetStorageCalled: func(key []byte, value []byte) {
			token := &ESDTDataV2{}
			_ = args.Marshalizer.Unmarshal(token, value)
			require.Len(t, token.SpecialRoles, 0)
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("unSetSpecialRole", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("myToken"), []byte("myAddress"), []byte(core.ESDTRoleLocalMint)}
	vmInput.CallerAddr = []byte("caller123")
	vmInput.CallValue = big.NewInt(0)
	vmInput.GasProvided = 50000000

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_StopNFTCreateForeverCheckArgumentsErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("stopNFTCreate", [][]byte{})
	vmInput.Arguments = [][]byte{{1}, {2}}
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	vmInput.CallValue = big.NewInt(0)
	vmInput.Arguments = [][]byte{{1}}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_StopNFTCreateForeverCallErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	token := &ESDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*ESDTRoles{
			{
				Address: []byte("myAddress"),
				Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("stopNFTCreate", [][]byte{[]byte("tokenID")})
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(0)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.CallerAddr = token.OwnerAddress
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.NonFungibleESDT)
	token.NFTCreateStopped = true
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.NFTCreateStopped = false
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_StopNFTCreateForeverCallShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	token := &ESDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*ESDTRoles{
			{
				Address: []byte("myAddress"),
				Roles:   [][]byte{[]byte(core.ESDTRoleNFTCreate)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTUnSetRole@746f6b656e4944@45534454526f6c654e4654437265617465"), input)
			return nil
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("stopNFTCreate", [][]byte{[]byte("tokenID")})
	vmInput.CallerAddr = token.OwnerAddress
	vmInput.CallValue = big.NewInt(0)

	token.TokenType = []byte(core.NonFungibleESDT)
	token.NFTCreateStopped = false
	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_TransferNFTCreateCheckArgumentsErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("transferNFTCreateRole", [][]byte{})
	vmInput.Arguments = [][]byte{{1}, {2}}
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(1)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	vmInput.CallValue = big.NewInt(0)
	vmInput.Arguments = [][]byte{{1}, []byte("caller3"), {3}}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_TransferNFTCreateCallErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	token := &ESDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*ESDTRoles{
			{
				Address: []byte("caller1"),
				Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("transferNFTCreateRole", [][]byte{[]byte("tokenID"), []byte("caller3"), []byte("caller22")})
	vmInput.CallerAddr = []byte("caller2")
	vmInput.CallValue = big.NewInt(0)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.CallerAddr = token.OwnerAddress
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.FungibleESDT)
	token.CanTransferNFTCreateRole = true
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	token.TokenType = []byte(core.NonFungibleESDT)
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	vmInput.Arguments[2] = vmInput.Arguments[1]
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.FunctionWrongSignature, retCode)

	vmInput.Arguments[2] = []byte("caller2")
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_TransferNFTCreateCallShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	token := &ESDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*ESDTRoles{
			{
				Address: []byte("caller3"),
				Roles:   [][]byte{[]byte(core.ESDTRoleNFTCreate)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTNFTCreateRoleTransfer@746f6b656e4944@63616c6c657232"), input)
			require.Equal(t, destination, []byte("caller3"))
			return nil
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("transferNFTCreateRole", [][]byte{[]byte("tokenID"), []byte("caller3"), []byte("caller2")})
	vmInput.CallerAddr = token.OwnerAddress
	vmInput.CallValue = big.NewInt(0)

	token.TokenType = []byte(core.NonFungibleESDT)
	token.CanTransferNFTCreateRole = true
	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_TransferNFTCreateCallMultiShardShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	token := &ESDTDataV2{
		OwnerAddress: []byte("caller1"),
		SpecialRoles: []*ESDTRoles{
			{
				Address: []byte("3caller"),
				Roles:   [][]byte{[]byte(core.ESDTRoleNFTCreate)},
			},
		},
	}
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
		TransferCalled: func(destination []byte, sender []byte, value *big.Int, input []byte) error {
			require.Equal(t, []byte("ESDTNFTCreateRoleTransfer@746f6b656e4944@3263616c6c6572"), input)
			require.Equal(t, destination, []byte("3caller"))
			return nil
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("transferNFTCreateRole", [][]byte{[]byte("tokenID"), []byte("3caller"), []byte("caller2")})
	vmInput.CallerAddr = token.OwnerAddress
	vmInput.CallValue = big.NewInt(0)

	token.TokenType = []byte(core.NonFungibleESDT)
	token.CanTransferNFTCreateRole = true
	token.CanCreateMultiShard = true
	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)

	vmInput.Arguments = [][]byte{[]byte("tokenID"), []byte("3caller"), []byte("2caller")}
	retCode = e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_SetNewGasCost(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	e.SetNewGasCost(vm.GasCost{BuiltInCost: vm.BuiltInCost{
		ChangeOwnerAddress: 10000,
	}})

	require.Equal(t, uint64(10000), e.gasCost.BuiltInCost.ChangeOwnerAddress)
}

func TestEsdt_GetAllAddressesAndRolesNoArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("getAllAddressesAndRoles", [][]byte{})
	vmInput.Arguments = nil

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_GetAllAddressesAndRolesCallWithValueShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("getAllAddressesAndRoles", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("arg")}
	vmInput.CallValue = big.NewInt(0)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.UserError, retCode)
}

func TestEsdt_GetAllAddressesAndRolesCallGetExistingTokenErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{
		GetStorageCalled: func(key []byte) []byte {
			token := &ESDTDataV2{
				OwnerAddress: []byte("caller123"),
				SpecialRoles: []*ESDTRoles{
					{
						Address: []byte("myAddress"),
						Roles:   [][]byte{[]byte(core.ESDTRoleLocalMint)},
					},
				},
			}
			tokenBytes, _ := args.Marshalizer.Marshal(token)
			return tokenBytes
		},
	}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("getAllAddressesAndRoles", [][]byte{})
	vmInput.Arguments = [][]byte{[]byte("arg")}
	vmInput.CallValue = big.NewInt(0)

	retCode := e.Execute(vmInput)
	require.Equal(t, vmcommon.Ok, retCode)
}

func TestEsdt_CanUseContract(t *testing.T) {
	args := createMockArgumentsForESDT()
	eei := &mock.SystemEIStub{}
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	require.True(t, e.CanUseContract())
}

func TestEsdt_ExecuteIssueMetaESDT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	e, _ := NewESDTSmartContract(args)

	enableEpochsHandler.IsMetaESDTSetFlagEnabledField = false
	vmInput := getDefaultVmInputForFunc("registerMetaESDT", nil)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "invalid method to call")

	eei.returnMessage = ""
	eei.gasRemaining = 9999
	enableEpochsHandler.IsMetaESDTSetFlagEnabledField = true
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)
	vmInput.Arguments = [][]byte{[]byte("tokenName")}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), big.NewInt(20).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of decimals"))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "ticker name is not valid"))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))
}

func TestEsdt_ExecuteChangeSFTToMetaESDT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	e, _ := NewESDTSmartContract(args)

	enableEpochsHandler.IsMetaESDTSetFlagEnabledField = false
	vmInput := getDefaultVmInputForFunc("changeSFTToMetaESDT", nil)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "invalid method to call")

	eei.returnMessage = ""
	eei.gasRemaining = 9999
	enableEpochsHandler.IsMetaESDTSetFlagEnabledField = true
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.Arguments = [][]byte{[]byte("tokenName"), big.NewInt(20).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of decimals"))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "no ticker with given name"))

	_ = e.saveToken(vmInput.Arguments[0], &ESDTDataV2{TokenType: []byte(core.NonFungibleESDT), OwnerAddress: vmInput.CallerAddr})
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "change can happen to semi fungible tokens only"))

	_ = e.saveToken(vmInput.Arguments[0], &ESDTDataV2{TokenType: []byte(core.SemiFungibleESDT), OwnerAddress: vmInput.CallerAddr})
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	token, _ := e.getExistingToken(vmInput.Arguments[0])
	assert.Equal(t, token.NumDecimals, uint32(10))
	assert.Equal(t, token.TokenType, []byte(metaESDT))
}

func TestEsdt_ExecuteIssueSFTAndChangeSFTToMetaESDT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	eei.returnMessage = ""
	eei.gasRemaining = 9999

	vmInput := getDefaultVmInputForFunc("issueSemiFungible", nil)
	vmInput.CallValue = e.baseIssuingCost
	vmInput.Arguments = [][]byte{[]byte("name"), []byte("TOKEN")}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	fullTicker := eei.output[0]

	token, _ := e.getExistingToken(fullTicker)
	assert.Equal(t, token.NumDecimals, uint32(0))
	assert.Equal(t, token.TokenType, []byte(core.SemiFungibleESDT))

	vmInput.CallValue = big.NewInt(0)
	vmInput.Function = "changeSFTToMetaESDT"
	vmInput.Arguments = [][]byte{fullTicker, big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	token, _ = e.getExistingToken(fullTicker)
	assert.Equal(t, token.NumDecimals, uint32(10))
	assert.Equal(t, token.TokenType, []byte(metaESDT))

	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "change can happen to semi fungible tokens only"))
}

func TestEsdt_ExecuteRegisterAndSetErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	e, _ := NewESDTSmartContract(args)

	enableEpochsHandler.IsESDTRegisterAndSetAllRolesFlagEnabledField = false
	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionNotFound, output)
	assert.Equal(t, eei.returnMessage, "invalid method to call")

	eei.returnMessage = ""
	eei.gasRemaining = 9999
	enableEpochsHandler.IsESDTRegisterAndSetAllRolesFlagEnabledField = true
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)
	vmInput.Arguments = [][]byte{[]byte("tokenName")}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "arguments length mismatch")

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), []byte("VAL"), big.NewInt(20).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidArgument.Error()))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), []byte("FNG"), big.NewInt(10).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "ticker name is not valid"))

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("ticker"), []byte("FNG"), big.NewInt(20).Bytes()}
	eei.returnMessage = ""
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of decimals"))
}

func TestEsdt_ExecuteRegisterAndSetFungible(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), []byte("FNG"), big.NewInt(10).Bytes()}
	eei.gasRemaining = 9999
	eei.returnMessage = ""
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))

	token, _ := e.getExistingToken(eei.output[0])
	assert.Equal(t, token.TokenType, []byte(core.FungibleESDT))
}

func TestEsdt_ExecuteRegisterAndSetNonFungible(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), []byte("NFT"), big.NewInt(10).Bytes()}
	eei.gasRemaining = 9999
	eei.returnMessage = ""
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))

	token, _ := e.getExistingToken(eei.output[0])
	assert.Equal(t, token.TokenType, []byte(core.NonFungibleESDT))
}

func TestEsdt_ExecuteRegisterAndSetSemiFungible(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), []byte("SFT"), big.NewInt(10).Bytes()}
	eei.gasRemaining = 9999
	eei.returnMessage = ""
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))

	token, _ := e.getExistingToken(eei.output[0])
	assert.Equal(t, token.TokenType, []byte(core.SemiFungibleESDT))
	lenOutTransfer := 0
	for _, outAcc := range eei.outputAccounts {
		lenOutTransfer += len(outAcc.OutputTransfers)
	}
	assert.Equal(t, uint32(lenOutTransfer), 1+eei.blockChainHook.NumberOfShards())
}

func TestEsdt_ExecuteRegisterAndSetMetaESDTShouldSetType(t *testing.T) {
	t.Parallel()

	registerAndSetAllRolesWithTypeCheck(t, []byte("NFT"), []byte(core.NonFungibleESDT))
	registerAndSetAllRolesWithTypeCheck(t, []byte("SFT"), []byte(core.SemiFungibleESDT))
	registerAndSetAllRolesWithTypeCheck(t, []byte("META"), []byte(metaESDT))
	registerAndSetAllRolesWithTypeCheck(t, []byte("FNG"), []byte(core.FungibleESDT))
}

func registerAndSetAllRolesWithTypeCheck(t *testing.T, typeArgument []byte, expectedType []byte) {
	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabledField = false
	vmInput := getDefaultVmInputForFunc("registerAndSetAllRoles", nil)
	vmInput.CallValue = big.NewInt(0).Set(e.baseIssuingCost)

	vmInput.Arguments = [][]byte{[]byte("tokenName"), []byte("TICKER"), typeArgument, big.NewInt(10).Bytes()}
	eei.gasRemaining = 9999
	eei.returnMessage = ""
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, len(eei.output), 1)
	assert.True(t, strings.Contains(string(eei.output[0]), "TICKER-"))

	token, _ := e.getExistingToken(eei.output[0])
	assert.Equal(t, expectedType, token.TokenType)

	lenOutTransfer := 0
	for _, outAcc := range eei.outputAccounts {
		lenOutTransfer += len(outAcc.OutputTransfers)
	}
	assert.Equal(t, lenOutTransfer, 1)
}

func TestEsdt_setBurnRoleGlobally(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("setBurnRoleGlobally", [][]byte{})

	enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabledField = false
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionNotFound, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid method to call"))

	enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabledField = true
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 1"))

	owner := bytes.Repeat([]byte{1}, 32)
	tokenName := []byte("TOKEN-ABABAB")
	tokensMap := map[string][]byte{}
	marshalizedData, _ := args.Marshalizer.Marshal(ESDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     true,
	})
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap

	vmInput.CallerAddr = owner
	vmInput.Arguments = [][]byte{tokenName}
	output = e.Execute(vmInput)
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
	expectedInput := vmcommon.BuiltInFunctionESDTSetBurnRoleForAll + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)

	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot set burn role globally as it was already set"))
}

func TestEsdt_unsetBurnRoleGlobally(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei

	e, _ := NewESDTSmartContract(args)
	vmInput := getDefaultVmInputForFunc("unsetBurnRoleGlobally", [][]byte{})

	enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabledField = false
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionNotFound, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid method to call"))

	enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabledField = true
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments, wanted 1"))

	owner := bytes.Repeat([]byte{1}, 32)
	tokenName := []byte("TOKEN-ABABAB")
	tokensMap := map[string][]byte{}
	token := &ESDTDataV2{
		TokenName:    tokenName,
		OwnerAddress: owner,
		CanPause:     true,
		IsPaused:     true,
	}

	burnForAllRole := &ESDTRoles{Roles: [][]byte{[]byte(vmcommon.ESDTRoleBurnForAll)}, Address: []byte{}}
	token.SpecialRoles = append(token.SpecialRoles, burnForAllRole)

	marshalizedData, _ := args.Marshalizer.Marshal(token)
	tokensMap[string(tokenName)] = marshalizedData
	eei.storageUpdate[string(eei.scAddress)] = tokensMap

	vmInput.CallerAddr = owner
	vmInput.Arguments = [][]byte{tokenName}
	output = e.Execute(vmInput)
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
	expectedInput := vmcommon.BuiltInFunctionESDTUnSetBurnRoleForAll + "@" + hex.EncodeToString(tokenName)
	assert.Equal(t, []byte(expectedInput), outputTransfer.Data)
	assert.Equal(t, vmData.DirectCall, outputTransfer.CallType)

	token, err := e.getExistingToken(tokenName)
	assert.Nil(t, err)
	assert.Equal(t, len(token.SpecialRoles), 0)

	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot unset burn role globally as it was not set"))
}

func TestEsdt_CheckRolesOnMetaESDT(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForESDT()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	eei := createDefaultEei()
	args.Eei = eei
	e, _ := NewESDTSmartContract(args)

	enableEpochsHandler.IsManagedCryptoAPIsFlagEnabledField = false
	err := e.checkSpecialRolesAccordingToTokenType([][]byte{[]byte("random")}, &ESDTDataV2{TokenType: []byte(metaESDT)})
	assert.Nil(t, err)

	enableEpochsHandler.IsManagedCryptoAPIsFlagEnabledField = true
	err = e.checkSpecialRolesAccordingToTokenType([][]byte{[]byte("random")}, &ESDTDataV2{TokenType: []byte(metaESDT)})
	assert.Equal(t, err, vm.ErrInvalidArgument)
}
