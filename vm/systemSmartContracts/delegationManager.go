//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. delegation.proto
package systemSmartContracts

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
)

const delegationManagementKey = "delegationManagement"
const delegationContractsList = "delegationContracts"

var nextAddressAdd = big.NewInt(1 << 24)

type delegationManager struct {
	eei                      vm.SystemEI
	delegationMgrSCAddress   []byte
	stakingSCAddr            []byte
	validatorSCAddr          []byte
	configChangeAddr         []byte
	gasCost                  vm.GasCost
	marshalizer              marshal.Marshalizer
	delegationMgrEnabled     atomic.Flag
	enableDelegationMgrEpoch uint32
	minCreationDeposit       *big.Int
	minDelegationAmount      *big.Int
	minFee                   uint64
	maxFee                   uint64
	mutExecution             sync.RWMutex
}

// ArgsNewDelegationManager defines the arguments to create the delegation manager system smart contract
type ArgsNewDelegationManager struct {
	DelegationMgrSCConfig  config.DelegationManagerSystemSCConfig
	DelegationSCConfig     config.DelegationSystemSCConfig
	EpochConfig            config.EpochConfig
	Eei                    vm.SystemEI
	DelegationMgrSCAddress []byte
	StakingSCAddress       []byte
	ValidatorSCAddress     []byte
	ConfigChangeAddress    []byte
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
	if len(args.ValidatorSCAddress) < 1 {
		return nil, fmt.Errorf("%w for validator sc address", vm.ErrInvalidAddress)
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	}
	if len(args.ConfigChangeAddress) < 1 {
		return nil, fmt.Errorf("%w for config change address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, vm.ErrNilEpochNotifier
	}

	minCreationDeposit, okConvert := big.NewInt(0).SetString(args.DelegationMgrSCConfig.MinCreationDeposit, conversionBase)
	if !okConvert || minCreationDeposit.Cmp(zero) < 0 {
		return nil, vm.ErrInvalidMinCreationDeposit
	}

	minDelegationAmount, okConvert := big.NewInt(0).SetString(args.DelegationMgrSCConfig.MinStakeAmount, conversionBase)
	if !okConvert || minDelegationAmount.Cmp(zero) <= 0 {
		return nil, vm.ErrInvalidMinStakeValue
	}

	d := &delegationManager{
		eei:                      args.Eei,
		stakingSCAddr:            args.StakingSCAddress,
		validatorSCAddr:          args.ValidatorSCAddress,
		delegationMgrSCAddress:   args.DelegationMgrSCAddress,
		configChangeAddr:         args.ConfigChangeAddress,
		gasCost:                  args.GasCost,
		marshalizer:              args.Marshalizer,
		delegationMgrEnabled:     atomic.Flag{},
		enableDelegationMgrEpoch: args.DelegationMgrSCConfig.EnabledEpoch,
		minCreationDeposit:       minCreationDeposit,
		minDelegationAmount:      minDelegationAmount,
		minFee:                   args.DelegationSCConfig.MinServiceFee,
		maxFee:                   args.DelegationSCConfig.MaxServiceFee,
	}

	args.EpochNotifier.RegisterNotifyHandler(d)

	return d, nil
}

// Execute calls one of the functions from the delegation manager contract and runs the code according to the input
func (d *delegationManager) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	d.mutExecution.RLock()
	defer d.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
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
	case "getContractConfig":
		return d.getContractConfig(args)
	case "changeMinDeposit":
		return d.changeMinDeposit(args)
	case "changeMinDelegationAmount":
		return d.changeMinDelegationAmount(args)
	}

	d.eei.AddReturnMessage("invalid function to call")
	return vmcommon.UserError
}

func (d *delegationManager) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}

	managementData := &DelegationManagement{
		NumOfContracts:      0,
		LastAddress:         vm.FirstDelegationSCAddress,
		MinServiceFee:       d.minFee,
		MaxServiceFee:       d.maxFee,
		MinDeposit:          d.minCreationDeposit,
		MinDelegationAmount: d.minDelegationAmount,
	}
	err := d.saveDelegationManagementData(managementData)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	delegationList := &DelegationContractList{Addresses: [][]byte{vm.FirstDelegationSCAddress}}
	err = d.saveDelegationContractList(delegationList)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegationManager) createNewDelegationContract(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 2 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}

	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationMgrOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if d.callerAlreadyDeployed(args.CallerAddr) {
		d.eei.AddReturnMessage("caller already deployed a delegation sc")
		return vmcommon.UserError
	}

	delegationManagement, err := d.getDelegationManagementData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	minValue := big.NewInt(0).Set(delegationManagement.MinDeposit)
	if args.CallValue.Cmp(minValue) < 0 {
		d.eei.AddReturnMessage("not enough call value")
		return vmcommon.UserError
	}

	delegationList, err := d.getDelegationContractList()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	depositValue := big.NewInt(0).Set(args.CallValue)
	newAddress := createNewAddress(delegationManagement.LastAddress)

	returnCode, err := d.eei.DeploySystemSC(vm.FirstDelegationSCAddress, newAddress, args.CallerAddr, depositValue, args.Arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	delegationManagement.NumOfContracts += 1
	delegationManagement.LastAddress = newAddress
	delegationList.Addresses = append(delegationList.Addresses, newAddress)

	d.eei.SetStorage(args.CallerAddr, newAddress)
	err = d.saveDelegationManagementData(delegationManagement)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.saveDelegationContractList(delegationList)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(newAddress)

	return vmcommon.Ok
}

func (d *delegationManager) checkConfigChangeInput(args *vmcommon.ContractCallInput) error {
	if args.CallValue.Cmp(zero) != 0 {
		return vm.ErrCallValueMustBeZero
	}
	if len(args.Arguments) != 1 {
		return vm.ErrInvalidNumOfArguments
	}
	if !bytes.Equal(args.CallerAddr, d.configChangeAddr) {
		return vm.ErrInvalidCaller
	}
	return nil
}

func (d *delegationManager) changeMinDeposit(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := d.checkConfigChangeInput(args)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	delegationManagement, err := d.getDelegationManagementData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	minDeposit := big.NewInt(0).SetBytes(args.Arguments[0])
	if minDeposit.Cmp(zero) < 0 {
		d.eei.AddReturnMessage("invalid min deposit")
		return vmcommon.UserError
	}
	delegationManagement.MinDeposit = minDeposit
	err = d.saveDelegationManagementData(delegationManagement)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegationManager) changeMinDelegationAmount(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := d.checkConfigChangeInput(args)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	delegationManagement, err := d.getDelegationManagementData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	minDelegationAmount := big.NewInt(0).SetBytes(args.Arguments[0])
	if minDelegationAmount.Cmp(zero) <= 0 {
		d.eei.AddReturnMessage("invalid min delegation amount")
		return vmcommon.UserError
	}
	delegationManagement.MinDelegationAmount = minDelegationAmount
	err = d.saveDelegationManagementData(delegationManagement)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegationManager) getAllContractAddresses(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, d.delegationMgrSCAddress) {
		d.eei.AddReturnMessage(vm.ErrInvalidCaller.Error())
		return vmcommon.UserError
	}

	contractList, err := d.getDelegationContractList()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if len(contractList.Addresses) == 0 {
		return vmcommon.Ok
	}

	for _, address := range contractList.Addresses[1:] {
		d.eei.Finish(address)
	}

	return vmcommon.Ok
}

func (d *delegationManager) getContractConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, d.delegationMgrSCAddress) {
		d.eei.AddReturnMessage(vm.ErrInvalidCaller.Error())
		return vmcommon.UserError
	}

	cfg, err := d.getDelegationManagementData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.eei.Finish(big.NewInt(0).SetUint64(uint64(cfg.NumOfContracts)).Bytes())
	d.eei.Finish(cfg.LastAddress)
	d.eei.Finish(big.NewInt(0).SetUint64(cfg.MinServiceFee).Bytes())
	d.eei.Finish(big.NewInt(0).SetUint64(cfg.MaxServiceFee).Bytes())
	d.eei.Finish(cfg.MinDeposit.Bytes())
	d.eei.Finish(cfg.MinDelegationAmount.Bytes())

	return vmcommon.Ok
}

func createNewAddress(lastAddress []byte) []byte {
	i := 0
	for ; i < len(lastAddress) && lastAddress[i] == 0; i++ {
	}
	prefixZeros := make([]byte, i)

	lastAddressAsBigInt := big.NewInt(0).SetBytes(lastAddress)
	lastAddressAsBigInt.Add(lastAddressAsBigInt, nextAddressAdd)

	newAddress := append(prefixZeros, lastAddressAsBigInt.Bytes()...)
	return newAddress
}

func (d *delegationManager) callerAlreadyDeployed(address []byte) bool {
	return len(d.eei.GetStorage(address)) > 0
}

func (d *delegationManager) getDelegationManagementData() (*DelegationManagement, error) {
	marshaledData := d.eei.GetStorage([]byte(delegationManagementKey))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w getDelegationManagementData", vm.ErrDataNotFoundUnderKey)
	}

	managementData := &DelegationManagement{}
	err := d.marshalizer.Unmarshal(managementData, marshaledData)
	if err != nil {
		return nil, err
	}
	return managementData, nil
}

func (d *delegationManager) saveDelegationManagementData(managementData *DelegationManagement) error {
	marshaledData, err := d.marshalizer.Marshal(managementData)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(delegationManagementKey), marshaledData)
	return nil
}

func (d *delegationManager) getDelegationContractList() (*DelegationContractList, error) {
	marshaledData := d.eei.GetStorage([]byte(delegationContractsList))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w getDelegationContractList", vm.ErrDataNotFoundUnderKey)
	}

	contractList := &DelegationContractList{}
	err := d.marshalizer.Unmarshal(contractList, marshaledData)
	if err != nil {
		return nil, err
	}
	return contractList, nil
}

func (d *delegationManager) saveDelegationContractList(list *DelegationContractList) error {
	marshaledData, err := d.marshalizer.Marshal(list)
	if err != nil {
		return err
	}

	d.eei.SetStorage([]byte(delegationContractsList), marshaledData)
	return nil
}

// SetNewGasCost is called whenever a gas cost was changed
func (d *delegationManager) SetNewGasCost(gasCost vm.GasCost) {
	d.mutExecution.Lock()
	d.gasCost = gasCost
	d.mutExecution.Unlock()
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (d *delegationManager) EpochConfirmed(epoch uint32) {
	d.delegationMgrEnabled.Toggle(epoch >= d.enableDelegationMgrEpoch)
	log.Debug("delegationManager", "enabled", d.delegationMgrEnabled.IsSet())
}

// CanUseContract returns true if contract can be used
func (d *delegationManager) CanUseContract() bool {
	return d.delegationMgrEnabled.IsSet()
}

// IsInterfaceNil returns true if underlying object is nil
func (d *delegationManager) IsInterfaceNil() bool {
	return d == nil
}
