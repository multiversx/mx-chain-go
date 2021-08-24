//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. liquidStaking.proto
package systemSmartContracts

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const tokenIDKey = "tokenID"
const noncePrefix = "n"

type liquidStaking struct {
	eei                      vm.SystemEI
	sigVerifier              vm.MessageSignVerifier
	delegationMgrSCAddress   []byte
	liquidStakingSCAddress   []byte
	endOfEpochAddr           []byte
	gasCost                  vm.GasCost
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	mutExecution             sync.RWMutex
	liquidStakingEnableEpoch uint32
	flagLiquidStaking        atomic.Flag
}

// ArgsNewLiquidStaking defines the arguments to create the liquid staking smart contract
type ArgsNewLiquidStaking struct {
	EpochConfig            config.EpochConfig
	Eei                    vm.SystemEI
	DelegationMgrSCAddress []byte
	LiquidStakingSCAddress []byte
	EndOfEpochAddress      []byte
	GasCost                vm.GasCost
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	EpochNotifier          vm.EpochNotifier
}

// NewLiquidStakingSystemSC creates a new liquid staking system SC
func NewLiquidStakingSystemSC(args ArgsNewLiquidStaking) (*liquidStaking, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation manager sc address", vm.ErrInvalidAddress)
	}
	if len(args.EndOfEpochAddress) < 1 {
		return nil, fmt.Errorf("%w for end of epoch address", vm.ErrInvalidAddress)
	}
	if len(args.LiquidStakingSCAddress) < 1 {
		return nil, fmt.Errorf("%w for liquid staking sc address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, vm.ErrNilHasher
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, vm.ErrNilEpochNotifier
	}

	l := &liquidStaking{
		eei:                      args.Eei,
		delegationMgrSCAddress:   args.DelegationMgrSCAddress,
		endOfEpochAddr:           args.EndOfEpochAddress,
		liquidStakingSCAddress:   args.LiquidStakingSCAddress,
		gasCost:                  args.GasCost,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		liquidStakingEnableEpoch: args.EpochConfig.EnableEpochs.BuiltInFunctionOnMetaEnableEpoch,
	}
	log.Debug("liquid staking: enable epoch", "epoch", l.liquidStakingEnableEpoch)

	args.EpochNotifier.RegisterNotifyHandler(l)

	return l, nil
}

// Execute calls one of the functions from the delegation contract and runs the code according to the input
func (l *liquidStaking) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	l.mutExecution.RLock()
	defer l.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		l.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if !l.flagLiquidStaking.IsSet() {
		l.eei.AddReturnMessage("liquid staking contract is not enabled")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return l.init(args)
	case "claimDelegatedPosition":
		return l.claimDelegatedPosition(args)
	case "claimRewardsFromPosition":
		return l.claimRewardsFromDelegatedPosition(args)
	case "reDelegateRewardsFromPosition":
		return l.reDelegateRewardsFromPosition(args)
	case "unDelegateWithPosition":
		return l.unDelegateWithPosition(args)
	case "returnPosition":
		return l.returnPosition(args)
	}

	l.eei.AddReturnMessage(args.Function + " is an unknown function")
	return vmcommon.UserError
}

func (l *liquidStaking) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, l.liquidStakingSCAddress) {
		l.eei.AddReturnMessage("invalid caller")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		l.eei.AddReturnMessage("not a payable function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		l.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.UserError
	}
	tokenID := args.Arguments[0]
	l.eei.SetStorage([]byte(tokenIDKey), tokenID)

	return vmcommon.Ok
}

func (l *liquidStaking) getTokenID() []byte {
	return l.eei.GetStorage([]byte(tokenIDKey))
}

func (l *liquidStaking) checkArgumentsWhenPositionIsInput(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.ESDTTransfers) < 1 {
		l.eei.AddReturnMessage("function requires liquid staking input")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		l.eei.AddReturnMessage("function is not payable in eGLD")
		return vmcommon.UserError
	}
	for _, esdtTransfer := range args.ESDTTransfers {
		if !bytes.Equal(esdtTransfer.ESDTTokenName, l.getTokenID()) {
			l.eei.AddReturnMessage("wrong liquid staking position as input")
			return vmcommon.UserError
		}
	}
	err := l.eei.UseGas(uint64(len(args.ESDTTransfers)) * l.gasCost.MetaChainSystemSCsCost.LiquidStakingOps)
	if err != nil {
		l.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	return vmcommon.Ok
}

func (l *liquidStaking) claimDelegatedPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		l.eei.AddReturnMessage("function is not payable in eGLD")
		return vmcommon.UserError
	}
	if len(args.Arguments) == 0 {
		l.eei.AddReturnMessage("not enough arguments")
		return vmcommon.UserError
	}
	if len(args.ESDTTransfers) > 0 {
		l.eei.AddReturnMessage("function is not payable in ESDT")
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (l *liquidStaking) claimRewardsFromDelegatedPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := l.checkArgumentsWhenPositionIsInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	return vmcommon.Ok
}

func (l *liquidStaking) reDelegateRewardsFromPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := l.checkArgumentsWhenPositionIsInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (l *liquidStaking) unDelegateWithPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := l.checkArgumentsWhenPositionIsInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (l *liquidStaking) returnPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := l.checkArgumentsWhenPositionIsInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

// SetNewGasCost is called whenever a gas cost was changed
func (l *liquidStaking) SetNewGasCost(gasCost vm.GasCost) {
	l.mutExecution.Lock()
	l.gasCost = gasCost
	l.mutExecution.Unlock()
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (l *liquidStaking) EpochConfirmed(epoch uint32, _ uint64) {
	l.flagLiquidStaking.Toggle(epoch >= l.liquidStakingEnableEpoch)
	log.Debug("liquid staking system sc", "enabled", l.flagLiquidStaking.IsSet())
}

// CanUseContract returns true if contract can be used
func (l *liquidStaking) CanUseContract() bool {
	return l.flagLiquidStaking.IsSet()
}

// IsInterfaceNil returns true if underlying object is nil
func (l *liquidStaking) IsInterfaceNil() bool {
	return l == nil
}
