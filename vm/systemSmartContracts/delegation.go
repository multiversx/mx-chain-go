package systemSmartContracts

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type delegation struct {
	eei                    vm.SystemEI
	delegationMgrSCAddress []byte
	stakingSCAddr          []byte
	auctionSCAddr          []byte
	gasCost                vm.GasCost
	marshalizer            marshal.Marshalizer
	delegationEnabled      atomic.Flag
	enableDelegationEpoch  uint32
}

// ArgsNewDelegation -
type ArgsNewDelegation struct {
	DelegationSCConfig     config.DelegationSystemSCConfig
	Eei                    vm.SystemEI
	DelegationMgrSCAddress []byte
	StakingSCAddress       []byte
	AuctionSCAddress       []byte
	GasCost                vm.GasCost
	Marshalizer            marshal.Marshalizer
	EpochNotifier          vm.EpochNotifier
}

// NewDelegationSystemSC creates a new delegation system SC
func NewDelegationSystemSC(args ArgsNewDelegation) (*delegation, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingSCAddress) < 1 {
		return nil, fmt.Errorf("%w for staking sc address", vm.ErrInvalidAddress)
	}
	if len(args.AuctionSCAddress) < 1 {
		return nil, fmt.Errorf("%w for auction sc address", vm.ErrInvalidAddress)
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, vm.ErrNilEpochNotifier
	}

	d := &delegation{
		eei:                    args.Eei,
		stakingSCAddr:          args.StakingSCAddress,
		auctionSCAddr:          args.AuctionSCAddress,
		delegationMgrSCAddress: args.DelegationMgrSCAddress,
		gasCost:                args.GasCost,
		marshalizer:            args.Marshalizer,
		delegationEnabled:      atomic.Flag{},
		enableDelegationEpoch:  args.DelegationSCConfig.EnabledEpoch,
	}

	args.EpochNotifier.RegisterNotifyHandler(d)

	return d, nil
}

// Execute  calls one of the functions from the delegation manager contract and runs the code according to the input
func (d *delegation) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if !d.delegationEnabled.IsSet() {
		d.eei.AddReturnMessage("delegation manager contract is not enabled")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return d.init(args)
	}

	return vmcommon.UserError
}

func (d *delegation) init(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

// EpochConfirmed  is called whenever a new epoch is confirmed
func (d *delegation) EpochConfirmed(epoch uint32) {
	d.delegationEnabled.Toggle(epoch >= d.enableDelegationEpoch)
	log.Debug("delegationManager", "enabled", d.delegationEnabled.IsSet())
}

// IsContractEnabled returns true if contract can be used
func (d *delegation) IsContractEnabled() bool {
	return d.delegationEnabled.IsSet()
}

// IsInterfaceNil returns true if underlying object is nil
func (d *delegation) IsInterfaceNil() bool {
	return d == nil
}
