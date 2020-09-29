package systemSmartContracts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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

func TestEsdt_ExecuteIssue(t *testing.T) {
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
		Function:      "issue",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("1000")}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments[0] = []byte("01234567891")
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue
	output = e.Execute(vmInput)

	assert.Equal(t, vmcommon.Ok, output)

	vmInput.Arguments[0] = []byte("01234567891&*@")
	output = e.Execute(vmInput)
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

func TestEsdt_ExecuteIssueProtected(t *testing.T) {
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
		Function:      "issueProtected",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	vmInput.CallerAddr = e.ownerAddress
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("1000")}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("newOwner"), []byte("name"), []byte("1000")}

	vmInput.Arguments[0] = []byte("01234567891")
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.ESDTIssue
	output = e.Execute(vmInput)

	assert.Equal(t, vmcommon.Ok, output)
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
