//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. auction.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const minArgsLenToChangeValidatorKey = 4
const unJailedFunds = "unJailFunds"

var zero = big.NewInt(0)

// return codes for each input blskey
const (
	ok uint8 = iota
	invalidKey
	failed
	waiting
)

type stakingAuctionSC struct {
	eei                   vm.SystemEI
	unBondPeriod          uint64
	sigVerifier           vm.MessageSignVerifier
	baseConfig            AuctionConfig
	stakingV2Epoch        uint32
	stakingSCAddress      []byte
	auctionSCAddress      []byte
	walletAddressLen      int
	enableStakingEpoch    uint32
	gasCost               vm.GasCost
	marshalizer           marshal.Marshalizer
	flagEnableStaking     atomic.Flag
	flagEnableTopUp       atomic.Flag
	minUnstakeTokensValue *big.Int
}

// ArgsStakingAuctionSmartContract is the arguments structure to create a new StakingAuctionSmartContract
type ArgsStakingAuctionSmartContract struct {
	StakingSCConfig    config.StakingSystemSCConfig
	GenesisTotalSupply *big.Int
	NumOfNodesToSelect uint64
	Eei                vm.SystemEI
	SigVerifier        vm.MessageSignVerifier
	StakingSCAddress   []byte
	AuctionSCAddress   []byte
	GasCost            vm.GasCost
	Marshalizer        marshal.Marshalizer
	EpochNotifier      vm.EpochNotifier
}

// NewStakingAuctionSmartContract creates an auction smart contract
func NewStakingAuctionSmartContract(
	args ArgsStakingAuctionSmartContract,
) (*stakingAuctionSC, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingSCAddress) == 0 {
		return nil, vm.ErrNilStakingSmartContractAddress
	}
	if len(args.AuctionSCAddress) == 0 {
		return nil, vm.ErrNilAuctionSmartContractAddress
	}
	if args.NumOfNodesToSelect < 1 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinNumberOfNodes, args.NumOfNodesToSelect)
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

	baseConfig := AuctionConfig{
		NumNodes:    uint32(args.NumOfNodesToSelect),
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

	reg := &stakingAuctionSC{
		eei:                   args.Eei,
		unBondPeriod:          args.StakingSCConfig.UnBondPeriod,
		sigVerifier:           args.SigVerifier,
		baseConfig:            baseConfig,
		stakingV2Epoch:        args.StakingSCConfig.StakingV2Epoch,
		enableStakingEpoch:    args.StakingSCConfig.StakeEnableEpoch,
		stakingSCAddress:      args.StakingSCAddress,
		auctionSCAddress:      args.AuctionSCAddress,
		gasCost:               args.GasCost,
		marshalizer:           args.Marshalizer,
		minUnstakeTokensValue: minUnstakeTokensValue,
		walletAddressLen:      len(args.AuctionSCAddress),
	}

	args.EpochNotifier.RegisterNotifyHandler(reg)

	return reg, nil
}

// Execute calls one of the functions from the auction staking smart contract and runs the code according to the input
func (s *stakingAuctionSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := CheckIfNil(args)
	if err != nil {
		s.eei.AddReturnMessage("nil arguments: " + err.Error())
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return s.init(args)
	case "stake":
		return s.stake(args)
	case "unStake":
		return s.unStake(args)
	case "unBond":
		return s.unBond(args)
	case "claim":
		return s.claim(args)
	case "get":
		return s.get(args)
	case "setConfig":
		return s.setConfig(args)
	case "changeRewardAddress":
		return s.changeRewardAddress(args)
	case "unJail":
		return s.unJail(args)
	case "getTotalStaked":
		return s.getTotalStaked(args)
	case "getTopUp":
		return s.getTopUp(args)
	case "getBlsKeysStatus":
		return s.getBlsKeysStatus(args)
	case "unStakeTokens":
		return s.unStakeTokens(args)
	case "unBondTokens":
		return s.unBondTokens(args)
	case "updateStakingV2":
		return s.updateStakingV2(args)
	case "unBondTokensWithNodes":
		return s.unBondTokensWithNodes(args)
	case "unStakeTokensWithNodes":
		return s.unStakeTokensWithNodes(args)
	}

	s.eei.AddReturnMessage("invalid method to call")
	return vmcommon.UserError
}

func (s *stakingAuctionSC) addToUnJailFunds(value *big.Int) {
	currentValue := big.NewInt(0)
	storageData := s.eei.GetStorage([]byte(unJailedFunds))
	if len(storageData) > 0 {
		currentValue.SetBytes(storageData)
	}

	currentValue.Add(currentValue, value)
	s.eei.SetStorage([]byte(unJailedFunds), currentValue.Bytes())
}

func (s *stakingAuctionSC) unJailV1(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) == 0 {
		s.eei.AddReturnMessage("invalid number of arguments: expected min 1, got 0")
		return vmcommon.UserError
	}

	auctionConfig := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())
	totalUnJailPrice := big.NewInt(0).Mul(auctionConfig.UnJailPrice, big.NewInt(int64(len(args.Arguments))))

	if totalUnJailPrice.Cmp(args.CallValue) != 0 {
		s.eei.AddReturnMessage("insufficient funds sent for unJail")
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnJail * uint64(len(args.Arguments)))
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	err = verifyBLSPublicKeys(registrationData, args.Arguments)
	if err != nil {
		s.eei.AddReturnMessage("could not get all blsKeys from registration data: error " + vm.ErrBLSPublicKeyMissmatch.Error())
		return vmcommon.UserError
	}

	for _, blsKey := range args.Arguments {
		vmOutput, errExec := s.executeOnStakingSC([]byte("unJail@" + hex.EncodeToString(blsKey)))
		if errExec != nil {
			s.eei.AddReturnMessage(errExec.Error())
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
		}
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) unJail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.flagEnableStaking.IsSet() {
		return s.unJailV1(args)
	}

	if len(args.Arguments) == 0 {
		s.eei.AddReturnMessage("invalid number of arguments: expected at least 1")
		return vmcommon.UserError
	}

	numBLSKeys := len(args.Arguments)
	auctionConfig := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())
	totalUnJailPrice := big.NewInt(0).Mul(auctionConfig.UnJailPrice, big.NewInt(int64(numBLSKeys)))

	if totalUnJailPrice.Cmp(args.CallValue) != 0 {
		s.eei.AddReturnMessage("wanted exact unjail price * numNodes")
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnJail * uint64(numBLSKeys))
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	err = verifyBLSPublicKeys(registrationData, args.Arguments)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return vmcommon.UserError
	}

	transferBack := big.NewInt(0)
	for _, blsKey := range args.Arguments {
		vmOutput, errExec := s.executeOnStakingSC([]byte("unJail@" + hex.EncodeToString(blsKey)))
		if errExec != nil || vmOutput.ReturnCode != vmcommon.Ok {
			transferBack.Add(transferBack, auctionConfig.UnJailPrice)
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}
	}

	if transferBack.Cmp(zero) > 0 {
		err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, transferBack, nil, 0)
		if err != nil {
			s.eei.AddReturnMessage("transfer error on unJail function")
			return vmcommon.UserError
		}
	}

	finalUnJailFunds := big.NewInt(0).Sub(args.CallValue, transferBack)
	s.addToUnJailFunds(finalUnJailFunds)

	return vmcommon.Ok
}

func (s *stakingAuctionSC) changeRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}
	if len(args.Arguments[0]) != s.walletAddressLen {
		s.eei.AddReturnMessage("wrong reward address")
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("cannot change reward address, key is not registered")
		return vmcommon.UserError
	}
	if bytes.Equal(registrationData.RewardAddress, args.Arguments[0]) {
		s.eei.AddReturnMessage("new reward address is equal with the old reward address")
		return vmcommon.UserError
	}

	err = s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.ChangeRewardAddress * uint64(len(registrationData.BlsPubKeys)))
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData.RewardAddress = args.Arguments[0]
	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	txData := "changeRewardAddress@" + hex.EncodeToString(registrationData.RewardAddress)
	for _, blsKey := range registrationData.BlsPubKeys {
		txData += "@" + hex.EncodeToString(blsKey)
	}

	vmOutput, err := s.executeOnStakingSC([]byte(txData))
	if err != nil {
		s.eei.AddReturnMessage("cannot change reward address: error " + err.Error())
		return vmcommon.UserError
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return vmOutput.ReturnCode
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) changeValidatorKeys(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	// list of arguments are NumNodes, (OldKey, NewKey, SignedMessage) X NumNodes
	if len(args.Arguments) < minArgsLenToChangeValidatorKey {
		retMessage := fmt.Sprintf("invalid number of arguments: expected min %d, got %d", minArgsLenToChangeValidatorKey, len(args.Arguments))
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}

	numNodesToChange := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	expectedNumArguments := numNodesToChange*3 + 1
	if uint64(len(args.Arguments)) < expectedNumArguments {
		retMessage := fmt.Sprintf("invalid number of arguments: expected min %d, got %d", expectedNumArguments, len(args.Arguments))
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.ChangeValidatorKeys * numNodesToChange)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.BlsPubKeys) == 0 {
		s.eei.AddReturnMessage("no bls key in storage")
		return vmcommon.UserError
	}

	for i := 1; i < len(args.Arguments); i += 3 {
		oldBlsKey := args.Arguments[i]
		newBlsKey := args.Arguments[i+1]
		signedWithNewKey := args.Arguments[i+2]

		err = s.sigVerifier.Verify(args.CallerAddr, signedWithNewKey, newBlsKey)
		if err != nil {
			s.eei.AddReturnMessage("invalid signature: error " + err.Error())
			return vmcommon.UserError
		}

		err = s.replaceBLSKey(registrationData, oldBlsKey, newBlsKey)
		if err != nil {
			s.eei.AddReturnMessage("cannot replace bls key: error " + err.Error())
			return vmcommon.UserError
		}
	}

	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) replaceBLSKey(registrationData *AuctionDataV2, oldBlsKey []byte, newBlsKey []byte) error {
	foundOldKey := false
	for i, registeredKey := range registrationData.BlsPubKeys {
		if bytes.Equal(registeredKey, oldBlsKey) {
			foundOldKey = true
			registrationData.BlsPubKeys[i] = newBlsKey
			break
		}
	}

	if !foundOldKey {
		return vm.ErrBLSPublicKeyMismatch
	}

	vmOutput, err := s.executeOnStakingSC([]byte("changeValidatorKeys@" + hex.EncodeToString(oldBlsKey) + "@" + hex.EncodeToString(newBlsKey)))
	if err != nil {
		s.eei.AddReturnMessage("cannot change validator key: error " + err.Error())
		return vm.ErrOnExecutionAtStakingSC
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return vm.ErrOnExecutionAtStakingSC
	}

	return nil
}

func (s *stakingAuctionSC) get(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected exactly %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	value := s.eei.GetStorage(args.Arguments[0])
	s.eei.Finish(value)

	return vmcommon.Ok
}

func (s *stakingAuctionSC) verifyConfig(auctionConfig *AuctionConfig) vmcommon.ReturnCode {
	if auctionConfig.MinStakeValue.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, auctionConfig.MinStakeValue).Error()
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if auctionConfig.NumNodes < 1 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidMinNumberOfNodes, auctionConfig.NumNodes).Error()
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if auctionConfig.TotalSupply.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidGenesisTotalSupply, auctionConfig.TotalSupply).Error()
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if auctionConfig.MinStep.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStepValue, auctionConfig.MinStep).Error()
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if auctionConfig.NodePrice.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, auctionConfig.NodePrice).Error()
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	if auctionConfig.UnJailPrice.Cmp(zero) <= 0 {
		retMessage := fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, auctionConfig.UnJailPrice).Error()
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}
	return vmcommon.Ok
}

func (s *stakingAuctionSC) checkConfigCorrectness(config AuctionConfig) error {
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

func (s *stakingAuctionSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := s.eei.GetStorage([]byte(ownerKey))
	if ownerAddress != nil {
		s.eei.AddReturnMessage("smart contract was already initialized")
		return vmcommon.UserError
	}

	s.eei.SetStorage([]byte(ownerKey), args.CallerAddr)

	return vmcommon.Ok
}

func (s *stakingAuctionSC) getBLSRegisteredData(blsKey []byte) (*vmcommon.VMOutput, error) {
	return s.executeOnStakingSC([]byte("get@" + hex.EncodeToString(blsKey)))
}

func (s *stakingAuctionSC) getNewValidKeys(registeredKeys [][]byte, keysFromArgument [][]byte) ([][]byte, error) {
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
		vmOutput, err := s.getBLSRegisteredData(newKey)
		if err != nil ||
			(len(vmOutput.ReturnData) > 0 && len(vmOutput.ReturnData[0]) > 0) {
			return nil, vm.ErrKeyAlreadyRegistered
		}
	}

	return newKeys, nil
}

func (s *stakingAuctionSC) registerBLSKeys(
	registrationData *AuctionDataV2,
	pubKey []byte,
	ownerAddress []byte,
	args [][]byte,
) ([][]byte, error) {
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()
	if uint64(len(args)) < maxNodesToRun+1 {
		s.eei.AddReturnMessage(fmt.Sprintf("not enough arguments to process stake function: expected min %d, got %d", maxNodesToRun+1, len(args)))
		return nil, vm.ErrNotEnoughArgumentsToStake
	}

	blsKeys := s.getVerifiedBLSKeysFromArgs(pubKey, args)
	newKeys, err := s.getNewValidKeys(registrationData.BlsPubKeys, blsKeys)
	if err != nil {
		return nil, err
	}

	for _, blsKey := range newKeys {
		vmOutput, errExec := s.executeOnStakingSC([]byte("register@" +
			hex.EncodeToString(blsKey) + "@" +
			hex.EncodeToString(registrationData.RewardAddress) + "@" +
			hex.EncodeToString(ownerAddress) + "@",
		))
		if errExec != nil {
			s.eei.AddReturnMessage("cannot do register: " + errExec.Error())
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			return nil, err
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.AddReturnMessage("cannot do register: " + vmOutput.ReturnCode.String())
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			return nil, vm.ErrKeyAlreadyRegistered
		}

		registrationData.BlsPubKeys = append(registrationData.BlsPubKeys, blsKey)
	}

	return blsKeys, nil
}

func (s *stakingAuctionSC) updateStakeValue(registrationData *AuctionDataV2, caller []byte) vmcommon.ReturnCode {
	if len(registrationData.BlsPubKeys) == 0 && !core.IsSmartContractAddress(caller) {
		s.eei.AddReturnMessage("no bls keys has been provided")
		return vmcommon.UserError
	}

	err := s.saveRegistrationData(caller, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) getVerifiedBLSKeysFromArgs(txPubKey []byte, args [][]byte) [][]byte {
	blsKeys := make([][]byte, 0)
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()

	invalidBlsKeys := make([]string, 0)
	for i := uint64(1); i < maxNodesToRun*2+1; i += 2 {
		blsKey := args[i]
		signedMessage := args[i+1]
		err := s.sigVerifier.Verify(txPubKey, signedMessage, blsKey)
		if err != nil {
			invalidBlsKeys = append(invalidBlsKeys, hex.EncodeToString(blsKey))
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{invalidKey})
			continue
		}

		blsKeys = append(blsKeys, blsKey)
	}
	if len(invalidBlsKeys) != 0 {
		returnMessage := "invalid BLS keys: " + strings.Join(invalidBlsKeys, ", ")
		s.eei.AddReturnMessage(returnMessage)
	}

	return blsKeys
}

func (s *stakingAuctionSC) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	isGenesis := s.eei.BlockChainHook().CurrentNonce() == 0
	stakeEnabled := isGenesis || s.flagEnableStaking.IsSet()
	if !stakeEnabled {
		s.eei.AddReturnMessage(vm.StakeNotEnabled)
		return vmcommon.UserError
	}

	auctionConfig := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())
	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	registrationData.TotalStakeValue.Add(registrationData.TotalStakeValue, args.CallValue)
	if registrationData.TotalStakeValue.Cmp(auctionConfig.NodePrice) < 0 {
		s.eei.AddReturnMessage(
			fmt.Sprintf("insufficient stake value: expected %s, got %s",
				auctionConfig.NodePrice.String(),
				registrationData.TotalStakeValue.String(),
			),
		)
		return vmcommon.UserError
	}

	lenArgs := len(args.Arguments)
	if lenArgs == 0 {
		return s.updateStakeValue(registrationData, args.CallerAddr)
	}

	if !isNumArgsCorrectToStake(args.Arguments) {
		s.eei.AddReturnMessage("invalid number of arguments to call stake")
		return vmcommon.UserError
	}

	maxNodesToRun := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	if maxNodesToRun == 0 {
		s.eei.AddReturnMessage("number of nodes argument must be greater than zero")
		return vmcommon.UserError
	}

	err = s.eei.UseGas((maxNodesToRun - 1) * s.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	isAlreadyRegistered := len(registrationData.RewardAddress) > 0
	if !isAlreadyRegistered {
		registrationData.RewardAddress = args.CallerAddr
	}

	registrationData.MaxStakePerNode = big.NewInt(0).Set(registrationData.TotalStakeValue)
	registrationData.Epoch = s.eei.BlockChainHook().CurrentEpoch()

	blsKeys, err := s.registerBLSKeys(registrationData, args.CallerAddr, args.CallerAddr, args.Arguments)
	if err != nil {
		s.eei.AddReturnMessage("cannot register bls key: error " + err.Error())
		return vmcommon.UserError
	}

	numQualified := big.NewInt(0).Div(registrationData.TotalStakeValue, auctionConfig.NodePrice)
	if uint64(len(registrationData.BlsPubKeys)) > numQualified.Uint64() {
		s.eei.AddReturnMessage("insufficient funds")
		return vmcommon.OutOfFunds
	}

	// do the optionals - rewardAddress and maxStakePerNode
	if uint64(lenArgs) > maxNodesToRun*2+1 {
		for i := maxNodesToRun*2 + 1; i < uint64(lenArgs); i++ {
			if len(args.Arguments[i]) == s.walletAddressLen {
				if !isAlreadyRegistered {
					registrationData.RewardAddress = args.Arguments[i]
				} else {
					s.eei.AddReturnMessage("reward address after being registered can be changed only through changeRewardAddress")
				}
				continue
			}

			maxStakePerNode := big.NewInt(0).SetBytes(args.Arguments[i])
			registrationData.MaxStakePerNode.Set(maxStakePerNode)
		}
	}

	s.activateStakingFor(
		blsKeys,
		numQualified.Uint64(),
		registrationData,
		auctionConfig.NodePrice,
		registrationData.RewardAddress,
		args.CallerAddr,
	)

	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) activateStakingFor(
	blsKeys [][]byte,
	numQualified uint64,
	registrationData *AuctionDataV2,
	fixedStakeValue *big.Int,
	rewardAddress []byte,
	ownerAddress []byte,
) {
	numRegistered := uint64(registrationData.NumRegistered)
	for i := uint64(0); numRegistered < numQualified && i < uint64(len(blsKeys)); i++ {
		currentBLSKey := blsKeys[i]
		stakedData, err := s.getStakedData(currentBLSKey)
		if err != nil {
			continue
		}

		if stakedData.Staked || stakedData.Waiting {
			continue
		}

		vmOutput, err := s.executeOnStakingSC([]byte("stake@" +
			hex.EncodeToString(currentBLSKey) + "@" +
			hex.EncodeToString(rewardAddress) + "@" +
			hex.EncodeToString(ownerAddress),
		))
		if err != nil {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do stake for key %s, error %s", hex.EncodeToString(currentBLSKey), err.Error()))
			s.eei.Finish(currentBLSKey)
			s.eei.Finish([]byte{failed})
			continue

		}
		if vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do stake for key %s, error %s", hex.EncodeToString(currentBLSKey), vmOutput.ReturnCode.String()))
			s.eei.Finish(currentBLSKey)
			s.eei.Finish([]byte{failed})
			continue
		}

		if len(vmOutput.ReturnData) > 0 && bytes.Equal(vmOutput.ReturnData[0], []byte{waiting}) {
			s.eei.Finish(currentBLSKey)
			s.eei.Finish([]byte{waiting})
		}

		if stakedData.UnStakedNonce == 0 {
			numRegistered++
		}
	}

	registrationData.NumRegistered = uint32(numRegistered)
	registrationData.LockedStake.Mul(fixedStakeValue, big.NewInt(0).SetUint64(numRegistered))
}

func (s *stakingAuctionSC) executeOnStakingSC(data []byte) (*vmcommon.VMOutput, error) {
	return s.eei.ExecuteOnDestContext(s.stakingSCAddress, s.auctionSCAddress, big.NewInt(0), data)
}

func (s *stakingAuctionSC) setOwnerOfBlsKey(blsKey []byte, ownerAddress []byte) bool {
	vmOutput, err := s.executeOnStakingSC([]byte("setOwner@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(ownerAddress)))
	if err != nil {
		s.eei.AddReturnMessage(fmt.Sprintf("cannot set owner for key %s, error %s", hex.EncodeToString(blsKey), err.Error()))
		s.eei.Finish(blsKey)
		s.eei.Finish([]byte{failed})
		return false

	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		s.eei.AddReturnMessage(fmt.Sprintf("cannot set owner for key %s, error %s", hex.EncodeToString(blsKey), vmOutput.ReturnCode.String()))
		s.eei.Finish(blsKey)
		s.eei.Finish([]byte{failed})
		return false
	}

	return true
}

func (s *stakingAuctionSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	if len(args.Arguments) == 0 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	if !s.flagEnableStaking.IsSet() {
		s.eei.AddReturnMessage(vm.UnStakeNotEnabled)
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	err = s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnStake * uint64(len(args.Arguments)))
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	blsKeys := args.Arguments
	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return vmcommon.UserError
	}

	for _, blsKey := range blsKeys {
		vmOutput, errExec := s.executeOnStakingSC([]byte("unStake@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(registrationData.RewardAddress)))
		if errExec != nil {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do unStake for key %s: %s", hex.EncodeToString(blsKey), errExec.Error()))
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do unStake for key %s: %s", hex.EncodeToString(blsKey), vmOutput.ReturnCode.String()))
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}
	}

	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) unStakeNodeOneNodeFromStakingSC(blsKey []byte, rewardAddress []byte) vmcommon.ReturnCode {
	vmOutput, errExec := s.executeOnStakingSC([]byte("unStake@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(rewardAddress)))
	if errExec != nil {
		return vmcommon.UserError
	}
	return vmOutput.ReturnCode
}

func (s *stakingAuctionSC) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	if len(args.Arguments) == 0 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	if !s.flagEnableStaking.IsSet() {
		s.eei.AddReturnMessage(vm.UnBondNotEnabled)
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	blsKeys := args.Arguments
	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return vmcommon.UserError
	}

	err = s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnBond * uint64(len(args.Arguments)))
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	unBondedKeys := make([][]byte, 0)
	totalUnBond := big.NewInt(0)
	totalSlashed := big.NewInt(0)
	for _, blsKey := range blsKeys {
		nodeData, errGet := s.getStakedData(blsKey)
		if errGet != nil {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do unBond for key: %s, error: %s", hex.EncodeToString(blsKey), errGet.Error()))
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}
		// returns what value is still under the selected bls key
		vmOutput, errExec := s.executeOnStakingSC([]byte("unBond@" + hex.EncodeToString(blsKey)))
		if errExec != nil || vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do unBond for key: %s", hex.EncodeToString(blsKey)))
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}

		registrationData.NumRegistered -= 1
		auctionConfig := s.getConfig(nodeData.UnStakedEpoch)
		unBondedKeys = append(unBondedKeys, blsKey)
		totalUnBond.Add(totalUnBond, auctionConfig.NodePrice)
		totalSlashed.Add(totalSlashed, nodeData.SlashValue)
	}

	totalUnBond.Sub(totalUnBond, totalSlashed)
	if totalUnBond.Cmp(zero) < 0 {
		totalUnBond.Set(zero)
	}

	returnCode := s.updateRegistrationDataAfterUnBond(registrationData, totalUnBond, unBondedKeys, args.CallerAddr)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBond, nil, 0)
	if err != nil {
		s.eei.AddReturnMessage("transfer error on unBond function")
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) updateRegistrationDataAfterUnBond(
	registrationData *AuctionDataV2,
	totalUnBond *big.Int,
	unBondedKeys [][]byte,
	callerAddr []byte,
) vmcommon.ReturnCode {
	if registrationData.LockedStake.Cmp(totalUnBond) < 0 {
		s.eei.AddReturnMessage("contract error on unBond function, lockedStake < totalUnBond")
		return vmcommon.UserError
	}

	registrationData.LockedStake.Sub(registrationData.LockedStake, totalUnBond)
	registrationData.TotalStakeValue.Sub(registrationData.TotalStakeValue, totalUnBond)
	if registrationData.TotalStakeValue.Cmp(zero) < 0 {
		s.eei.AddReturnMessage("contract error on unBond function, total stake < 0")
		return vmcommon.UserError
	}

	if registrationData.LockedStake.Cmp(zero) == 0 && registrationData.TotalStakeValue.Cmp(zero) == 0 {
		s.eei.SetStorage(callerAddr, nil)
	} else {
		s.deleteUnBondedKeys(registrationData, unBondedKeys)
		errSave := s.saveRegistrationData(callerAddr, registrationData)
		if errSave != nil {
			s.eei.AddReturnMessage("cannot save registration data: error " + errSave.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) deleteUnBondedKeys(registrationData *AuctionDataV2, unBondedKeys [][]byte) {
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

func (s *stakingAuctionSC) claim(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.flagEnableTopUp.IsSet() {
		//claim function will become unavailable after enabling staking v2
		s.eei.AddReturnMessage("claim function is disabled")
		return vmcommon.UserError
	}

	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get registration data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("key is not registered, claim is not possible")
		return vmcommon.UserError
	}
	err = s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Claim)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	claimable := big.NewInt(0).Sub(registrationData.TotalStakeValue, registrationData.LockedStake)
	if claimable.Cmp(zero) <= 0 {
		return vmcommon.Ok
	}

	registrationData.TotalStakeValue.Set(registrationData.LockedStake)
	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, claimable, nil, 0)
	if err != nil {
		s.eei.AddReturnMessage("transfer error on finalizeUnStake function: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) unStakeTokens(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, returnCode := s.basicCheckForUnStakeUnBond(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnStakeTokens)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("should have specified one argument containing the unstake value")
		return vmcommon.UserError
	}

	maxValUnstake := big.NewInt(0).Sub(registrationData.TotalStakeValue, registrationData.LockedStake)
	unStakeValue := big.NewInt(0).SetBytes(args.Arguments[0])
	unstakeValueIsOk := unStakeValue.Cmp(s.minUnstakeTokensValue) >= 0 || unStakeValue.Cmp(maxValUnstake) == 0
	if !unstakeValueIsOk {
		s.eei.AddReturnMessage("can not unstake the provided value either because is under the minimum threshold or " +
			"is not the value left to be unStaked")
		return vmcommon.UserError
	}
	if unStakeValue.Cmp(maxValUnstake) > 0 {
		s.eei.AddReturnMessage("can not unstake a bigger value than the possible allowed value which is " + maxValUnstake.String())
		return vmcommon.UserError
	}

	returnCode = s.unStakeValue(registrationData, unStakeValue)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) unStakeValue(registrationData *AuctionDataV2, unStakeValue *big.Int) vmcommon.ReturnCode {
	registrationData.TotalStakeValue.Sub(registrationData.TotalStakeValue, unStakeValue)
	if registrationData.TotalStakeValue.Cmp(zero) < 0 {
		s.eei.AddReturnMessage("contract error on unStakeTokens function, total stake < unstake value")
		return vmcommon.UserError
	}

	registrationData.TotalUnstaked.Add(registrationData.TotalUnstaked, unStakeValue)
	registrationData.UnstakedInfo = append(
		registrationData.UnstakedInfo,
		&UnstakedValue{
			UnstakedNonce: s.eei.BlockChainHook().CurrentNonce(),
			UnstakedValue: unStakeValue,
		},
	)
	return vmcommon.Ok
}

func (s *stakingAuctionSC) basicCheckForUnStakeUnBond(args *vmcommon.ContractCallInput) (*AuctionDataV2, vmcommon.ReturnCode) {
	if !s.flagEnableTopUp.IsSet() {
		s.eei.AddReturnMessage("invalid method to call")
		return nil, vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return nil, vmcommon.UserError
	}
	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get registration data: error " + err.Error())
		return nil, vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("key is not registered, auction operation is not possible")
		return nil, vmcommon.UserError
	}
	if registrationData.TotalUnstaked == nil {
		registrationData.TotalUnstaked = big.NewInt(0)
	}

	return registrationData, vmcommon.Ok
}

func (s *stakingAuctionSC) unStakeTokensWithNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, returnCode := s.basicCheckForUnStakeUnBond(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnStakeTokens)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("should have specified one argument containing the unstake value")
		return vmcommon.UserError
	}
	unStakeValue := big.NewInt(0).SetBytes(args.Arguments[0])
	if unStakeValue.Cmp(zero) <= 0 {
		s.eei.AddReturnMessage("cannot unStake negative value")
		return vmcommon.UserError
	}

	maxValUnStake := big.NewInt(0).Sub(registrationData.TotalStakeValue, registrationData.LockedStake)
	valToUnStakeNodes := big.NewInt(0).Sub(unStakeValue, maxValUnStake)

	valToUnStakeFromTopUp := big.NewInt(0)
	if maxValUnStake.Cmp(zero) >= 0 {
		valToUnStakeFromTopUp.Set(maxValUnStake)
		if maxValUnStake.Cmp(unStakeValue) > 0 {
			valToUnStakeFromTopUp.Set(unStakeValue)
		}
		returnCode = s.unStakeValue(registrationData, valToUnStakeFromTopUp)
		if returnCode != vmcommon.Ok {
			return returnCode
		}
	}

	totalUnStakedFromNodes := big.NewInt(0)
	if valToUnStakeNodes.Cmp(zero) > 0 {
		totalUnStakedFromNodes, err = s.unStakeNodesWithPreferences(registrationData, valToUnStakeNodes)
		if err != nil {
			s.eei.AddReturnMessage("cannot unStake with preference error " + err.Error())
			return vmcommon.UserError
		}
	}

	totalUnStaked := big.NewInt(0).Add(valToUnStakeFromTopUp, totalUnStakedFromNodes)
	s.eei.Finish(totalUnStaked.Bytes())
	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) isInAdditionalQueue(blsKey []byte) (bool, error) {
	stakedData, err := s.getStakedData(blsKey)
	if err != nil {
		return false, err
	}
	return stakedData.Waiting, nil
}

func (s *stakingAuctionSC) isWithStatusWaiting(blsKey []byte) (bool, error) {
	return s.eei.StatusFromValidatorStatistics(blsKey) == string(core.WaitingList), nil
}

func (s *stakingAuctionSC) isWithStatusEligible(blsKey []byte) (bool, error) {
	return s.eei.StatusFromValidatorStatistics(blsKey) == string(core.EligibleList), nil
}

func (s *stakingAuctionSC) isWithStatusInactive(blsKey []byte) bool {
	return s.eei.StatusFromValidatorStatistics(blsKey) == string(core.InactiveList)
}

func searchInList(key []byte, list [][]byte) bool {
	for _, element := range list {
		if bytes.Equal(key, element) {
			return true
		}
	}
	return false
}

func (s *stakingAuctionSC) unStakeNodesWithStatus(
	registrationData *AuctionDataV2,
	listUnStakedKeys [][]byte,
	numNodesToUnStake int,
	checkStatus func(blsKey []byte) (bool, error),
) ([][]byte, error) {
	if len(listUnStakedKeys) >= numNodesToUnStake {
		return listUnStakedKeys, nil
	}

	for _, blsKey := range registrationData.BlsPubKeys {
		if searchInList(blsKey, listUnStakedKeys) {
			continue
		}

		okStatus, err := checkStatus(blsKey)
		if err != nil {
			return nil, err
		}
		if !okStatus {
			continue
		}

		err = s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnStake)
		if err != nil {
			return nil, err
		}

		returnCode := s.unStakeNodeOneNodeFromStakingSC(blsKey, registrationData.RewardAddress)
		if returnCode != vmcommon.Ok {
			continue
		}

		listUnStakedKeys = append(listUnStakedKeys, blsKey)
		if len(listUnStakedKeys) >= numNodesToUnStake {
			break
		}
	}

	return listUnStakedKeys, nil
}

func (s *stakingAuctionSC) unStakeNodesWithPreferences(registrationData *AuctionDataV2, valToUnStakeNodes *big.Int) (*big.Int, error) {
	divValue, modValue := big.NewInt(0).DivMod(valToUnStakeNodes, s.baseConfig.NodePrice, big.NewInt(0))
	numNodesToUnStake := int(divValue.Uint64())
	if modValue.Cmp(zero) > 0 {
		numNodesToUnStake += 1
	}

	listUnStakedKeys := make([][]byte, 0)

	var err error
	listUnStakedKeys, err = s.unStakeNodesWithStatus(registrationData, listUnStakedKeys, numNodesToUnStake, s.isInAdditionalQueue)
	if err != nil {
		return nil, err
	}

	listUnStakedKeys, err = s.unStakeNodesWithStatus(registrationData, listUnStakedKeys, numNodesToUnStake, s.isWithStatusWaiting)
	if err != nil {
		return nil, err
	}

	listUnStakedKeys, err = s.unStakeNodesWithStatus(registrationData, listUnStakedKeys, numNodesToUnStake, s.isWithStatusEligible)
	if err != nil {
		return nil, err
	}

	for _, unStakedKey := range listUnStakedKeys {
		s.eei.Finish(unStakedKey)
	}

	numUnStaked := big.NewInt(0).SetUint64(uint64(len(listUnStakedKeys)))
	unStakedFromNodes := big.NewInt(0).Mul(numUnStaked, s.baseConfig.NodePrice)
	return unStakedFromNodes, nil
}

func (s *stakingAuctionSC) unBondTokens(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, returnCode := s.basicCheckForUnStakeUnBond(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnBondTokens)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 0 {
		s.eei.AddReturnMessage("should have not specified any arguments")
		return vmcommon.UserError
	}

	totalUnBond, index := s.computeUnBondTokens(registrationData)
	if totalUnBond.Cmp(zero) == 0 {
		s.eei.AddReturnMessage("no tokens that can be unbond at this time")
		return vmcommon.Ok
	}

	registrationData.UnstakedInfo = registrationData.UnstakedInfo[index:]
	registrationData.TotalUnstaked.Sub(registrationData.TotalUnstaked, totalUnBond)
	if registrationData.TotalUnstaked.Cmp(zero) < 0 {
		s.eei.AddReturnMessage("contract error on unBondTokens function, total unstaked < total unbond")
		return vmcommon.UserError
	}

	err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBond, nil, 0)
	if err != nil {
		s.eei.AddReturnMessage("transfer error on unBond function")
		return vmcommon.UserError
	}

	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) unBondNecessaryNodes(
	registrationData *AuctionDataV2,
	valueToUnBond *big.Int,
) (*big.Int, [][]byte, error) {
	totalUnBond := big.NewInt(0)
	unBondedNodes := make([][]byte, 0)
	for _, blsKey := range registrationData.BlsPubKeys {
		if !s.isWithStatusInactive(blsKey) {
			continue
		}

		nodeData, errGet := s.getStakedData(blsKey)
		if errGet != nil {
			return nil, nil, errGet
		}
		if len(nodeData.RewardAddress) == 0 {
			continue
		}
		// returns what value is still under the selected bls key
		vmOutput, errExec := s.executeOnStakingSC([]byte("unBond@" + hex.EncodeToString(blsKey)))
		if errExec != nil || vmOutput.ReturnCode != vmcommon.Ok {
			continue
		}

		registrationData.NumRegistered -= 1
		totalUnBond.Add(totalUnBond, s.baseConfig.NodePrice)
		totalUnBond.Sub(totalUnBond, nodeData.SlashValue)

		s.eei.Finish(blsKey)
		unBondedNodes = append(unBondedNodes, blsKey)
		if totalUnBond.Cmp(valueToUnBond) >= 0 {
			break
		}
	}

	return totalUnBond, unBondedNodes, nil
}

func (s *stakingAuctionSC) unBondTokensWithNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, returnCode := s.basicCheckForUnStakeUnBond(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnBondTokens)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("invalid num of arguments")
		return vmcommon.UserError
	}

	unBondValue := big.NewInt(0).SetBytes(args.Arguments[0])
	if unBondValue.Cmp(zero) <= 0 {
		s.eei.AddReturnMessage("cannot unBond negative value")
		return vmcommon.UserError
	}

	totalUnBond, index := s.computeUnBondTokens(registrationData)
	if totalUnBond.Cmp(zero) > 0 {
		registrationData.UnstakedInfo = registrationData.UnstakedInfo[index:]
		registrationData.TotalUnstaked.Sub(registrationData.TotalUnstaked, totalUnBond)
		if registrationData.TotalUnstaked.Cmp(zero) < 0 {
			s.eei.AddReturnMessage("contract error on unBondTokens function, total unstaked < total unbond")
			return vmcommon.UserError
		}
	}

	unBondedNodes := make([][]byte, 0)
	totalUnBondedFromNodes := big.NewInt(0)
	if totalUnBond.Cmp(unBondValue) < 0 {
		remainingToUnBond := big.NewInt(0).Sub(unBondValue, totalUnBond)
		totalUnBondedFromNodes, unBondedNodes, err = s.unBondNecessaryNodes(registrationData, remainingToUnBond)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	totalUnBond.Add(totalUnBond, totalUnBondedFromNodes)

	s.eei.Finish(totalUnBond.Bytes())
	err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBond, nil, 0)
	if err != nil {
		s.eei.AddReturnMessage("transfer error on unBond function")
		return vmcommon.UserError
	}

	returnCode = s.updateRegistrationDataAfterUnBond(registrationData, totalUnBond, unBondedNodes, args.CallerAddr)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) computeUnBondTokens(registrationData *AuctionDataV2) (*big.Int, int) {
	var unstakedValue *UnstakedValue
	currentNonce := s.eei.BlockChainHook().CurrentNonce()
	totalUnBond := big.NewInt(0)
	index := 0
	for _, unstakedValue = range registrationData.UnstakedInfo {
		canUnbond := currentNonce-unstakedValue.UnstakedNonce >= s.unBondPeriod
		if !canUnbond {
			break
		}

		totalUnBond.Add(totalUnBond, unstakedValue.UnstakedValue)
		index++
	}

	return totalUnBond, index
}

func (s *stakingAuctionSC) calculateNodePrice(bids []AuctionDataV2) (*big.Int, error) {
	auctionConfig := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())

	minNodePrice := big.NewInt(0).Set(auctionConfig.MinStakeValue)
	maxNodePrice := big.NewInt(0).Div(auctionConfig.TotalSupply, big.NewInt(int64(auctionConfig.NumNodes)))
	numNodes := auctionConfig.NumNodes

	for nodePrice := maxNodePrice; nodePrice.Cmp(minNodePrice) >= 0; nodePrice.Sub(nodePrice, auctionConfig.MinStep) {
		qualifiedNodes := calcNumQualifiedNodes(nodePrice, bids)
		if qualifiedNodes >= numNodes {
			return nodePrice, nil
		}
	}

	return nil, vm.ErrNotEnoughQualifiedNodes
}

func (s *stakingAuctionSC) selection(bids []AuctionDataV2) [][]byte {
	nodePrice, err := s.calculateNodePrice(bids)
	if err != nil {
		return nil
	}

	totalQualifyingStake := big.NewFloat(0).SetInt(calcTotalQualifyingStake(nodePrice, bids))

	reservePool := make(map[string]float64)
	toBeSelectedRandom := make(map[string]float64)
	finalSelectedNodes := make([][]byte, 0)
	for _, validator := range bids {
		if validator.MaxStakePerNode.Cmp(nodePrice) < 0 {
			continue
		}

		numAllocatedNodes, allocatedNodes := calcNumAllocatedAndProportion(validator, nodePrice, totalQualifyingStake)
		if numAllocatedNodes < uint64(len(validator.BlsPubKeys)) {
			selectorProp := allocatedNodes - float64(numAllocatedNodes)
			if numAllocatedNodes == 0 {
				toBeSelectedRandom[string(validator.BlsPubKeys[numAllocatedNodes])] = selectorProp
			}

			for i := numAllocatedNodes + 1; i < uint64(len(validator.BlsPubKeys)); i++ {
				reservePool[string(validator.BlsPubKeys[i])] = selectorProp
			}
		}

		if numAllocatedNodes > 0 {
			finalSelectedNodes = append(finalSelectedNodes, validator.BlsPubKeys[:numAllocatedNodes]...)
		}
	}

	randomlySelected := s.fillRemainingSpace(uint32(len(finalSelectedNodes)), toBeSelectedRandom, reservePool)
	finalSelectedNodes = append(finalSelectedNodes, randomlySelected...)

	return finalSelectedNodes
}

func (s *stakingAuctionSC) fillRemainingSpace(
	alreadySelected uint32,
	toBeSelectedRandom map[string]float64,
	reservePool map[string]float64,
) [][]byte {
	auctionConfig := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())
	stillNeeded := uint32(0)
	if auctionConfig.NumNodes > alreadySelected {
		stillNeeded = auctionConfig.NumNodes - alreadySelected
	}

	randomlySelected := s.selectRandomly(toBeSelectedRandom, stillNeeded)
	alreadySelected += uint32(len(randomlySelected))

	if auctionConfig.NumNodes > alreadySelected {
		stillNeeded = auctionConfig.NumNodes - alreadySelected
		randomlySelected = append(randomlySelected, s.selectRandomly(reservePool, stillNeeded)...)
	}

	return randomlySelected
}

func (s *stakingAuctionSC) selectRandomly(selectable map[string]float64, numNeeded uint32) [][]byte {
	randomlySelected := make([][]byte, 0)
	if numNeeded == 0 {
		return randomlySelected
	}

	expandedList := make([]string, 0)
	for key, proportion := range selectable {
		expandVal := uint8(proportion)*10 + 1
		for i := uint8(0); i < expandVal; i++ {
			expandedList = append(expandedList, key)
		}
	}

	random := s.eei.BlockChainHook().CurrentRandomSeed()
	shuffleList(expandedList, random)

	selectedKeys := make(map[string]struct{})
	selected := uint32(0)
	for i := 0; selected < numNeeded && i < len(expandedList); i++ {
		_, exists := selectedKeys[expandedList[i]]
		if !exists {
			selected++
			selectedKeys[expandedList[i]] = struct{}{}
		}
	}

	for key := range selectedKeys {
		randomlySelected = append(randomlySelected, []byte(key))
	}

	return randomlySelected
}

func (s *stakingAuctionSC) getTotalStaked(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("caller not registered in staking/auction sc")
		return vmcommon.UserError
	}

	s.eei.Finish([]byte(registrationData.TotalStakeValue.String()))
	return vmcommon.Ok
}

func (s *stakingAuctionSC) getTopUp(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.flagEnableTopUp.IsSet() {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		s.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("caller not registered in staking/auction sc")
		return vmcommon.UserError
	}
	registrationData.TotalStakeValue.Sub(registrationData.TotalStakeValue, registrationData.LockedStake)
	if registrationData.TotalStakeValue.Cmp(zero) < 0 {
		s.eei.AddReturnMessage("contract error on getTopUp function, total stake < locked stake value")
		return vmcommon.UserError
	}

	s.eei.Finish([]byte(registrationData.TotalStakeValue.String()))
	return vmcommon.Ok
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (s *stakingAuctionSC) EpochConfirmed(epoch uint32) {
	s.flagEnableStaking.Toggle(epoch >= s.enableStakingEpoch)
	log.Debug("stakingAuctionSC: stake/unstake/unbond", "enabled", s.flagEnableStaking.IsSet())

	s.flagEnableTopUp.Toggle(epoch >= s.stakingV2Epoch)
	log.Debug("stakingAuctionSC: top up mechanism", "enabled", s.flagEnableTopUp.IsSet())
}

// CanUseContract returns true if contract can be used
func (s *stakingAuctionSC) CanUseContract() bool {
	return true
}

func (s *stakingAuctionSC) getBlsKeysStatus(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.auctionSCAddress) {
		s.eei.AddReturnMessage("this is only a view function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("number of arguments must be equal to 1")
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegistrationData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	if len(registrationData.BlsPubKeys) == 0 {
		s.eei.AddReturnMessage("no bls keys")
		return vmcommon.Ok
	}

	for _, blsKey := range registrationData.BlsPubKeys {
		vmOutput, errExec := s.executeOnStakingSC([]byte("getBLSKeyStatus@" + hex.EncodeToString(blsKey)))
		if errExec != nil {
			s.eei.AddReturnMessage("cannot get bls key status: bls key - " + hex.EncodeToString(blsKey) + " error - " + errExec.Error())
			continue
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.AddReturnMessage("error in getting bls key status: bls key - " + hex.EncodeToString(blsKey))
			continue
		}

		if len(vmOutput.ReturnData) != 1 {
			s.eei.AddReturnMessage("cannot get bls key status for key " + hex.EncodeToString(blsKey))
			continue
		}

		s.eei.Finish(blsKey)
		s.eei.Finish(vmOutput.ReturnData[0])
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) updateStakingV2(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.flagEnableTopUp.IsSet() {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.auctionSCAddress) {
		s.eei.AddReturnMessage("this is a function that has to be called internally")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("should have provided only one argument: the owner address")
		return vmcommon.UserError
	}
	if len(args.Arguments[0]) != s.walletAddressLen {
		s.eei.AddReturnMessage("wrong owner address")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	registrationData, err := s.getOrCreateRegistrationData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage("cannot get registration data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("key is not registered, updateStakingV2 is not possible")
		return vmcommon.UserError
	}
	for _, blsKey := range registrationData.BlsPubKeys {
		okSet := s.setOwnerOfBlsKey(blsKey, args.Arguments[0])
		if !okSet {
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (s *stakingAuctionSC) IsInterfaceNil() bool {
	return s == nil
}
