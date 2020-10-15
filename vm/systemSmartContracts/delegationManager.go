//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. delegation.proto
package systemSmartContracts

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const delegationManagmentKey = "delegationManagement"
const delegationContractsList = "delegationContracts"

type delegationManager struct {
	eei                      vm.SystemEI
	delegationMgrSCAddress   []byte
	stakingSCAddr            []byte
	auctionSCAddr            []byte
	gasCost                  vm.GasCost
	marshalizer              marshal.Marshalizer
	delegationMgrEnabled     atomic.Flag
	enableDelegationMgrEpoch uint32
	baseIssuingCost          *big.Int
}

// ArgsNewDelegationManager defines the arguments to create the delegation manager system smart contract
type ArgsNewDelegationManager struct {
	DelegationMgrSCConfig  config.DelegationManagerSystemSCConfig
	Eei                    vm.SystemEI
	DelegationMgrSCAddress []byte
	StakingSCAddress       []byte
	AuctionSCAddress       []byte
	GasCost                vm.GasCost
	Marshalizer            marshal.Marshalizer
	EpochNotifier          vm.EpochNotifier
}

// NewDelegationManagerSystemSC creates a new delegation manager system SC
func NewDelegationManagerSystemSC(args ArgsNewDelegationManager) (*delegationManager, error) {
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

	baseIssuingCost, okConvert := big.NewInt(0).SetString(args.DelegationMgrSCConfig.BaseIssuingCost, conversionBase)
	if !okConvert || baseIssuingCost.Cmp(zero) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	d := &delegationManager{
		eei:                      args.Eei,
		stakingSCAddr:            args.StakingSCAddress,
		auctionSCAddr:            args.AuctionSCAddress,
		delegationMgrSCAddress:   args.DelegationMgrSCAddress,
		gasCost:                  args.GasCost,
		marshalizer:              args.Marshalizer,
		delegationMgrEnabled:     atomic.Flag{},
		enableDelegationMgrEpoch: args.DelegationMgrSCConfig.EnabledEpoch,
		baseIssuingCost:          baseIssuingCost,
	}

	args.EpochNotifier.RegisterNotifyHandler(d)

	return d, nil
}

// Execute  calls one of the functions from the delegation manager contract and runs the code according to the input
func (d *delegationManager) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if !d.delegationMgrEnabled.IsSet() {
		d.eei.AddReturnMessage("delegation manager contract is not enabled")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return d.init(args)
	case "createNewDelegationContract":
		return d.createNewDelegationContract(args)
	case "getAllContractAddresses":
		return d.getAllContractAddresses(args)
	}

	return vmcommon.UserError
}

func (d *delegationManager) getAllContractAddresses(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegationManager) createNewDelegationContract(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (d *delegationManager) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.UserError
	}

	managementData := &DelegationManagement{
		NumberOfContract:     0,
		LastAddress:          nil,
		MinContractCreateFee: nil,
		MinServiceFee:        0,
		MaxServiceFee:        math.MaxUint64,
		MinOperationValue:    nil,
	}
	err := d.saveDelegationManagementData(managementData)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegationManager) getDelegationManagementData() (*DelegationManagement, error) {
	marshalledData := d.eei.GetStorage([]byte(delegationManagmentKey))
	if len(marshalledData) == 0 {
		return nil, fmt.Errorf("%w getDelegationManagementData", vm.ErrDataNotFoundUnderKey)
	}

	managementData := &DelegationManagement{}
	err := d.marshalizer.Unmarshal(managementData, marshalledData)
	if err != nil {
		return nil, err
	}
	return managementData, nil
}

func (d *delegationManager) saveDelegationManagementData(managementData *DelegationManagement) error {
	marshalledData, err := d.marshalizer.Marshal(managementData)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(delegationManagmentKey), marshalledData)
	return nil
}

func (d *delegationManager) getDelegationContractList() (*DelegationContractList, error) {
	marshalledData := d.eei.GetStorage([]byte(delegationContractsList))
	if len(marshalledData) == 0 {
		return nil, fmt.Errorf("%w getDelegationContractList", vm.ErrDataNotFoundUnderKey)
	}

	contractList := &DelegationContractList{}
	err := d.marshalizer.Unmarshal(contractList, marshalledData)
	if err != nil {
		return nil, err
	}
	return contractList, nil
}

func (d *delegationManager) saveDelegationContractList(list *DelegationContractList) error {
	marshalledData, err := d.marshalizer.Marshal(list)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(delegationContractsList), marshalledData)
	return nil
}

// EpochConfirmed  is called whenever a new epoch is confirmed
func (d *delegationManager) EpochConfirmed(epoch uint32) {
	d.delegationMgrEnabled.Toggle(epoch >= d.enableDelegationMgrEpoch)
	log.Debug("delegationManager", "enabled", d.delegationMgrEnabled.IsSet())
}

// IsContractEnabled returns true if contract can be used
func (d *delegationManager) IsContractEnabled() bool {
	return d.delegationMgrEnabled.IsSet()
}

// IsInterfaceNil returns true if underlying object is nil
func (d *delegationManager) IsInterfaceNil() bool {
	return d == nil
}
