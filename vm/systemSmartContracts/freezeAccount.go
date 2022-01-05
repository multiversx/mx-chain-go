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

type Guardians struct {
	Guardians [][]byte
	Epoch     uint32
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

func (fa *freezeAccount) setGuardian(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {

	if !isZero(args.CallValue) {
		fa.systemEI.AddReturnMessage(fmt.Sprintf("expected value must be zero, got %v instead", args.CallValue))
		return vmcommon.UserError
	}
	if len(args.Arguments) == 0 {
		fa.systemEI.AddReturnMessage("not enough arguments, expected at least one")
		return vmcommon.FunctionWrongSignature
	}
	// TODO: HERE GET VALUE FROM CONFIG
	if len(args.Arguments) > 2 {
		fa.systemEI.AddReturnMessage(fmt.Sprintf("cannot set more than %v guardians", 2))
		return vmcommon.FunctionWrongSignature
	}

	err := fa.systemEI.UseGas(uint64(len(args.Arguments)) * fa.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		fa.systemEI.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	guardianAddresses := make([][]byte, len(args.Arguments))
	for _, guardianAddress := range args.Arguments {
		addr, err := fa.pubKeyConverter.Decode(string(guardianAddress))
		if err != nil {
			fa.systemEI.AddReturnMessage(fmt.Sprintf("invalid address: %s", string(guardianAddress)))
			return vmcommon.UserError
		}
		guardianAddresses = append(guardianAddresses, addr)
	}
	currentEpoch := fa.blockChainHook.CurrentEpoch()
	dataToStore := &Guardians{
		Guardians: guardianAddresses,
		Epoch:     currentEpoch,
	}

	marshalledData, err := fa.marshaller.Marshal(dataToStore)
	if err != nil {
		fa.systemEI.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	fa.systemEI.SetStorageForAddress(args.CallerAddr, []byte(keyGuardian), marshalledData)

	return vmcommon.Ok
}

func isZero(n *big.Int) bool {
	return len(n.Bits()) == 0
}

func (fa *freezeAccount) CanUseContract() bool {
	return true
}

func (fa *freezeAccount) SetNewGasCost(gasCost vm.GasCost) {

}

func (fa *freezeAccount) IsInterfaceNil() bool {
	return fa == nil
}
