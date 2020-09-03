//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. auction.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
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
	eei                vm.SystemEI
	unBondPeriod       uint64
	sigVerifier        vm.MessageSignVerifier
	baseConfig         AuctionConfig
	enableAuctionNonce uint64
	stakingSCAddress   []byte
	auctionSCAddress   []byte
	enableStakingNonce uint64
	gasCost            vm.GasCost
	marshalizer        marshal.Marshalizer
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
		return nil, vm.ErrInvalidMinNumberOfNodes
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}

	baseConfig := AuctionConfig{
		NumNodes:    uint32(args.NumOfNodesToSelect),
		TotalSupply: big.NewInt(0).Set(args.GenesisTotalSupply),
	}
	conversionBase := 10
	ok := true
	baseConfig.UnJailPrice, ok = big.NewInt(0).SetString(args.StakingSCConfig.UnJailValue, conversionBase)
	if !ok || baseConfig.UnJailPrice.Cmp(zero) <= 0 {
		return nil, vm.ErrInvalidUnJailCost
	}
	baseConfig.MinStakeValue, ok = big.NewInt(0).SetString(args.StakingSCConfig.MinStakeValue, conversionBase)
	if !ok || baseConfig.MinStakeValue.Cmp(zero) <= 0 {
		return nil, vm.ErrNegativeInitialStakeValue
	}
	baseConfig.NodePrice, ok = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !ok || baseConfig.NodePrice.Cmp(zero) < 0 {
		return nil, vm.ErrNegativeInitialStakeValue
	}
	baseConfig.MinStep, ok = big.NewInt(0).SetString(args.StakingSCConfig.MinStepValue, conversionBase)
	if !ok || baseConfig.NodePrice.Cmp(zero) < 0 {
		return nil, vm.ErrNegativeInitialStakeValue
	}

	reg := &stakingAuctionSC{
		eei:                args.Eei,
		unBondPeriod:       args.StakingSCConfig.UnBondPeriod,
		sigVerifier:        args.SigVerifier,
		baseConfig:         baseConfig,
		enableAuctionNonce: args.StakingSCConfig.AuctionEnableNonce,
		enableStakingNonce: args.StakingSCConfig.StakeEnableNonce,
		stakingSCAddress:   args.StakingSCAddress,
		auctionSCAddress:   args.AuctionSCAddress,
		gasCost:            args.GasCost,
		marshalizer:        args.Marshalizer,
	}
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
	// TODO: activate this after validation
	//case "changeValidatorKeys":
	//	return s.changeValidatorKeys(args)
	case "unJail":
		return s.unJail(args)
	case "getTotalStaked":
		return s.getTotalStaked(args)
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

func (s *stakingAuctionSC) unJail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) == 0 || len(args.Arguments)%2 != 0 {
		s.eei.AddReturnMessage("invalid number of arguments: expected even number")
		return vmcommon.UserError
	}

	numBLSKeys := len(args.Arguments) / 2
	auctionConfig := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())
	totalUnJailPrice := big.NewInt(0).Mul(auctionConfig.UnJailPrice, big.NewInt(int64(numBLSKeys)))

	if totalUnJailPrice.Cmp(args.CallValue) != 0 {
		s.eei.AddReturnMessage("wanted exact unjail price * numNodes")
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnJail * uint64(numBLSKeys))
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	blsKeys := make([][]byte, numBLSKeys)
	for i := 0; i < len(args.Arguments); i += 2 {
		blsKeys[i/2] = args.Arguments[i]
	}

	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		s.eei.AddReturnMessage("could not get all blsKeys from registration data: " + err.Error())
		return vmcommon.UserError
	}

	transferBack := big.NewInt(0)
	for i, blsKey := range blsKeys {
		vmOutput, err := s.executeOnStakingSC([]byte("unJail@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(args.Arguments[2*i+1])))
		if err != nil || vmOutput.ReturnCode != vmcommon.Ok {
			transferBack.Add(transferBack, auctionConfig.UnJailPrice)
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}
	}

	err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, transferBack, nil, 0)
	if err != nil {
		s.eei.AddReturnMessage("transfer error on unJail function")
		return vmcommon.UserError
	}

	finalUnJailFunds := big.NewInt(0).Sub(args.CallValue, transferBack)
	s.addToUnJailFunds(finalUnJailFunds)

	return vmcommon.Ok
}

func (s *stakingAuctionSC) changeRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage("transaction value must be zero")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}
	if len(args.Arguments[0]) != len(args.CallerAddr) {
		s.eei.AddReturnMessage("wrong reward address")
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
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
		s.eei.AddReturnMessage("insufficient gas limit")
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
		s.eei.AddReturnMessage("transaction value must be zero")
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
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
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

		err := s.sigVerifier.Verify(args.CallerAddr, signedWithNewKey, newBlsKey)
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

func (s *stakingAuctionSC) replaceBLSKey(registrationData *AuctionData, oldBlsKey []byte, newBlsKey []byte) error {
	foundOldKey := false
	for i, registeredKey := range registrationData.BlsPubKeys {
		if bytes.Equal(registeredKey, oldBlsKey) {
			foundOldKey = true
			registrationData.BlsPubKeys[i] = newBlsKey
			break
		}
	}

	if !foundOldKey {
		return vm.ErrBLSPublicKeyMissmatch
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
		s.eei.AddReturnMessage("transaction value must be zero")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected exactly %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	value := s.eei.GetStorage(args.Arguments[0])
	s.eei.Finish(value)

	return vmcommon.Ok
}

func (s *stakingAuctionSC) setConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := s.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		s.eei.AddReturnMessage("setConfig function was not called by the owner address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 7 {
		retMessage := fmt.Sprintf("setConfig function called with wrong number of arguments expected %d, got %d", 7, len(args.Arguments))
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}

	auctionConfig := &AuctionConfig{
		MinStakeValue: big.NewInt(0).SetBytes(args.Arguments[0]),
		NumNodes:      uint32(big.NewInt(0).SetBytes(args.Arguments[1]).Uint64()),
		TotalSupply:   big.NewInt(0).SetBytes(args.Arguments[2]),
		MinStep:       big.NewInt(0).SetBytes(args.Arguments[3]),
		NodePrice:     big.NewInt(0).SetBytes(args.Arguments[4]),
		UnJailPrice:   big.NewInt(0).SetBytes(args.Arguments[5]),
	}

	configData, err := s.marshalizer.Marshal(auctionConfig)
	if err != nil {
		s.eei.AddReturnMessage("setConfig marshal auctionConfig error")
		return vmcommon.UserError
	}
	epochBytes := args.Arguments[6]
	s.eei.SetStorage(epochBytes, configData)

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

func (s *stakingAuctionSC) getConfig(epoch uint32) AuctionConfig {
	epochKey := big.NewInt(int64(epoch)).Bytes()
	configData := s.eei.GetStorage(epochKey)
	if len(configData) == 0 {
		return s.baseConfig
	}

	auctionConfig := &AuctionConfig{}
	err := s.marshalizer.Unmarshal(auctionConfig, configData)
	if err != nil {
		log.Warn("unmarshal error on getConfig function, returning baseConfig",
			"error", err.Error(),
		)
		return s.baseConfig
	}

	if s.checkConfigCorrectness(*auctionConfig) != nil {
		baseConfigData, err := s.marshalizer.Marshal(&s.baseConfig)
		if err != nil {
			log.Warn("marshal error on getConfig function, returning baseConfig")
			return s.baseConfig
		}
		s.eei.SetStorage(epochKey, baseConfigData)
		return s.baseConfig
	}

	return *auctionConfig
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
		if _, ok := registeredKeysMap[string(keysFromArgument[i])]; ok {
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
	registrationData *AuctionData,
	pubKey []byte,
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
		vmOutput, err := s.executeOnStakingSC([]byte("register@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(registrationData.RewardAddress)))
		if err != nil {
			s.eei.AddReturnMessage("cannot do register: " + err.Error())
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			return nil, nil
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.AddReturnMessage("cannot do register: " + vmOutput.ReturnCode.String())
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			return nil, nil
		}

		registrationData.BlsPubKeys = append(registrationData.BlsPubKeys, blsKey)
	}

	return blsKeys, nil
}

func (s *stakingAuctionSC) updateStakeValue(registrationData *AuctionData, caller []byte) vmcommon.ReturnCode {
	if len(registrationData.BlsPubKeys) == 0 {
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

func (s *stakingAuctionSC) isFunctionEnabled(nonce uint64) bool {
	currentNonce := s.eei.BlockChainHook().CurrentNonce()
	if currentNonce == 0 {
		return true
	}

	if currentNonce < nonce {
		return false
	}

	return true
}

func (s *stakingAuctionSC) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	isStakeEnabled := s.isFunctionEnabled(s.enableStakingNonce)
	if !isStakeEnabled {
		s.eei.AddReturnMessage("stake is not enabled")
		return vmcommon.UserError
	}

	auctionConfig := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())
	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	registrationData.TotalStakeValue.Add(registrationData.TotalStakeValue, args.CallValue)
	if registrationData.TotalStakeValue.Cmp(auctionConfig.NodePrice) < 0 {
		s.eei.AddReturnMessage(
			fmt.Sprintf("insufficient stake value expected %s, got %s",
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
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	isAlreadyRegistered := len(registrationData.RewardAddress) > 0
	if !isAlreadyRegistered {
		registrationData.RewardAddress = args.CallerAddr
	}

	registrationData.MaxStakePerNode = big.NewInt(0).Set(registrationData.TotalStakeValue)
	registrationData.Epoch = s.eei.BlockChainHook().CurrentEpoch()

	blsKeys, err := s.registerBLSKeys(registrationData, args.CallerAddr, args.Arguments)
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
			if len(args.Arguments[i]) == len(args.CallerAddr) {
				if !isAlreadyRegistered {
					registrationData.RewardAddress = args.Arguments[i]
				} else {
					s.eei.AddReturnMessage("reward address after being registered can be changed only through changeRewardAddress")
				}
				continue
			}

			maxStakePerNode, ok := big.NewInt(0).SetString(string(args.Arguments[i]), conversionBase)
			if !ok {
				continue
			}
			registrationData.MaxStakePerNode.Set(maxStakePerNode)
		}
	}

	currentNonce := s.eei.BlockChainHook().CurrentNonce()
	if currentNonce == 0 || !s.isFunctionEnabled(s.enableAuctionNonce) {
		s.activateStakingFor(
			blsKeys,
			numQualified.Uint64(),
			registrationData,
			auctionConfig.NodePrice,
			registrationData.RewardAddress,
		)
	}

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
	registrationData *AuctionData,
	fixedStakeValue *big.Int,
	rewardAddress []byte,
) {
	numRegistered := uint64(registrationData.NumRegistered)
	for i := uint64(0); numRegistered < numQualified && i < uint64(len(blsKeys)); i++ {
		stakedData, err := s.getStakedData(blsKeys[i])
		if err != nil {
			continue
		}

		if stakedData.Staked || stakedData.Waiting {
			continue
		}

		vmOutput, err := s.executeOnStakingSC([]byte("stake@" + hex.EncodeToString(blsKeys[i]) + "@" + hex.EncodeToString(rewardAddress)))
		if err != nil {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do stake for key %s, error %s", hex.EncodeToString(blsKeys[i]), err.Error()))
			s.eei.Finish(blsKeys[i])
			s.eei.Finish([]byte{failed})
			continue
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do stake for key %s, error %s", hex.EncodeToString(blsKeys[i]), vmOutput.ReturnCode.String()))
			s.eei.Finish(blsKeys[i])
			s.eei.Finish([]byte{failed})
			continue
		}

		if len(vmOutput.ReturnData) > 0 && bytes.Equal(vmOutput.ReturnData[0], []byte{waiting}) {
			s.eei.Finish(blsKeys[i])
			s.eei.Finish([]byte{waiting})
		}

		numRegistered++
	}

	registrationData.NumRegistered = uint32(numRegistered)
	registrationData.LockedStake.Mul(fixedStakeValue, big.NewInt(0).SetUint64(numRegistered))
}

func (s *stakingAuctionSC) executeOnStakingSC(data []byte) (*vmcommon.VMOutput, error) {
	return s.eei.ExecuteOnDestContext(s.stakingSCAddress, s.auctionSCAddress, big.NewInt(0), data)
}

func (s *stakingAuctionSC) getOrCreateRegistrationData(key []byte) (*AuctionData, error) {
	data := s.eei.GetStorage(key)
	registrationData := &AuctionData{
		RewardAddress:   nil,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0),
		LockedStake:     big.NewInt(0),
		MaxStakePerNode: big.NewInt(0),
	}

	if len(data) > 0 {
		err := s.marshalizer.Unmarshal(registrationData, data)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}
	}

	return registrationData, nil
}

func (s *stakingAuctionSC) saveRegistrationData(key []byte, auction *AuctionData) error {
	data, err := s.marshalizer.Marshal(auction)
	if err != nil {
		log.Debug("marshal error on staking SC stake function ",
			"error", err.Error(),
		)
		return err
	}

	s.eei.SetStorage(key, data)
	return nil
}

func (s *stakingAuctionSC) getStakedData(key []byte) (*StakedData, error) {
	data := s.eei.GetStorageFromAddress(s.stakingSCAddress, key)
	stakedData := &StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: nil,
		StakeValue:    big.NewInt(0),
		SlashValue:    big.NewInt(0),
	}

	if len(data) > 0 {
		err := s.marshalizer.Unmarshal(stakedData, data)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}

		if stakedData.SlashValue == nil {
			stakedData.SlashValue = big.NewInt(0)
		}
	}

	return stakedData, nil
}

func (s *stakingAuctionSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage("transaction value must be zero")
		return vmcommon.UserError
	}
	if len(args.Arguments) == 0 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	isUnStakeEnabled := s.isFunctionEnabled(s.enableStakingNonce)
	if !isUnStakeEnabled {
		s.eei.AddReturnMessage("unStake is not enabled")
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	err = s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnStake * uint64(len(args.Arguments)))
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	blsKeys := args.Arguments
	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		s.eei.AddReturnMessage("could not get all blsKeys from registration data: " + err.Error())
		return vmcommon.UserError
	}

	for _, blsKey := range blsKeys {
		if registrationData.NumRegistered == 0 {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do unStake for key %s as it is not staked", hex.EncodeToString(blsKey)))
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}

		vmOutput, err := s.executeOnStakingSC([]byte("unStake@" + hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(registrationData.RewardAddress)))
		if err != nil {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do unStake for key %s: %s", hex.EncodeToString(blsKey), err.Error()))
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

		registrationData.NumRegistered -= 1
	}

	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func verifyBLSPublicKeys(registrationData *AuctionData, arguments [][]byte) error {
	for _, argKey := range arguments {
		found := false
		for _, blsKey := range registrationData.BlsPubKeys {
			if bytes.Equal(argKey, blsKey) {
				found = true
				break
			}
		}

		if !found {
			return vm.ErrBLSPublicKeyMissmatch
		}
	}

	return nil
}

func (s *stakingAuctionSC) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage("transaction value must be zero")
		return vmcommon.UserError
	}
	if len(args.Arguments) == 0 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	isStakeEnabled := s.isFunctionEnabled(s.enableStakingNonce)
	if !isStakeEnabled {
		s.eei.AddReturnMessage("unBond is not enabled")
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	blsKeys := args.Arguments
	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		s.eei.AddReturnMessage("could not get all blsKeys from registration data: error " + err.Error())
		return vmcommon.UserError
	}

	err = s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.UnBond * uint64(len(args.Arguments)))
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	unBondedKeys := make([][]byte, 0)
	totalUnBond := big.NewInt(0)
	totalSlashed := big.NewInt(0)
	for _, blsKey := range blsKeys {
		nodeData, err := s.getStakedData(blsKey)
		if err != nil {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do unBond for key: %s, error: %s", hex.EncodeToString(blsKey), err.Error()))
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
		}
		// returns what value is still under the selected bls key
		vmOutput, err := s.executeOnStakingSC([]byte("unBond@" + hex.EncodeToString(blsKey)))
		if err != nil || vmOutput.ReturnCode != vmcommon.Ok {
			s.eei.AddReturnMessage(fmt.Sprintf("cannot do unBond for key: %s", hex.EncodeToString(blsKey)))
			s.eei.Finish(blsKey)
			s.eei.Finish([]byte{failed})
			continue
		}

		auctionConfig := s.getConfig(nodeData.UnStakedEpoch)
		if nodeData.Waiting {
			registrationData.NumRegistered -= 1
		}

		unBondedKeys = append(unBondedKeys, blsKey)
		totalUnBond.Add(totalUnBond, auctionConfig.NodePrice)
		totalSlashed.Add(totalSlashed, nodeData.SlashValue)
	}

	totalUnBond.Sub(totalUnBond, totalSlashed)
	if totalUnBond.Cmp(zero) < 0 {
		totalUnBond.Set(zero)
	}
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

	err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBond, nil, 0)
	if err != nil {
		s.eei.AddReturnMessage("transfer error on unBond function")
		return vmcommon.UserError
	}

	if registrationData.LockedStake.Cmp(zero) == 0 && registrationData.TotalStakeValue.Cmp(zero) == 0 {
		s.eei.SetStorage(args.CallerAddr, nil)
	} else {
		s.deleteUnBondedKeys(registrationData, unBondedKeys)
		err := s.saveRegistrationData(args.CallerAddr, registrationData)
		if err != nil {
			s.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) deleteUnBondedKeys(registrationData *AuctionData, unBondedKeys [][]byte) {
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
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage("transaction value must be zero")
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
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	zero := big.NewInt(0)
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

func (s *stakingAuctionSC) calculateNodePrice(bids []AuctionData) (*big.Int, error) {
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

func (s *stakingAuctionSC) selection(bids []AuctionData) [][]byte {
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
		if _, ok := selectedKeys[expandedList[i]]; !ok {
			selected++
			selectedKeys[expandedList[i]] = struct{}{}
		}
	}

	for key := range selectedKeys {
		randomlySelected = append(randomlySelected, []byte(key))
	}

	return randomlySelected
}

func calcTotalQualifyingStake(nodePrice *big.Int, bids []AuctionData) *big.Int {
	totalQualifyingStake := big.NewInt(0)
	for _, validator := range bids {
		if validator.MaxStakePerNode.Cmp(nodePrice) < 0 {
			continue
		}

		maxPossibleNodes := big.NewInt(0).Div(validator.TotalStakeValue, nodePrice)
		if maxPossibleNodes.Uint64() > uint64(len(validator.BlsPubKeys)) {
			validatorQualifyingStake := big.NewInt(0).Mul(nodePrice, big.NewInt(int64(len(validator.BlsPubKeys))))
			totalQualifyingStake.Add(totalQualifyingStake, validatorQualifyingStake)
		} else {
			totalQualifyingStake.Add(totalQualifyingStake, validator.TotalStakeValue)
		}
	}

	return totalQualifyingStake
}

func calcNumQualifiedNodes(nodePrice *big.Int, bids []AuctionData) uint32 {
	numQualifiedNodes := uint32(0)
	for _, validator := range bids {
		if validator.MaxStakePerNode.Cmp(nodePrice) < 0 {
			continue
		}
		if validator.TotalStakeValue.Cmp(nodePrice) < 0 {
			continue
		}

		maxPossibleNodes := big.NewInt(0).Div(validator.TotalStakeValue, nodePrice)
		if maxPossibleNodes.Uint64() > uint64(len(validator.BlsPubKeys)) {
			numQualifiedNodes += uint32(len(validator.BlsPubKeys))
		} else {
			numQualifiedNodes += uint32(maxPossibleNodes.Uint64())
		}
	}

	return numQualifiedNodes
}

func calcNumAllocatedAndProportion(
	validator AuctionData,
	nodePrice *big.Int,
	totalQualifyingStake *big.Float,
) (uint64, float64) {
	maxPossibleNodes := big.NewInt(0).Div(validator.TotalStakeValue, nodePrice)
	validatorQualifyingStake := big.NewFloat(0).SetInt(validator.TotalStakeValue)
	qualifiedNodes := maxPossibleNodes.Uint64()

	if maxPossibleNodes.Uint64() > uint64(len(validator.BlsPubKeys)) {
		validatorQualifyingStake = big.NewFloat(0).SetInt(big.NewInt(0).Mul(nodePrice, big.NewInt(int64(len(validator.BlsPubKeys)))))
		qualifiedNodes = uint64(len(validator.BlsPubKeys))
	}

	proportionOfTotalStake := big.NewFloat(0).Quo(validatorQualifyingStake, totalQualifyingStake)
	proportion, _ := proportionOfTotalStake.Float64()
	allocatedNodes := float64(qualifiedNodes) * proportion
	numAllocatedNodes := uint64(allocatedNodes)

	return numAllocatedNodes, allocatedNodes
}

func shuffleList(list []string, random []byte) {
	randomSeed := big.NewInt(0).SetBytes(random[:8])
	r := rand.New(rand.NewSource(randomSeed.Int64()))

	for n := len(list); n > 0; n-- {
		randIndex := r.Intn(n)
		list[n-1], list[randIndex] = list[randIndex], list[n-1]
	}
}

func isNumArgsCorrectToStake(args [][]byte) bool {
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()
	areEnoughArgs := uint64(len(args)) >= 2*maxNodesToRun+1       // NumNodes + LIST(BLS_KEY+SignedMessage)
	areNotTooManyArgs := uint64(len(args)) <= 2*maxNodesToRun+1+2 // +2 are the optionals - reward address, maxStakePerNode
	return areEnoughArgs && areNotTooManyArgs
}

func (s *stakingAuctionSC) getTotalStaked(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas limit")
		return vmcommon.OutOfGas
	}

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registration data: error " + err.Error())
		return vmcommon.UserError
	}

	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("caller not registered in staking/auction sc")
		return vmcommon.UserError
	}

	s.eei.Finish([]byte(registrationData.TotalStakeValue.String()))
	return vmcommon.Ok
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (s *stakingAuctionSC) IsInterfaceNil() bool {
	return s == nil
}
