//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. delegation.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const delegationManagementKey = "delegationManagement"
const delegationContractsList = "delegationContracts"

var nextAddressAdd = big.NewInt(1 << 24)

type delegationManager struct {
	eei                    vm.SystemEI
	delegationMgrSCAddress []byte
	stakingSCAddr          []byte
	validatorSCAddr        []byte
	configChangeAddr       []byte
	gasCost                vm.GasCost
	marshalizer            marshal.Marshalizer
	minCreationDeposit     *big.Int
	minDelegationAmount    *big.Int
	minFee                 uint64
	maxFee                 uint64
	enableEpochsHandler    common.EnableEpochsHandler
	mutExecution           sync.RWMutex
}

// ArgsNewDelegationManager defines the arguments to create the delegation manager system smart contract
type ArgsNewDelegationManager struct {
	DelegationMgrSCConfig  config.DelegationManagerSystemSCConfig
	DelegationSCConfig     config.DelegationSystemSCConfig
	Eei                    vm.SystemEI
	DelegationMgrSCAddress []byte
	StakingSCAddress       []byte
	ValidatorSCAddress     []byte
	ConfigChangeAddress    []byte
	GasCost                vm.GasCost
	Marshalizer            marshal.Marshalizer
	EnableEpochsHandler    common.EnableEpochsHandler
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
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
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
		eei:                    args.Eei,
		stakingSCAddr:          args.StakingSCAddress,
		validatorSCAddr:        args.ValidatorSCAddress,
		delegationMgrSCAddress: args.DelegationMgrSCAddress,
		configChangeAddr:       args.ConfigChangeAddress,
		gasCost:                args.GasCost,
		marshalizer:            args.Marshalizer,
		minCreationDeposit:     minCreationDeposit,
		minDelegationAmount:    minDelegationAmount,
		minFee:                 args.DelegationSCConfig.MinServiceFee,
		maxFee:                 args.DelegationSCConfig.MaxServiceFee,
		enableEpochsHandler:    args.EnableEpochsHandler,
	}

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

	currentEpoch := d.enableEpochsHandler.GetCurrentEpoch()
	if !d.enableEpochsHandler.IsDelegationManagerFlagEnabledInEpoch(currentEpoch) {
		d.eei.AddReturnMessage("delegation manager contract is not enabled")
		return vmcommon.UserError
	}

	if len(args.ESDTTransfers) > 0 {
		d.eei.AddReturnMessage("cannot transfer ESDT to system SCs")
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
	case "makeNewContractFromValidatorData":
		return d.makeNewContractFromValidatorData(args)
	case "mergeValidatorToDelegationSameOwner":
		return d.mergeValidatorToDelegation(args, d.checkCallerIsOwnerOfContract)
	case "mergeValidatorToDelegationWithWhitelist":
		return d.mergeValidatorToDelegation(args, d.isAddressWhiteListedForMerge)
	case "claimMulti":
		return d.claimMulti(args)
	case "reDelegateMulti":
		return d.reDelegateMulti(args)
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
	err := saveDelegationManagementData(d.eei, d.marshalizer, d.delegationMgrSCAddress, managementData)
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

	_, returnCode := d.deployNewContract(args, true, core.SCDeployInitFunctionName, args.CallerAddr, args.CallValue, args.Arguments)

	return returnCode
}

func (d *delegationManager) deployNewContract(
	args *vmcommon.ContractCallInput,
	checkMinDeposit bool,
	initFunction string,
	deployerAddr []byte,
	depositValue *big.Int,
	arguments [][]byte,
) ([]byte, vmcommon.ReturnCode) {
	delegationManagement, err := d.getDelegationManagementData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}

	minValue := big.NewInt(0).Set(delegationManagement.MinDeposit)
	if args.CallValue.Cmp(minValue) < 0 && checkMinDeposit {
		d.eei.AddReturnMessage("not enough call value")
		return nil, vmcommon.UserError
	}

	delegationList, err := getDelegationContractList(d.eei, d.marshalizer, d.delegationMgrSCAddress)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}

	newAddress := createNewAddress(delegationManagement.LastAddress)

	returnCode, err := d.eei.DeploySystemSC(vm.FirstDelegationSCAddress, newAddress, deployerAddr, initFunction, depositValue, arguments)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}
	if returnCode != vmcommon.Ok {
		return nil, returnCode
	}

	delegationManagement.NumOfContracts += 1
	delegationManagement.LastAddress = newAddress
	delegationList.Addresses = append(delegationList.Addresses, newAddress)

	currentStorage := d.eei.GetStorage(args.CallerAddr)
	currentStorage = append(currentStorage, newAddress...)

	d.eei.SetStorage(args.CallerAddr, currentStorage)
	err = saveDelegationManagementData(d.eei, d.marshalizer, d.delegationMgrSCAddress, delegationManagement)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}

	err = d.saveDelegationContractList(delegationList)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}

	d.eei.Finish(newAddress)

	return newAddress, vmcommon.Ok
}

func (d *delegationManager) correctOwnerOnAccount(newAddress []byte, caller []byte) error {
	if !d.enableEpochsHandler.FixDelegationChangeOwnerOnAccountEnabled() {
		return nil // backwards compatibility
	}

	return d.eei.UpdateCodeDeployerAddress(string(newAddress), caller)
}

func (d *delegationManager) makeNewContractFromValidatorData(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.checkValidatorToDelegationInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) != 2 {
		d.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.UserError
	}
	if d.callerAlreadyDeployed(args.CallerAddr) {
		d.eei.AddReturnMessage("caller already deployed a delegation sc")
		return vmcommon.UserError
	}

	arguments := append([][]byte{args.CallerAddr}, args.Arguments...)
	newAddress, returnCode := d.deployNewContract(args, false, initFromValidatorData, d.delegationMgrSCAddress, big.NewInt(0), arguments)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err := d.correctOwnerOnAccount(newAddress, args.CallerAddr)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegationManager) checkValidatorToDelegationInput(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	currentEpoch := d.enableEpochsHandler.GetCurrentEpoch()
	if !d.enableEpochsHandler.IsValidatorToDelegationFlagEnabledInEpoch(currentEpoch) {
		d.eei.AddReturnMessage("invalid function to call")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if core.IsSmartContractAddress(args.CallerAddr) {
		d.eei.AddReturnMessage("cannot change from validator to delegation contract for a smart contract")
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegationManager) checkCallerIsOwnerOfContract(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	scAddress := args.Arguments[0]
	buff := d.eei.GetStorage(args.CallerAddr)
	if len(buff) == 0 {
		d.eei.AddReturnMessage("the caller does not own a delegation sc")
		return vmcommon.UserError
	}

	found := false
	lenAddress := len(args.CallerAddr)
	for i := 0; i < len(buff); i += lenAddress {
		savedAddress := buff[i : i+lenAddress]
		if bytes.Equal(savedAddress, scAddress) {
			found = true
			break
		}
	}

	if !found {
		d.eei.AddReturnMessage("did not find delegation contract with given address for this caller")
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (d *delegationManager) mergeValidatorToDelegation(
	args *vmcommon.ContractCallInput,
	checkFunc func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode,
) vmcommon.ReturnCode {
	returnCode := d.checkValidatorToDelegationInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if len(args.Arguments) != 1 {
		d.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.UserError
	}
	lenAddress := len(args.CallerAddr)
	scAddress := args.Arguments[0]
	if len(scAddress) != lenAddress {
		d.eei.AddReturnMessage("invalid argument, wanted an address")
		return vmcommon.UserError
	}

	returnCode = checkFunc(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	txData := mergeValidatorDataToDelegation + "@" + hex.EncodeToString(args.CallerAddr)
	vmOutput, err := d.eei.ExecuteOnDestContext(scAddress, d.delegationMgrSCAddress, big.NewInt(0), []byte(txData))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	txData = deleteWhitelistForMerge
	vmOutput, err = d.eei.ExecuteOnDestContext(scAddress, d.delegationMgrSCAddress, big.NewInt(0), []byte(txData))
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmOutput.ReturnCode
}

func (d *delegationManager) isAddressWhiteListedForMerge(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	scAddress := args.Arguments[0]
	savedAddress := d.eei.GetStorageFromAddress(scAddress, []byte(whitelistedAddress))
	if !bytes.Equal(savedAddress, args.CallerAddr) {
		d.eei.AddReturnMessage("address is not whitelisted for merge")
		return vmcommon.UserError
	}
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
	err = saveDelegationManagementData(d.eei, d.marshalizer, d.delegationMgrSCAddress, delegationManagement)
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
	err = saveDelegationManagementData(d.eei, d.marshalizer, d.delegationMgrSCAddress, delegationManagement)
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

	contractList, err := getDelegationContractList(d.eei, d.marshalizer, d.delegationMgrSCAddress)
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

func (d *delegationManager) claimMulti(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.executeFuncOnListAddresses(args, claimRewards)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	totalSent := d.eei.GetTotalSentToUser(args.CallerAddr)
	d.eei.Finish(totalSent.Bytes())

	return vmcommon.Ok
}

func (d *delegationManager) reDelegateMulti(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := d.executeFuncOnListAddresses(args, reDelegateRewards)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	logs := d.eei.GetLogs()
	totalReDelegated := getTotalReDelegatedFromLogs(logs)
	d.eei.Finish(totalReDelegated.Bytes())

	return vmcommon.Ok
}

func getTotalReDelegatedFromLogs(logs []*vmcommon.LogEntry) *big.Int {
	totalReDelegated := big.NewInt(0)
	for _, reDelegateLog := range logs {
		if len(reDelegateLog.Topics) < 1 {
			continue
		}
		if !bytes.Equal(reDelegateLog.Identifier, []byte(delegate)) {
			continue
		}
		valueFromFirstTopic := big.NewInt(0).SetBytes(reDelegateLog.Topics[0])
		totalReDelegated.Add(totalReDelegated, valueFromFirstTopic)
	}

	return totalReDelegated
}

func (d *delegationManager) executeFuncOnListAddresses(
	args *vmcommon.ContractCallInput,
	funcName string,
) vmcommon.ReturnCode {
	if !d.enableEpochsHandler.IsMultiClaimOnDelegationEnabled() {
		d.eei.AddReturnMessage("invalid function to call")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		d.eei.AddReturnMessage(vm.ErrInvalidNumOfArguments.Error())
		return vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	mapAddresses := make(map[string]struct{})
	var vmOutput *vmcommon.VMOutput
	var found bool
	for _, address := range args.Arguments {
		if len(address) != len(args.CallerAddr) {
			d.eei.AddReturnMessage(vm.ErrInvalidArgument.Error())
			return vmcommon.UserError
		}
		_, found = mapAddresses[string(address)]
		if found {
			d.eei.AddReturnMessage("duplicated input")
			return vmcommon.UserError
		}

		mapAddresses[string(address)] = struct{}{}
		vmOutput, err = d.eei.ExecuteOnDestContext(address, args.CallerAddr, big.NewInt(0), []byte(funcName))
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			return vmOutput.ReturnCode
		}
	}

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

func saveDelegationManagementData(
	eei vm.SystemEI,
	marshalizer marshal.Marshalizer,
	delegationMgrAddress []byte,
	managementData *DelegationManagement,
) error {
	marshaledData, err := marshalizer.Marshal(managementData)
	if err != nil {
		return err
	}

	eei.SetStorageForAddress(delegationMgrAddress, []byte(delegationManagementKey), marshaledData)
	return nil
}

func getDelegationContractList(
	eei vm.SystemEI,
	marshalizer marshal.Marshalizer,
	delegationMgrAddress []byte,
) (*DelegationContractList, error) {
	marshaledData := eei.GetStorageFromAddress(delegationMgrAddress, []byte(delegationContractsList))
	if len(marshaledData) == 0 {
		return nil, fmt.Errorf("%w getDelegationContractList", vm.ErrDataNotFoundUnderKey)
	}

	contractList := &DelegationContractList{}
	err := marshalizer.Unmarshal(contractList, marshaledData)
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

// CanUseContract returns true if contract can be used
func (d *delegationManager) CanUseContract() bool {
	return d.enableEpochsHandler.IsDelegationManagerFlagEnabledInEpoch(d.enableEpochsHandler.GetCurrentEpoch())
}

// IsInterfaceNil returns true if underlying object is nil
func (d *delegationManager) IsInterfaceNil() bool {
	return d == nil
}
