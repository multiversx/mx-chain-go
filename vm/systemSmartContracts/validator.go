//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. validator.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
)

const unJailedFunds = "unJailFunds"
const unStakeUnBondPauseKey = "unStakeUnBondPause"

var zero = big.NewInt(0)

// return codes for each input blskey
const (
	ok uint8 = iota
	invalidKey
	failed
	waiting
)

type validatorSC struct {
	eei                   vm.SystemEI
	unBondPeriod          uint64
	sigVerifier           vm.MessageSignVerifier
	baseConfig            ValidatorConfig
	stakingV2Epoch        uint32
	stakingSCAddress      []byte
	validatorSCAddress    []byte
	walletAddressLen      int
	enableStakingEpoch    uint32
	enableDoubleKeyEpoch  uint32
	gasCost               vm.GasCost
	marshalizer           marshal.Marshalizer
	flagEnableStaking     atomic.Flag
	flagEnableTopUp       atomic.Flag
	flagDoubleKey         atomic.Flag
	minUnstakeTokensValue *big.Int
	minDeposit            *big.Int
	mutExecution          sync.RWMutex
	endOfEpochAddress     []byte
}

// ArgsValidatorSmartContract is the arguments structure to create a new ValidatorSmartContract
type ArgsValidatorSmartContract struct {
	StakingSCConfig    config.StakingSystemSCConfig
	GenesisTotalSupply *big.Int
	Eei                vm.SystemEI
	SigVerifier        vm.MessageSignVerifier
	StakingSCAddress   []byte
	ValidatorSCAddress []byte
	GasCost            vm.GasCost
	Marshalizer        marshal.Marshalizer
	EpochNotifier      vm.EpochNotifier
	EndOfEpochAddress  []byte
	MinDeposit         string
}

// NewValidatorSmartContract creates an validator smart contract
func NewValidatorSmartContract(
	args ArgsValidatorSmartContract,
) (*validatorSC, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingSCAddress) == 0 {
		return nil, vm.ErrNilStakingSmartContractAddress
	}
	if len(args.ValidatorSCAddress) == 0 {
		return nil, vm.ErrNilValidatorSmartContractAddress
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.SigVerifier) {
		return nil, vm.ErrNilMessageSignVerifier
	}
	if args.GenesisTotalSupply == nil || args.GenesisTotalSupply.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidGenesisTotalSupply, args.GenesisTotalSupply)
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, vm.ErrNilEpochNotifier
	}
	if len(args.EndOfEpochAddress) < 1 {
		return nil, vm.ErrInvalidEndOfEpochAccessAddress
	}

	baseConfig := ValidatorConfig{
		TotalSupply: big.NewInt(0).Set(args.GenesisTotalSupply),
	}

	okValue := true
	baseConfig.UnJailPrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.UnJailValue, conversionBase)
	if !okValue || baseConfig.UnJailPrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, args.StakingSCConfig.UnJailValue)
	}
	baseConfig.MinStakeValue, okValue = big.NewInt(0).SetString(args.StakingSCConfig.MinStakeValue, conversionBase)
	if !okValue || baseConfig.MinStakeValue.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, args.StakingSCConfig.MinStakeValue)
	}
	baseConfig.NodePrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !okValue || baseConfig.NodePrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, args.StakingSCConfig.GenesisNodePrice)
	}
	baseConfig.MinStep, okValue = big.NewInt(0).SetString(args.StakingSCConfig.MinStepValue, conversionBase)
	if !okValue || baseConfig.MinStep.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStepValue, args.StakingSCConfig.MinStepValue)
	}
	minUnstakeTokensValue, okValue := big.NewInt(0).SetString(args.StakingSCConfig.MinUnstakeTokensValue, conversionBase)
	if !okValue || minUnstakeTokensValue.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinUnstakeTokensValue, args.StakingSCConfig.MinUnstakeTokensValue)
	}
	minDeposit, okConvert := big.NewInt(0).SetString(args.MinDeposit, conversionBase)
	if !okConvert || minDeposit.Cmp(zero) < 0 {
		return nil, vm.ErrInvalidMinCreationDeposit
	}

	reg := &validatorSC{
		eei:                   args.Eei,
		unBondPeriod:          args.StakingSCConfig.UnBondPeriod,
		sigVerifier:           args.SigVerifier,
		baseConfig:            baseConfig,
		stakingV2Epoch:        args.StakingSCConfig.StakingV2Epoch,
		enableStakingEpoch:    args.StakingSCConfig.StakeEnableEpoch,
		stakingSCAddress:      args.StakingSCAddress,
		validatorSCAddress:    args.ValidatorSCAddress,
		gasCost:               args.GasCost,
		marshalizer:           args.Marshalizer,
		minUnstakeTokensValue: minUnstakeTokensValue,
		walletAddressLen:      len(args.ValidatorSCAddress),
		enableDoubleKeyEpoch:  args.StakingSCConfig.DoubleKeyProtectionEnableEpoch,
		endOfEpochAddress:     args.EndOfEpochAddress,
		minDeposit:            minDeposit,
	}

	args.EpochNotifier.RegisterNotifyHandler(reg)

	return reg, nil
}

// Execute calls one of the functions from the validator smart contract and runs the code according to the input
func (v *validatorSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	v.mutExecution.RLock()
	defer v.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		v.eei.AddReturnMessage("nil arguments: " + err.Error())
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return v.init(args)
	case "stake":
		return v.stake(args)
	case "unStake":
		return v.unStake(args)
	case "unStakeNodes":
		return v.unStakeNodes(args)
	case "unStakeTokens":
		return v.unStakeTokens(args)
	case "unBond":
		return v.unBond(args)
	case "unBondNodes":
		return v.unBondNodes(args)
	case "unBondTokens":
		return v.unBondTokens(args)
	case "claim":
		return v.claim(args)
	case "get":
		return v.get(args)
	case "setConfig":
		return v.setConfig(args)
	case "changeRewardAddress":
		return v.changeRewardAddress(args)
	case "unJail":
		return v.unJail(args)
	case "getTotalStaked":
		return v.getTotalStaked(args)
	case "getTotalStakedTopUpStakedBlsKeys":
		return v.getTotalStakedTopUpStakedBlsKeys(args)
	case "getBlsKeysStatus":
		return v.getBlsKeysStatus(args)
	case "cleanRegisteredData":
		return v.cleanRegisteredData(args)
	case "pauseUnStakeUnBond":
		return v.pauseUnStakeUnBond(args)
	case "unPauseUnStakeUnBond":
		return v.unPauseStakeUnBond(args)
	case "getUnStakedTokensList":
		return v.getUnStakedTokensList(args)
	case "reStakeUnStakedNodes":
		return v.reStakeUnStakedNodes(args)
	}

	v.eei.AddReturnMessage("invalid method to call")
	return vmcommon.UserError
}

func (v *validatorSC) pauseUnStakeUnBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagEnableTopUp.IsSet() {
		v.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, v.endOfEpochAddress) {
		v.eei.AddReturnMessage("only end of epoch address can call")
		return vmcommon.UserError
	}

	v.eei.SetStorage([]byte(unStakeUnBondPauseKey), []byte{1})
	return vmcommon.Ok
}

func (v *validatorSC) unPauseStakeUnBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagEnableTopUp.IsSet() {
		v.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, v.endOfEpochAddress) {
		v.eei.AddReturnMessage("only end of epoch address can call")
		return vmcommon.UserError
	}

	v.eei.SetStorage([]byte(unStakeUnBondPauseKey), []byte{0})
	return vmcommon.Ok
}

func (v *validatorSC) isUnStakeUnBondPaused() bool {
	storageData := v.eei.GetStorage([]byte(unStakeUnBondPauseKey))
	if len(storageData) == 0 {
		return false
	}

	return storageData[0] == 1
}

func (v *validatorSC) addToUnJailFunds(value *big.Int) {
	currentValue := big.NewInt(0)
	storageData := v.eei.GetStorage([]byte(unJailedFunds))
	if len(storageData) > 0 {
		currentValue.SetBytes(storageData)
	}

	currentValue.Add(currentValue, value)
	v.eei.SetStorage([]byte(unJailedFunds), currentValue.Bytes())
}

func (v *validatorSC) unJailV1(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) == 0 {
		v.eei.AddReturnMessage("invalid number of arguments: expected min 1, got 0")
		return vmcommon.UserError
	}

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	totalUnJailPrice := big.NewInt(0).Mul(validatorConfig.UnJailPrice, big.NewInt(int64(len(args.Arguments))))

	if totalUnJailPrice.Cmp(args.CallValue) != 0 {
		v.eei.AddReturnMessage("insufficient funds sent for unJail")
		return vmcommon.UserError
	}

	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnJail * uint64(len(args.Arguments)))
	if err != nil {
		v.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	err = verifyBLSPublicKeys(registrationData, args.Arguments)
	if err != nil {
		v.eei.AddReturnMessage("could not get all blsKeys from registration data: error " + vm.ErrBLSPublicKeyMissmatch.Error())
		return vmcommon.UserError
	}

	for _, blsKey := range args.Arguments {
		vmOutput, errExec := v.executeOnStakingSC([]byte("unJail@" + hex.EncodeToString(blsKey)))
		if errExec != nil {
			v.eei.AddReturnMessage(errExec.Error())
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
			continue
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
		}
	}

	return vmcommon.Ok
}

func (v *validatorSC) unJail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagEnableStaking.IsSet() {
		return v.unJailV1(args)
	}

	if len(args.Arguments) == 0 {
		v.eei.AddReturnMessage("invalid number of arguments: expected at least 1")
		return vmcommon.UserError
	}

	numBLSKeys := len(args.Arguments)
	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	totalUnJailPrice := big.NewInt(0).Mul(validatorConfig.UnJailPrice, big.NewInt(int64(numBLSKeys)))

	if totalUnJailPrice.Cmp(args.CallValue) != 0 {
		v.eei.AddReturnMessage("wanted exact unjail price * numNodes")
		return vmcommon.UserError
	}

	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnJail * uint64(numBLSKeys))
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	err = verifyBLSPublicKeys(registrationData, args.Arguments)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return vmcommon.UserError
	}

	transferBack := big.NewInt(0)
	for _, blsKey := range args.Arguments {
		vmOutput, errExec := v.executeOnStakingSC([]byte("unJail@" + hex.EncodeToString(blsKey)))
		if errExec != nil || vmOutput.ReturnCode != vmcommon.Ok {
			transferBack.Add(transferBack, validatorConfig.UnJailPrice)
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
			continue
		}
	}

	if transferBack.Cmp(zero) > 0 {
		err = v.eei.Transfer(args.CallerAddr, args.RecipientAddr, transferBack, nil, 0)
		if err != nil {
			v.eei.AddReturnMessage("transfer error on unJail function")
			return vmcommon.UserError
		}
	}

	finalUnJailFunds := big.NewInt(0).Sub(args.CallValue, transferBack)
	v.addToUnJailFunds(finalUnJailFunds)

	return vmcommon.Ok
}

func (v *validatorSC) changeRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		v.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}
	if len(args.Arguments[0]) != v.walletAddressLen {
		v.eei.AddReturnMessage("wrong reward address")
		return vmcommon.UserError
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		v.eei.AddReturnMessage("cannot change reward address, key is not registered")
		return vmcommon.UserError
	}
	if bytes.Equal(registrationData.RewardAddress, args.Arguments[0]) {
		v.eei.AddReturnMessage("new reward address is equal with the old reward address")
		return vmcommon.UserError
	}

	err = v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.ChangeRewardAddress * uint64(len(registrationData.BlsPubKeys)))
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData.RewardAddress = args.Arguments[0]
	err = v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	txData := "changeRewardAddress@" + hex.EncodeToString(registrationData.RewardAddress)
	for _, blsKey := range registrationData.BlsPubKeys {
		txData += "@" + hex.EncodeToString(blsKey)
	}

	vmOutput, err := v.executeOnStakingSC([]byte(txData))
	if err != nil {
		v.eei.AddReturnMessage("cannot change reward address: error " + err.Error())
		return vmcommon.UserError
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	return vmcommon.Ok
}

func (v *validatorSC) get(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		v.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected exactly %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	value := v.eei.GetStorage(args.Arguments[0])
	v.eei.Finish(value)

	return vmcommon.Ok
}

func (v *validatorSC) verifyConfig(validatorConfig *ValidatorConfig) vmcommon.ReturnCode {
	if validatorConfig.MinStakeValue.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, validatorConfig.MinStakeValue).Error()
		v.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if validatorConfig.TotalSupply.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidGenesisTotalSupply, validatorConfig.TotalSupply).Error()
		v.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if validatorConfig.MinStep.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStepValue, validatorConfig.MinStep).Error()
		v.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if validatorConfig.NodePrice.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, validatorConfig.NodePrice).Error()
		v.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if validatorConfig.UnJailPrice.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, validatorConfig.UnJailPrice).Error()
		v.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	return vmcommon.Ok
}

func (v *validatorSC) checkConfigCorrectness(config ValidatorConfig) error {
	if config.MinStakeValue == nil {
		return fmt.Errorf("%w for MinStakeValue", vm.ErrIncorrectConfig)
	}
	if config.NodePrice == nil {
		return fmt.Errorf("%w for NodePrice", vm.ErrIncorrectConfig)
	}
	if config.TotalSupply == nil {
		return fmt.Errorf("%w for GenesisTotalSupply", vm.ErrIncorrectConfig)
	}
	if config.MinStep == nil {
		return fmt.Errorf("%w for MinStep", vm.ErrIncorrectConfig)
	}
	if config.UnJailPrice == nil {
		return fmt.Errorf("%w for UnJailPrice", vm.ErrIncorrectConfig)
	}
	return nil
}

func (v *validatorSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := v.eei.GetStorage([]byte(ownerKey))
	if ownerAddress != nil {
		v.eei.AddReturnMessage("smart contract was already initialized")
		return vmcommon.UserError
	}

	v.eei.SetStorage([]byte(ownerKey), args.CallerAddr)

	return vmcommon.Ok
}

func (v *validatorSC) getBLSRegisteredData(blsKey []byte) (*vmcommon.VMOutput, error) {
	return v.executeOnStakingSC([]byte("get@" + hex.EncodeToString(blsKey)))
}

func (v *validatorSC) getNewValidKeys(registeredKeys [][]byte, keysFromArgument [][]byte) ([][]byte, error) {
	registeredKeysMap := make(map[string]struct{})

	for _, blsKey := range registeredKeys {
		registeredKeysMap[string(blsKey)] = struct{}{}
	}

	newKeys := make([][]byte, 0)
	for i := uint64(0); i < uint64(len(keysFromArgument)); i++ {
		_, exists := registeredKeysMap[string(keysFromArgument[i])]
		if exists {
			continue
		}

		newKeys = append(newKeys, keysFromArgument[i])
	}

	for _, newKey := range newKeys {
		vmOutput, err := v.getBLSRegisteredData(newKey)
		if err != nil ||
			(len(vmOutput.ReturnData) > 0 && len(vmOutput.ReturnData[0]) > 0) {
			return nil, vm.ErrKeyAlreadyRegistered
		}
	}

	return newKeys, nil
}

func (v *validatorSC) registerBLSKeys(
	registrationData *ValidatorDataV2,
	pubKey []byte,
	ownerAddress []byte,
	args [][]byte,
) ([][]byte, error) {
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()
	if uint64(len(args)) < maxNodesToRun+1 {
		v.eei.AddReturnMessage(fmt.Sprintf("not enough arguments to process stake function: expected min %d, got %d", maxNodesToRun+1, len(args)))
		return nil, vm.ErrNotEnoughArgumentsToStake
	}

	blsKeys := v.getVerifiedBLSKeysFromArgs(pubKey, args)
	newKeys, err := v.getNewValidKeys(registrationData.BlsPubKeys, blsKeys)
	if err != nil {
		return nil, err
	}

	for _, blsKey := range newKeys {
		vmOutput, errExec := v.executeOnStakingSC([]byte("register@" +
			hex.EncodeToString(blsKey) + "@" +
			hex.EncodeToString(registrationData.RewardAddress) + "@" +
			hex.EncodeToString(ownerAddress) + "@",
		))
		if errExec != nil {
			v.eei.AddReturnMessage("cannot do register: " + errExec.Error())
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
			return nil, err
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			v.eei.AddReturnMessage("cannot do register: " + vmOutput.ReturnCode.String())
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
			return nil, vm.ErrKeyAlreadyRegistered
		}

		registrationData.BlsPubKeys = append(registrationData.BlsPubKeys, blsKey)
	}

	return blsKeys, nil
}

func (v *validatorSC) updateStakeValue(registrationData *ValidatorDataV2, caller []byte) vmcommon.ReturnCode {
	if len(registrationData.BlsPubKeys) == 0 && !core.IsSmartContractAddress(caller) {
		v.eei.AddReturnMessage("no bls keys has been provided")
		return vmcommon.UserError
	}

	if len(registrationData.RewardAddress) == 0 {
		registrationData.RewardAddress = caller
	}

	err := v.saveRegistrationData(caller, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) getVerifiedBLSKeysFromArgs(txPubKey []byte, args [][]byte) [][]byte {
	blsKeys := make([][]byte, 0)
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()

	invalidBlsKeys := make([]string, 0)
	for i := uint64(1); i < maxNodesToRun*2+1; i += 2 {
		blsKey := args[i]
		signedMessage := args[i+1]
		err := v.sigVerifier.Verify(txPubKey, signedMessage, blsKey)
		if err != nil {
			invalidBlsKeys = append(invalidBlsKeys, hex.EncodeToString(blsKey))
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{invalidKey})
			continue
		}

		blsKeys = append(blsKeys, blsKey)
	}
	if len(invalidBlsKeys) != 0 {
		returnMessage := "invalid BLS keys: " + strings.Join(invalidBlsKeys, ", ")
		v.eei.AddReturnMessage(returnMessage)
	}

	return blsKeys
}

func checkDoubleBLSKeys(blsKeys [][]byte) bool {
	mapKeys := make(map[string]struct{})
	for _, blsKey := range blsKeys {
		_, found := mapKeys[string(blsKey)]
		if found {
			return true
		}

		mapKeys[string(blsKey)] = struct{}{}
	}
	return false
}

func (v *validatorSC) cleanRegisteredData(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagDoubleKey.IsSet() {
		v.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}

	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage("must be called with 0 value")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		v.eei.AddReturnMessage("must be called with 0 arguments")
		return vmcommon.UserError
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	if len(registrationData.BlsPubKeys) <= 1 {
		return vmcommon.Ok
	}

	changesMade := false
	newList := make([][]byte, 0)
	mapExistingKeys := make(map[string]struct{})
	for _, blsKey := range registrationData.BlsPubKeys {
		_, found := mapExistingKeys[string(blsKey)]
		if found {
			changesMade = true
			continue
		}

		mapExistingKeys[string(blsKey)] = struct{}{}
		newList = append(newList, blsKey)
	}

	if !changesMade {
		return vmcommon.Ok
	}

	registrationData.BlsPubKeys = make([][]byte, 0, len(newList))
	registrationData.BlsPubKeys = newList

	err = v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) reStakeUnStakedNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagEnableTopUp.IsSet() {
		v.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}

	if len(args.Arguments) == 0 {
		v.eei.AddReturnMessage("need arguments of which node to unStake")
		return vmcommon.UserError
	}

	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Stake * uint64(len(args.Arguments)))
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	numQualified := big.NewInt(0).Div(registrationData.TotalStakeValue, validatorConfig.NodePrice)
	if uint64(len(args.Arguments)) > numQualified.Uint64() {
		v.eei.AddReturnMessage("insufficient funds")
		return vmcommon.OutOfFunds
	}

	mapNodesToReStake, err := v.checkAllGivenKeysAreUnStaked(registrationData, args.Arguments)
	if err != nil {
		v.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	numStakedAndWaiting, _, err := v.getNumStakedAndWaitingNodes(registrationData, mapNodesToReStake, true)
	if err != nil {
		v.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if numStakedAndWaiting+uint64(len(args.Arguments)) > numQualified.Uint64() {
		v.eei.AddReturnMessage("insufficient funds to reactivate given nodes")
		return vmcommon.UserError
	}

	for _, blsKey := range args.Arguments {
		_ = v.stakeOneNode(blsKey, registrationData.RewardAddress, args.CallerAddr)
	}

	return vmcommon.Ok
}

func (v *validatorSC) getNumStakedAndWaitingNodes(
	registrationData *ValidatorDataV2,
	mapCheckedKeys map[string]struct{},
	checkJailed bool,
) (uint64, [][]byte, error) {
	errUseGas := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.GetAllNodeStates)
	if errUseGas != nil {
		return 0, nil, errUseGas
	}

	listActiveNodes := make([][]byte, 0)
	numActiveNodes := uint64(0)
	for _, blsKey := range registrationData.BlsPubKeys {
		_, exists := mapCheckedKeys[string(blsKey)]
		if exists {
			continue
		}

		stakedData, err := v.getStakedData(blsKey)
		if err != nil {
			return 0, nil, err
		}

		if checkJailed && stakedData.Jailed {
			return 0, nil, vm.ErrBLSPublicKeyAlreadyJailed
		}

		if stakedData.Staked || stakedData.Waiting || stakedData.Jailed {
			numActiveNodes++
		}

		if stakedData.Staked || stakedData.Waiting {
			listActiveNodes = append(listActiveNodes, blsKey)
		}
	}

	return numActiveNodes, listActiveNodes, nil
}

func (v *validatorSC) checkAllGivenKeysAreUnStaked(registrationData *ValidatorDataV2, blsKeys [][]byte) (map[string]struct{}, error) {
	if uint32(len(blsKeys)) > registrationData.NumRegistered {
		return nil, fmt.Errorf("%w arguments must be unStaked blsKeys", vm.ErrBLSPublicKeyMismatch)
	}

	registeredKeysMap := make(map[string]struct{})
	for _, blsKey := range registrationData.BlsPubKeys {
		registeredKeysMap[string(blsKey)] = struct{}{}
	}

	for i := uint64(0); i < uint64(len(blsKeys)); i++ {
		_, exists := registeredKeysMap[string(blsKeys[i])]
		if !exists {
			return nil, fmt.Errorf("%w argument is not registered", vm.ErrBLSPublicKeyMismatch)
		}
	}

	mapBlsKeys := make(map[string]struct{})
	for _, blsKey := range blsKeys {
		stakedData, err := v.getStakedData(blsKey)
		if err != nil {
			return nil, err
		}
		if stakedData.Jailed || stakedData.UnStakedNonce == 0 {
			return nil, fmt.Errorf("%w arguments is not unStaked blsKeys", vm.ErrBLSPublicKeyMismatch)
		}

		mapBlsKeys[string(blsKey)] = struct{}{}
	}

	return mapBlsKeys, nil
}

func (v *validatorSC) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	isGenesis := v.eei.BlockChainHook().CurrentNonce() == 0
	stakeEnabled := isGenesis || v.flagEnableStaking.IsSet()
	if !stakeEnabled {
		v.eei.AddReturnMessage(vm.StakeNotEnabled)
		return vmcommon.UserError
	}

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	registrationData.TotalStakeValue.Add(registrationData.TotalStakeValue, args.CallValue)
	if registrationData.TotalStakeValue.Cmp(validatorConfig.NodePrice) < 0 &&
		!core.IsSmartContractAddress(args.CallerAddr) {
		v.eei.AddReturnMessage(
			fmt.Sprintf("insufficient stake value: expected %s, got %s",
				validatorConfig.NodePrice.String(),
				registrationData.TotalStakeValue.String(),
			),
		)
		return vmcommon.UserError
	}

	lenArgs := len(args.Arguments)
	if lenArgs == 0 {
		return v.updateStakeValue(registrationData, args.CallerAddr)
	}

	if !isNumArgsCorrectToStake(args.Arguments) {
		v.eei.AddReturnMessage("invalid number of arguments to call stake")
		return vmcommon.UserError
	}

	maxNodesToRun := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	if maxNodesToRun == 0 {
		v.eei.AddReturnMessage("number of nodes argument must be greater than zero")
		return vmcommon.UserError
	}

	err = v.eei.UseGas((maxNodesToRun - 1) * v.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	isAlreadyRegistered := len(registrationData.RewardAddress) > 0
	if !isAlreadyRegistered {
		registrationData.RewardAddress = args.CallerAddr
	}

	registrationData.MaxStakePerNode = big.NewInt(0).Set(registrationData.TotalStakeValue)
	registrationData.Epoch = v.eei.BlockChainHook().CurrentEpoch()

	blsKeys, err := v.registerBLSKeys(registrationData, args.CallerAddr, args.CallerAddr, args.Arguments)
	if err != nil {
		v.eei.AddReturnMessage("cannot register bls key: error " + err.Error())
		return vmcommon.UserError
	}
	if v.flagDoubleKey.IsSet() && checkDoubleBLSKeys(blsKeys) {
		v.eei.AddReturnMessage("invalid arguments, found same bls key twice")
		return vmcommon.UserError
	}

	numQualified := big.NewInt(0).Div(registrationData.TotalStakeValue, validatorConfig.NodePrice)
	if uint64(len(registrationData.BlsPubKeys)) > numQualified.Uint64() {
		v.eei.AddReturnMessage("insufficient funds")
		return vmcommon.OutOfFunds
	}

	// do the optionals - rewardAddress and maxStakePerNode
	if uint64(lenArgs) > maxNodesToRun*2+1 {
		for i := maxNodesToRun*2 + 1; i < uint64(lenArgs); i++ {
			if len(args.Arguments[i]) == v.walletAddressLen {
				if !isAlreadyRegistered {
					registrationData.RewardAddress = args.Arguments[i]
				} else {
					v.eei.AddReturnMessage("reward address after being registered can be changed only through changeRewardAddress")
				}
				continue
			}

			maxStakePerNode := big.NewInt(0).SetBytes(args.Arguments[i])
			registrationData.MaxStakePerNode.Set(maxStakePerNode)
		}
	}

	v.activateStakingFor(
		blsKeys,
		numQualified.Uint64(),
		registrationData,
		validatorConfig.NodePrice,
		registrationData.RewardAddress,
		args.CallerAddr,
	)

	err = v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) activateStakingFor(
	blsKeys [][]byte,
	numQualified uint64,
	registrationData *ValidatorDataV2,
	fixedStakeValue *big.Int,
	rewardAddress []byte,
	ownerAddress []byte,
) {
	numRegistered := uint64(registrationData.NumRegistered)

	if v.flagEnableTopUp.IsSet() {
		v.reActivateStakingForNodes(blsKeys, numQualified, registrationData, rewardAddress, ownerAddress)
	}

	for i := uint64(0); numRegistered < numQualified && i < uint64(len(blsKeys)); i++ {
		currentBLSKey := blsKeys[i]
		stakedData, err := v.getStakedData(currentBLSKey)
		if err != nil {
			continue
		}

		if stakedData.Staked || stakedData.Waiting {
			continue
		}

		skipAdd := v.stakeOneNode(currentBLSKey, rewardAddress, ownerAddress)
		if skipAdd {
			continue
		}

		if stakedData.UnStakedNonce == 0 {
			numRegistered++
		}
	}

	registrationData.NumRegistered = uint32(numRegistered)
	registrationData.LockedStake.Mul(fixedStakeValue, big.NewInt(0).SetUint64(numRegistered))
}

func (v *validatorSC) reActivateStakingForNodes(
	blsKeys [][]byte,
	numQualified uint64,
	registrationData *ValidatorDataV2,
	rewardAddress []byte,
	ownerAddress []byte,
) {
	numRegistered := uint64(registrationData.NumRegistered)
	if numRegistered < numQualified {
		return
	}

	for _, blsKey := range blsKeys {
		stakedData, err := v.getStakedData(blsKey)
		if err != nil {
			continue
		}
		if stakedData.Jailed || stakedData.UnStakedNonce == 0 {
			continue
		}

		_ = v.stakeOneNode(blsKey, rewardAddress, ownerAddress)
	}
}

func (v *validatorSC) stakeOneNode(
	blsKey []byte,
	rewardAddress []byte,
	ownerAddress []byte,
) bool {
	vmOutput, err := v.executeOnStakingSC([]byte("stake@" +
		hex.EncodeToString(blsKey) + "@" +
		hex.EncodeToString(rewardAddress) + "@" +
		hex.EncodeToString(ownerAddress),
	))
	if err != nil {
		v.eei.AddReturnMessage(fmt.Sprintf("cannot do stake for key %s, error %s", hex.EncodeToString(blsKey), err.Error()))
		v.eei.Finish(blsKey)
		v.eei.Finish([]byte{failed})
		return true
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		v.eei.AddReturnMessage(fmt.Sprintf("cannot do stake for key %s, error %s", hex.EncodeToString(blsKey), vmOutput.ReturnCode.String()))
		v.eei.Finish(blsKey)
		v.eei.Finish([]byte{failed})
		return true
	}

	if len(vmOutput.ReturnData) > 0 && bytes.Equal(vmOutput.ReturnData[0], []byte{waiting}) {
		v.eei.Finish(blsKey)
		v.eei.Finish([]byte{waiting})
	}

	return false
}

func (v *validatorSC) executeOnStakingSC(data []byte) (*vmcommon.VMOutput, error) {
	return v.eei.ExecuteOnDestContext(v.stakingSCAddress, v.validatorSCAddress, big.NewInt(0), data)
}

func (v *validatorSC) setOwnerOfBlsKey(blsKey []byte, ownerAddress []byte) bool {
	vmOutput, err := v.executeOnStakingSC([]byte("setOwner@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(ownerAddress)))
	if err != nil {
		v.eei.AddReturnMessage(fmt.Sprintf("cannot set owner for key %s, error %s", hex.EncodeToString(blsKey), err.Error()))
		v.eei.Finish(blsKey)
		v.eei.Finish([]byte{failed})
		return false

	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		v.eei.AddReturnMessage(fmt.Sprintf("cannot set owner for key %s, error %s", hex.EncodeToString(blsKey), vmOutput.ReturnCode.String()))
		v.eei.Finish(blsKey)
		v.eei.Finish([]byte{failed})
		return false
	}

	return true
}

func (v *validatorSC) basicChecksForUnStakeNodes(args *vmcommon.ContractCallInput) (*ValidatorDataV2, vmcommon.ReturnCode) {
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return nil, vmcommon.UserError
	}
	if len(args.Arguments) == 0 {
		v.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return nil, vmcommon.UserError
	}
	if !v.flagEnableStaking.IsSet() {
		v.eei.AddReturnMessage(vm.UnStakeNotEnabled)
		return nil, vmcommon.UserError
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return nil, vmcommon.UserError
	}

	err = v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnStake * uint64(len(args.Arguments)))
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return nil, vmcommon.OutOfGas
	}

	blsKeys := args.Arguments
	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return nil, vmcommon.UserError
	}

	return registrationData, vmcommon.Ok
}

func (v *validatorSC) unStakeNodesFromStakingSC(blsKeys [][]byte, registrationData *ValidatorDataV2) uint64 {
	numSuccess := uint64(0)
	for _, blsKey := range blsKeys {
		vmOutput, errExec := v.executeOnStakingSC([]byte("unStake@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(registrationData.RewardAddress)))
		if errExec != nil {
			v.eei.AddReturnMessage(fmt.Sprintf("cannot do unStake for key %s: %s", hex.EncodeToString(blsKey), errExec.Error()))
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
			continue
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			v.eei.AddReturnMessage(fmt.Sprintf("cannot do unStake for key %s: %s", hex.EncodeToString(blsKey), vmOutput.ReturnCode.String()))
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
			continue
		}

		numSuccess++
	}

	return numSuccess
}

// This is the complete unStake - which after enabling economics V2 will create unStakedFunds on the registration data
func (v *validatorSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}

	registrationData, returnCode := v.basicChecksForUnStakeNodes(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	numSuccess := v.unStakeNodesFromStakingSC(args.Arguments, registrationData)
	if !v.flagEnableTopUp.IsSet() {
		return vmcommon.Ok
	}

	// continue by unstaking tokens as well
	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	unStakeFromNodes := big.NewInt(0).Mul(validatorConfig.NodePrice, big.NewInt(0).SetUint64(numSuccess))
	if unStakeFromNodes.Cmp(registrationData.TotalStakeValue) > 0 {
		unStakeFromNodes.Set(registrationData.TotalStakeValue)
	}

	returnCode = v.processUnStakeValue(registrationData, unStakeFromNodes)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err := v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) unStakeNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagEnableTopUp.IsSet() {
		v.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}

	registrationData, returnCode := v.basicChecksForUnStakeNodes(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	_ = v.unStakeNodesFromStakingSC(args.Arguments, registrationData)

	return vmcommon.Ok
}

func (v *validatorSC) unBondNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagEnableTopUp.IsSet() {
		v.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}

	registrationData, returnCode := v.checkUnBondArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	unBondedKeys := v.unBondNodesFromStakingSC(args.Arguments)
	returnCode = v.updateRegistrationDataAfterUnBond(registrationData, unBondedKeys, args.CallerAddr)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (v *validatorSC) checkUnBondArguments(args *vmcommon.ContractCallInput) (*ValidatorDataV2, vmcommon.ReturnCode) {
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return nil, vmcommon.UserError
	}
	if len(args.Arguments) == 0 {
		v.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return nil, vmcommon.UserError
	}
	if !v.flagEnableStaking.IsSet() {
		v.eei.AddReturnMessage(vm.UnBondNotEnabled)
		return nil, vmcommon.UserError
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return nil, vmcommon.UserError
	}

	err = verifyBLSPublicKeys(registrationData, args.Arguments)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return nil, vmcommon.UserError
	}

	err = v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnBond * uint64(len(args.Arguments)))
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return nil, vmcommon.OutOfGas
	}

	return registrationData, vmcommon.Ok
}

func (v *validatorSC) unBondNodesFromStakingSC(blsKeys [][]byte) [][]byte {
	unBondedKeys := make([][]byte, 0)
	for _, blsKey := range blsKeys {
		vmOutput, errExec := v.executeOnStakingSC([]byte("unBond@" + hex.EncodeToString(blsKey)))
		if errExec != nil || vmOutput.ReturnCode != vmcommon.Ok {
			v.eei.AddReturnMessage(fmt.Sprintf("cannot do unBond for key: %s", hex.EncodeToString(blsKey)))
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
			continue
		}

		unBondedKeys = append(unBondedKeys, blsKey)
	}

	return unBondedKeys
}

func (v *validatorSC) unBondV1(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, returnCode := v.checkUnBondArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	unBondedKeys := v.unBondNodesFromStakingSC(args.Arguments)
	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	totalUnBond := big.NewInt(0).Mul(validatorConfig.NodePrice, big.NewInt(int64(len(unBondedKeys))))

	if registrationData.LockedStake.Cmp(totalUnBond) < 0 {
		v.eei.AddReturnMessage("contract error on unBond function, lockedStake < totalUnBond")
		return vmcommon.UserError
	}

	if registrationData.NumRegistered < uint32(len(unBondedKeys)) {
		v.eei.AddReturnMessage("contract error on unBond function, missing nodes")
		return vmcommon.UserError
	}

	registrationData.NumRegistered -= uint32(len(unBondedKeys))
	registrationData.LockedStake.Sub(registrationData.LockedStake, totalUnBond)
	registrationData.TotalStakeValue.Sub(registrationData.TotalStakeValue, totalUnBond)
	if registrationData.TotalStakeValue.Cmp(zero) < 0 {
		v.eei.AddReturnMessage("contract error on unBond function, total stake < 0")
		return vmcommon.UserError
	}

	if registrationData.LockedStake.Cmp(zero) == 0 && registrationData.TotalStakeValue.Cmp(zero) == 0 {
		v.eei.SetStorage(args.CallerAddr, nil)
	} else {
		v.deleteUnBondedKeys(registrationData, unBondedKeys)
		errSave := v.saveRegistrationData(args.CallerAddr, registrationData)
		if errSave != nil {
			v.eei.AddReturnMessage("cannot save registration data: error " + errSave.Error())
			return vmcommon.UserError
		}
	}

	err := v.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBond, nil, 0)
	if err != nil {
		v.eei.AddReturnMessage("transfer error on unBond function")
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagEnableTopUp.IsSet() {
		return v.unBondV1(args)
	}

	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}
	registrationData, returnCode := v.checkUnBondArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	unBondedKeys := v.unBondNodesFromStakingSC(args.Arguments)

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	totalUnBond := big.NewInt(0).Mul(validatorConfig.NodePrice, big.NewInt(int64(len(unBondedKeys))))
	if len(unBondedKeys) > 0 {
		totalUnBond, returnCode = v.unBondTokensFromRegistrationData(registrationData, totalUnBond)
		if returnCode != vmcommon.Ok {
			return returnCode
		}
	}

	returnCode = v.updateRegistrationDataAfterUnBond(registrationData, unBondedKeys, args.CallerAddr)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err := v.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBond, nil, 0)
	if err != nil {
		v.eei.AddReturnMessage("transfer error on unBond function")
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) updateRegistrationDataAfterUnBond(
	registrationData *ValidatorDataV2,
	unBondedKeys [][]byte,
	callerAddr []byte,
) vmcommon.ReturnCode {
	if registrationData.NumRegistered < uint32(len(unBondedKeys)) {
		v.eei.AddReturnMessage("contract error on unBond function, missing nodes")
		return vmcommon.UserError
	}

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	registrationData.NumRegistered -= uint32(len(unBondedKeys))
	registrationData.LockedStake.Mul(validatorConfig.NodePrice, big.NewInt(0).SetUint64(uint64(registrationData.NumRegistered)))
	v.deleteUnBondedKeys(registrationData, unBondedKeys)

	shouldDeleteRegistrationData := registrationData.TotalStakeValue.Cmp(zero) == 0 && registrationData.LockedStake.Cmp(zero) == 0 &&
		len(registrationData.BlsPubKeys) == 0 && len(registrationData.UnstakedInfo) == 0
	if shouldDeleteRegistrationData {
		v.eei.SetStorage(callerAddr, nil)
	} else {
		errSave := v.saveRegistrationData(callerAddr, registrationData)
		if errSave != nil {
			v.eei.AddReturnMessage("cannot save registration data: error " + errSave.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (v *validatorSC) deleteUnBondedKeys(registrationData *ValidatorDataV2, unBondedKeys [][]byte) {
	for _, unBonded := range unBondedKeys {
		for i, registeredKey := range registrationData.BlsPubKeys {
			if bytes.Equal(unBonded, registeredKey) {
				lastIndex := len(registrationData.BlsPubKeys) - 1
				registrationData.BlsPubKeys[i] = registrationData.BlsPubKeys[lastIndex]
				registrationData.BlsPubKeys = registrationData.BlsPubKeys[:lastIndex]
				break
			}
		}
	}
}

func (v *validatorSC) claim(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if v.flagEnableTopUp.IsSet() {
		//claim function will become unavailable after enabling staking v2
		v.eei.AddReturnMessage("claim function is disabled")
		return vmcommon.UserError
	}

	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage("cannot get registration data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		v.eei.AddReturnMessage("key is not registered, claim is not possible")
		return vmcommon.UserError
	}
	err = v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Claim)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	claimable := big.NewInt(0).Sub(registrationData.TotalStakeValue, registrationData.LockedStake)
	if claimable.Cmp(zero) <= 0 {
		return vmcommon.Ok
	}

	registrationData.TotalStakeValue.Set(registrationData.LockedStake)
	err = v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	err = v.eei.Transfer(args.CallerAddr, args.RecipientAddr, claimable, nil, 0)
	if err != nil {
		v.eei.AddReturnMessage("transfer error on finalizeUnStake function: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) unStakeTokens(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, returnCode := v.basicCheckForUnStakeUnBond(args, args.CallerAddr)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}

	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnStakeTokens)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		v.eei.AddReturnMessage("should have specified one argument containing the unstake value")
		return vmcommon.UserError
	}

	unStakeValue := big.NewInt(0).SetBytes(args.Arguments[0])
	returnCode = v.processUnStakeValue(registrationData, unStakeValue)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if registrationData.NumRegistered > 0 && registrationData.TotalStakeValue.Cmp(v.minDeposit) < 0 {
		v.eei.AddReturnMessage("cannot unStake tokens, the validator would remain without min deposit, nodes are still active")
		return vmcommon.UserError
	}

	err = v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) processUnStakeValue(registrationData *ValidatorDataV2, unStakeValue *big.Int) vmcommon.ReturnCode {
	unstakeValueIsOk := unStakeValue.Cmp(v.minUnstakeTokensValue) >= 0 || unStakeValue.Cmp(registrationData.TotalStakeValue) == 0
	if !unstakeValueIsOk {
		v.eei.AddReturnMessage("can not unstake the provided value either because is under the minimum threshold or " +
			"is not the value left to be unStaked")
		return vmcommon.UserError
	}
	if unStakeValue.Cmp(registrationData.TotalStakeValue) > 0 {
		v.eei.AddReturnMessage("can not unstake a bigger value than the possible allowed value which is " + registrationData.TotalStakeValue.String())
		return vmcommon.UserError
	}

	registrationData.TotalStakeValue.Sub(registrationData.TotalStakeValue, unStakeValue)
	registrationData.TotalUnstaked.Add(registrationData.TotalUnstaked, unStakeValue)
	registrationData.UnstakedInfo = append(
		registrationData.UnstakedInfo,
		&UnstakedValue{
			UnstakedNonce: v.eei.BlockChainHook().CurrentNonce(),
			UnstakedValue: unStakeValue,
		},
	)
	return vmcommon.Ok
}

func (v *validatorSC) basicCheckForUnStakeUnBond(args *vmcommon.ContractCallInput, address []byte) (*ValidatorDataV2, vmcommon.ReturnCode) {
	if !v.flagEnableTopUp.IsSet() {
		v.eei.AddReturnMessage("invalid method to call")
		return nil, vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return nil, vmcommon.UserError
	}
	registrationData, err := v.getOrCreateRegistrationData(address)
	if err != nil {
		v.eei.AddReturnMessage("cannot get registration data: error " + err.Error())
		return nil, vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		v.eei.AddReturnMessage("key is not registered, validator operation is not possible")
		return nil, vmcommon.UserError
	}
	if registrationData.TotalUnstaked == nil {
		registrationData.TotalUnstaked = big.NewInt(0)
	}

	return registrationData, vmcommon.Ok
}

func (v *validatorSC) getUnStakedTokensList(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 1 {
		v.eei.AddReturnMessage("number of arguments must be equal to 1")
		return vmcommon.UserError
	}

	registrationData, returnCode := v.basicCheckForUnStakeUnBond(args, args.Arguments[0])
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	currentNonce := v.eei.BlockChainHook().CurrentNonce()
	for _, unStakedValue := range registrationData.UnstakedInfo {
		v.eei.Finish(unStakedValue.UnstakedValue.Bytes())
		elapsedNonce := currentNonce - unStakedValue.UnstakedNonce
		if elapsedNonce >= v.unBondPeriod {
			v.eei.Finish(zero.Bytes())
			continue
		}

		remainingNonce := v.unBondPeriod - elapsedNonce
		v.eei.Finish(big.NewInt(0).SetUint64(remainingNonce).Bytes())
	}
	return vmcommon.Ok
}

func (v *validatorSC) unBondTokens(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, returnCode := v.basicCheckForUnStakeUnBond(args, args.CallerAddr)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}
	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnBondTokens)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	valueToUnBond := big.NewInt(0)
	if len(args.Arguments) > 1 {
		v.eei.AddReturnMessage("too many arguments")
		return vmcommon.UserError
	}
	if len(args.Arguments) == 1 {
		valueToUnBond = big.NewInt(0).SetBytes(args.Arguments[0])
		if valueToUnBond.Cmp(zero) <= 0 {
			v.eei.AddReturnMessage("cannot unBond negative value or zero value")
			return vmcommon.UserError
		}
	}

	totalUnBond, returnCode := v.unBondTokensFromRegistrationData(registrationData, valueToUnBond)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if totalUnBond.Cmp(zero) == 0 {
		v.eei.AddReturnMessage("no tokens that can be unbond at this time")
		return vmcommon.Ok
	}

	if registrationData.NumRegistered > 0 && registrationData.TotalStakeValue.Cmp(v.minDeposit) < 0 {
		v.eei.AddReturnMessage("cannot unStake tokens, the validator would remain without min deposit, nodes are still active")
		return vmcommon.UserError
	}

	err = v.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBond, nil, 0)
	if err != nil {
		v.eei.AddReturnMessage("transfer error on unBond function")
		return vmcommon.UserError
	}

	err = v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSC) unBondTokensFromRegistrationData(
	registrationData *ValidatorDataV2,
	valueToUnBond *big.Int,
) (*big.Int, vmcommon.ReturnCode) {
	var unstakedValue *UnstakedValue
	currentNonce := v.eei.BlockChainHook().CurrentNonce()
	totalUnBond := big.NewInt(0)
	index := 0

	stopAtUnBondValue := valueToUnBond.Cmp(zero) > 0

	splitUnStakedInfo := &UnstakedValue{UnstakedValue: big.NewInt(0)}
	for _, unstakedValue = range registrationData.UnstakedInfo {
		canUnbond := currentNonce-unstakedValue.UnstakedNonce >= v.unBondPeriod
		if !canUnbond {
			break
		}

		totalUnBond.Add(totalUnBond, unstakedValue.UnstakedValue)
		index++
		if stopAtUnBondValue && totalUnBond.Cmp(valueToUnBond) >= 0 {
			splitUnStakedInfo.UnstakedValue.Sub(totalUnBond, valueToUnBond)
			splitUnStakedInfo.UnstakedNonce = unstakedValue.UnstakedNonce
			totalUnBond.Set(valueToUnBond)
			break
		}
	}

	if splitUnStakedInfo.UnstakedValue.Cmp(zero) > 0 {
		index--
		registrationData.UnstakedInfo[index] = splitUnStakedInfo
	}

	registrationData.UnstakedInfo = registrationData.UnstakedInfo[index:]
	registrationData.TotalUnstaked.Sub(registrationData.TotalUnstaked, totalUnBond)
	if registrationData.TotalUnstaked.Cmp(zero) < 0 {
		v.eei.AddReturnMessage("too much requested to unBond")
		return nil, vmcommon.UserError
	}

	return totalUnBond, vmcommon.Ok
}

func (v *validatorSC) getTotalStaked(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	if len(registrationData.RewardAddress) == 0 {
		v.eei.AddReturnMessage("caller not registered in staking/validator sc")
		return vmcommon.UserError
	}

	v.eei.Finish([]byte(registrationData.TotalStakeValue.String()))
	return vmcommon.Ok
}

func (v *validatorSC) getTotalStakedTopUpStakedBlsKeys(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !v.flagEnableTopUp.IsSet() {
		v.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	if len(registrationData.RewardAddress) == 0 {
		v.eei.AddReturnMessage("caller not registered in staking/validator sc")
		return vmcommon.UserError
	}

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())

	numActive, listActiveNodes, err := v.getNumStakedAndWaitingNodes(registrationData, make(map[string]struct{}), false)
	if err != nil {
		v.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	stakeForNodes := big.NewInt(0).Mul(validatorConfig.NodePrice, big.NewInt(0).SetUint64(numActive))

	topUp := big.NewInt(0).Set(registrationData.TotalStakeValue)
	topUp.Sub(topUp, stakeForNodes)
	if topUp.Cmp(zero) < 0 {
		log.Warn("topup value is less than 0")
		topUp.Set(zero)
	}

	if registrationData.TotalStakeValue.Cmp(zero) < 0 {
		v.eei.AddReturnMessage("contract error on getTopUp function, total stake < locked stake value")
		return vmcommon.UserError
	}

	v.eei.Finish(topUp.Bytes())
	v.eei.Finish(registrationData.TotalStakeValue.Bytes())
	v.eei.Finish(big.NewInt(0).SetUint64(numActive).Bytes())

	for _, blsKey := range listActiveNodes {
		v.eei.Finish(blsKey)
	}

	return vmcommon.Ok
}

func (v *validatorSC) slash(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	// TODO: implement this. It is needed as last component of slashing. Slashing should happen to the funds of the
	// validator which is running the nodes
	return vmcommon.Ok
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (v *validatorSC) EpochConfirmed(epoch uint32) {
	v.flagEnableStaking.Toggle(epoch >= v.enableStakingEpoch)
	log.Debug("validatorSC: stake/unstake/unbond", "enabled", v.flagEnableStaking.IsSet())

	v.flagEnableTopUp.Toggle(epoch >= v.stakingV2Epoch)
	log.Debug("validatorSC: top up mechanism", "enabled", v.flagEnableTopUp.IsSet())

	v.flagDoubleKey.Toggle(epoch >= v.enableDoubleKeyEpoch)
	log.Debug("stakingAuctionSC: doubleKeyProtection", "enabled", v.flagDoubleKey.IsSet())

}

// CanUseContract returns true if contract can be used
func (v *validatorSC) CanUseContract() bool {
	return true
}

func (v *validatorSC) getBlsKeysStatus(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, v.validatorSCAddress) {
		v.eei.AddReturnMessage("this is only a view function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		v.eei.AddReturnMessage("number of arguments must be equal to 1")
		return vmcommon.UserError
	}

	registrationData, err := v.getOrCreateRegistrationData(args.Arguments[0])
	if err != nil {
		v.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	if len(registrationData.BlsPubKeys) == 0 {
		v.eei.AddReturnMessage("no bls keys")
		return vmcommon.Ok
	}

	for _, blsKey := range registrationData.BlsPubKeys {
		vmOutput, errExec := v.executeOnStakingSC([]byte("getBLSKeyStatus@" + hex.EncodeToString(blsKey)))
		if errExec != nil {
			v.eei.AddReturnMessage("cannot get bls key status: bls key - " + hex.EncodeToString(blsKey) + " error - " + errExec.Error())
			continue
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			v.eei.AddReturnMessage("error in getting bls key status: bls key - " + hex.EncodeToString(blsKey))
			continue
		}

		if len(vmOutput.ReturnData) != 1 {
			v.eei.AddReturnMessage("cannot get bls key status for key " + hex.EncodeToString(blsKey))
			continue
		}

		v.eei.Finish(blsKey)
		v.eei.Finish(vmOutput.ReturnData[0])
	}

	return vmcommon.Ok
}

// SetNewGasCost is called whenever a gas cost was changed
func (v *validatorSC) SetNewGasCost(gasCost vm.GasCost) {
	v.mutExecution.Lock()
	v.gasCost = gasCost
	v.mutExecution.Unlock()
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (v *validatorSC) IsInterfaceNil() bool {
	return v == nil
}
