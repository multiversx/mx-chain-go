package systemSmartContracts

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var logFreezeAccount = logger.GetOrCreate("systemSmartContracts/freezeAccount")

// Key prefixes
const (
	keyGuardian = "guardians"
)

// Functions
const (
	setGuardian = "setGuardian"
)

type Guardian struct {
	Address         []byte
	ActivationEpoch uint32
}

type Guardians struct {
	Data []*Guardian
}

type freezeAccount struct {
	blockChainHook  vmcommon.BlockchainHook
	gasCost         vm.GasCost
	marshaller      marshal.Marshalizer
	systemEI        vm.SystemEI
	pubKeyConverter core.PubkeyConverter
}

func (fa *freezeAccount) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	switch args.Function {
	case setGuardian:
		return fa.setGuardian(args)
	default:
		return vmcommon.UserError
	}
}

// 1. User does NOT have any guardian => set guardian
// 2. User has ONE guardian pending => does not set, wait until first one is set
// 3. User has ONE guardian enabled => add it
// 4. User has TWO guardians. FIRST is enabled, SECOND is pending => change pending with current one / does nothing until it is set
// 5. User has TWO guardians. FIRST is enabled, SECOND is enabled => replace oldest one

func (fa *freezeAccount) setGuardian(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !isZero(args.CallValue) {
		fa.systemEI.AddReturnMessage(fmt.Sprintf("expected value must be zero, got %v instead", args.CallValue))
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		fa.systemEI.AddReturnMessage(fmt.Sprintf("invalid number of arguments, expected 1, got %d ", len(args.Arguments)))
		return vmcommon.FunctionWrongSignature
	}

	err := fa.systemEI.UseGas(uint64(len(args.Arguments)) * fa.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		fa.systemEI.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	//TODO: Cannot bet owner a guardian

	if !fa.isAddressValid(args.Arguments[0]) {
		fa.systemEI.AddReturnMessage("invalid address")
		return vmcommon.UserError
	}

	guardians, err := fa.getGuardians(args.CallerAddr)
	if err != nil {
		fa.systemEI.AddReturnMessage(err.Error())
		return vmcommon.ExecutionFailed
	}
	// Case 1
	if len(guardians.Data) == 0 {
		err = fa.addGuardian(args.CallerAddr, args.Arguments[0], guardians)
		if err != nil {
			fa.systemEI.AddReturnMessage(err.Error())
			return vmcommon.ExecutionFailed
		}
		return vmcommon.Ok
		// Case 2
	} else if len(guardians.Data) == 1 && fa.pending(guardians.Data[0]) {
		fa.systemEI.AddReturnMessage(fmt.Sprintf("owner already has one guardian pending: %s",
			fa.pubKeyConverter.Encode(guardians.Data[0].Address)))
		return vmcommon.UserError
		// Case 3
	} else if len(guardians.Data) == 1 && !fa.pending(guardians.Data[0]) {
		err = fa.addGuardian(args.CallerAddr, args.Arguments[0], guardians)
		if err != nil {
			fa.systemEI.AddReturnMessage(err.Error())
			return vmcommon.ExecutionFailed
		}
		return vmcommon.Ok
		// Case 4
	} else if len(guardians.Data) == 2 && fa.pending(guardians.Data[1]) {
		fa.systemEI.AddReturnMessage(fmt.Sprintf("owner already has one guardian pending: %s",
			fa.pubKeyConverter.Encode(guardians.Data[1].Address)))
		return vmcommon.UserError
		// Case 5
	} else if len(guardians.Data) == 2 && !fa.pending(guardians.Data[1]) {
		guardians.Data = guardians.Data[1:] // remove oldest guardian
		err = fa.addGuardian(args.CallerAddr, args.Arguments[0], guardians)
		if err != nil {
			fa.systemEI.AddReturnMessage(err.Error())
			return vmcommon.ExecutionFailed
		}
		return vmcommon.Ok
	} else {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (fa *freezeAccount) pending(guardian *Guardian) bool {
	currEpoch := fa.blockChainHook.CurrentEpoch()
	remaining := currEpoch - guardian.ActivationEpoch // any edge case here for which we should use abs?
	return remaining < 20
}

func (fa *freezeAccount) addGuardian(address []byte, guardianAddress []byte, guardians Guardians) error {
	guardian := &Guardian{
		Address:         guardianAddress,
		ActivationEpoch: fa.blockChainHook.CurrentEpoch() + 20,
	}

	guardians.Data = append(guardians.Data, guardian)
	marshalledData, err := fa.marshaller.Marshal(guardians)
	if err != nil {
		return err
	}

	fa.systemEI.SetStorageForAddress(address, []byte(keyGuardian), marshalledData)
	return nil
}

func (fa *freezeAccount) getGuardians(address []byte) (Guardians, error) {
	marshalledData := fa.systemEI.GetStorageFromAddress(address, []byte(keyGuardian))

	guardians := Guardians{}
	err := fa.marshaller.Unmarshal(guardians, marshalledData)
	if err != nil {
		return Guardians{}, err
	}

	return guardians, nil
}

func isZero(n *big.Int) bool {
	return len(n.Bits()) == 0
}

// TODO: Move this to common  + remove from esdt.go
func (fa *freezeAccount) isAddressValid(addressBytes []byte) bool {
	isLengthOk := len(addressBytes) == fa.pubKeyConverter.Len()
	if !isLengthOk {
		return false
	}

	encodedAddress := fa.pubKeyConverter.Encode(addressBytes)

	return encodedAddress != ""
}

func (fa *freezeAccount) CanUseContract() bool {
	return true
}

func (fa *freezeAccount) SetNewGasCost(gasCost vm.GasCost) {

}

func (fa *freezeAccount) IsInterfaceNil() bool {
	return fa == nil
}
