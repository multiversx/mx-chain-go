package builtInFunctions

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

const argumentsPerESDTTransfer = 3

type unbackedEGLDCreditGuard struct {
	inner               vmcommon.BuiltinFunction
	enableEpochsHandler vmcommon.EnableEpochsHandler
	shardCoordinator    vmcommon.Coordinator
	callArgsParser      process.CallArgumentsParser
}

func newUnbackedEGLDCreditGuard(
	inner vmcommon.BuiltinFunction,
	enableEpochsHandler vmcommon.EnableEpochsHandler,
	shardCoordinator vmcommon.Coordinator,
) (*unbackedEGLDCreditGuard, error) {
	if check.IfNil(inner) {
		return nil, process.ErrNilBuiltInFunction
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &unbackedEGLDCreditGuard{
		inner:               inner,
		enableEpochsHandler: enableEpochsHandler,
		shardCoordinator:    shardCoordinator,
		callArgsParser:      parsers.NewCallArgsParser(),
	}, nil
}

func wrapMultiESDTNFTTransferWithUnbackedEGLDGuard(
	container vmcommon.BuiltInFunctionContainer,
	enableEpochsHandler vmcommon.EnableEpochsHandler,
	shardCoordinator vmcommon.Coordinator,
) error {
	if check.IfNil(container) {
		return process.ErrNilBuiltInFunction
	}

	inner, err := container.Get(core.BuiltInFunctionMultiESDTNFTTransfer)
	if err != nil {
		return err
	}
	if _, alreadyWrapped := inner.(*unbackedEGLDCreditGuard); alreadyWrapped {
		return nil
	}

	wrapper, err := newUnbackedEGLDCreditGuard(inner, enableEpochsHandler, shardCoordinator)
	if err != nil {
		return err
	}

	return container.Replace(core.BuiltInFunctionMultiESDTNFTTransfer, wrapper)
}

// ProcessBuiltinFunction enforces CallValue backing for dest-path EGLD-000000 credits
// and attaches that amount to cross-shard output transfers on the sender path so
// legitimate MultiESDTNFTTransfer executions still settle after the dest-path check.
func (g *unbackedEGLDCreditGuard) ProcessBuiltinFunction(
	acntSnd, acntDst vmcommon.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	if g.enableEpochsHandler.IsFlagEnabled(common.GuardUnbackedEGLDInMultiTransferFlag) {
		err := g.rejectUnbackedDestPathCredit(acntSnd, acntDst, vmInput)
		if err != nil {
			return nil, err
		}
	}

	vmOutput, err := g.inner.ProcessBuiltinFunction(acntSnd, acntDst, vmInput)
	if err != nil || vmOutput == nil {
		return vmOutput, err
	}

	if g.enableEpochsHandler.IsFlagEnabled(common.GuardUnbackedEGLDInMultiTransferFlag) {
		g.attachCallValueOnCrossShardEGLDTransfers(vmInput, vmOutput)
	}

	return vmOutput, nil
}

func (g *unbackedEGLDCreditGuard) rejectUnbackedDestPathCredit(
	acntSnd, acntDst vmcommon.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) error {
	if vmInput == nil {
		return nil
	}
	if !check.IfNil(acntSnd) || check.IfNil(acntDst) {
		return nil
	}

	egldSum := sumEGLDInMultiTransfer(vmInput)
	if egldSum.Sign() <= 0 {
		return nil
	}

	callValue := vmInput.CallValue
	if callValue == nil {
		callValue = big.NewInt(0)
	}
	if callValue.Cmp(egldSum) < 0 {
		return process.ErrUnbackedEGLDInMultiTransfer
	}

	return nil
}

func (g *unbackedEGLDCreditGuard) attachCallValueOnCrossShardEGLDTransfers(
	vmInput *vmcommon.ContractCallInput,
	vmOutput *vmcommon.VMOutput,
) {
	egldSum := sumEGLDInMultiTransfer(vmInput)
	if egldSum.Sign() <= 0 {
		return
	}

	for _, outAcc := range vmOutput.OutputAccounts {
		if outAcc == nil || len(outAcc.Address) == 0 {
			continue
		}
		if g.shardCoordinator.ComputeId(outAcc.Address) == g.shardCoordinator.SelfId() {
			continue
		}

		for i := range outAcc.OutputTransfers {
			if !g.outputTransferCreditsEGLD(&outAcc.OutputTransfers[i]) {
				continue
			}
			if outAcc.OutputTransfers[i].Value == nil {
				outAcc.OutputTransfers[i].Value = big.NewInt(0)
			}
			if outAcc.OutputTransfers[i].Value.Cmp(egldSum) < 0 {
				outAcc.OutputTransfers[i].Value = new(big.Int).Set(egldSum)
			}
		}
	}
}

func (g *unbackedEGLDCreditGuard) outputTransferCreditsEGLD(transfer *vmcommon.OutputTransfer) bool {
	if transfer == nil || len(transfer.Data) == 0 {
		return false
	}

	function, args, err := g.callArgsParser.ParseData(string(transfer.Data))
	if err != nil || function != core.BuiltInFunctionMultiESDTNFTTransfer {
		return false
	}

	return sumEGLDFromMultiTransferArgs(args, false).Sign() > 0
}

func sumEGLDInMultiTransfer(vmInput *vmcommon.ContractCallInput) *big.Int {
	if vmInput == nil {
		return big.NewInt(0)
	}

	isSenderFormat := bytes.Equal(vmInput.CallerAddr, vmInput.RecipientAddr)
	return sumEGLDFromMultiTransferArgs(vmInput.Arguments, isSenderFormat)
}

func sumEGLDFromMultiTransferArgs(args [][]byte, isSenderFormat bool) *big.Int {
	sum := big.NewInt(0)
	if len(args) == 0 {
		return sum
	}

	start := 0
	if isSenderFormat {
		if len(args) < 2 {
			return sum
		}
		start = 1
	}

	numOfTransfers := big.NewInt(0).SetBytes(args[start]).Uint64()
	idx := start + 1
	for i := uint64(0); i < numOfTransfers; i++ {
		if idx+2 >= len(args) {
			break
		}
		tokenID := args[idx]
		transferredValue := big.NewInt(0).SetBytes(args[idx+2])
		if bytes.Equal(tokenID, []byte(vmcommon.EGLDIdentifier)) && transferredValue.Sign() > 0 {
			sum.Add(sum, transferredValue)
		}
		idx += argumentsPerESDTTransfer
	}

	return sum
}

// SetNewGasConfig forwards gas schedule updates to the inner builtin
func (g *unbackedEGLDCreditGuard) SetNewGasConfig(gasCost *vmcommon.GasCost) {
	g.inner.SetNewGasConfig(gasCost)
}

// IsActive forwards to the inner builtin
func (g *unbackedEGLDCreditGuard) IsActive() bool {
	return g.inner.IsActive()
}

// SetPayableChecker forwards payable setup so factory.SetPayableHandler keeps working
func (g *unbackedEGLDCreditGuard) SetPayableChecker(payableHandler vmcommon.PayableChecker) error {
	setter, ok := g.inner.(vmcommon.AcceptPayableChecker)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	return setter.SetPayableChecker(payableHandler)
}

// CheckIsExecutable forwards executable checks when the inner builtin implements them
func (g *unbackedEGLDCreditGuard) CheckIsExecutable(
	senderAddr []byte,
	value *big.Int,
	receiverAddr []byte,
	gasProvidedForCall uint64,
	arguments [][]byte,
) error {
	checker, ok := g.inner.(scrCommon.ExecutableChecker)
	if !ok {
		return nil
	}

	return checker.CheckIsExecutable(senderAddr, value, receiverAddr, gasProvidedForCall, arguments)
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *unbackedEGLDCreditGuard) IsInterfaceNil() bool {
	return g == nil
}
