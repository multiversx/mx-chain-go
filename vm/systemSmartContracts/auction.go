package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"math/big"
	"math/rand"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// AuctionData represents what is saved for each validator / bid
type AuctionData struct {
	RewardAddress   []byte   `json:"RewardAddress"`
	RegisterNonce   uint64   `json:"RegisterNonce"`
	Epoch           uint32   `json:"Epoch"`
	BlsPubKeys      [][]byte `json:"BlsPubKeys"`
	TotalStakeValue *big.Int `json:"TotalStakeValue"`
	LockedStake     *big.Int `json:"LockedStake"`
	MaxStakePerNode *big.Int `json:"MaxStakePerNode"`
}

// StakedData represents the data which is saved for the selected nodes
type StakedData struct {
	RegisterNonce uint64   `json:"RegisterNonce"`
	Staked        bool     `json:"Staked"`
	UnStakedNonce uint64   `json:"UnStakedNonce"`
	UnStakedEpoch uint32   `json:"UnStakedEpoch"`
	RewardAddress []byte   `json:"RewardAddress"`
	StakeValue    *big.Int `json:"StakeValue"`
}

// AuctionConfig represents the settings for a specific epoch
type AuctionConfig struct {
	MinStakeValue *big.Int `json:"MinStakeValue"`
	NumNodes      uint32   `json:"NumNodes"`
	TotalSupply   *big.Int `json:"TotalSupply"`
	MinStep       *big.Int `json:"MinStep"`
	NodePrice     *big.Int `json:"NodePrice"`
}

type stakingAuctionSC struct {
	eei            vm.SystemEI
	unBondPeriod   uint64
	sigVerifier    vm.MessageSignVerifier
	baseConfig     AuctionConfig
	auctionEnabled bool
}

type ArgsStakingAuctionSmartContract struct {
	MinStakeValue  *big.Int
	MinStepValue   *big.Int
	TotalSupply    *big.Int
	UnBondPeriod   uint64
	NumNodes       uint32
	Eei            vm.SystemEI
	SigVerifier    vm.MessageSignVerifier
	AuctionEnabled bool
}

// NewStakingAuctionSmartContract creates an auction smart contract
func NewStakingAuctionSmartContract(
	args ArgsStakingAuctionSmartContract,
) (*stakingAuctionSC, error) {
	if args.MinStakeValue == nil {
		return nil, vm.ErrNilInitialStakeValue
	}
	if args.MinStakeValue.Cmp(big.NewInt(0)) < 1 {
		return nil, vm.ErrNegativeInitialStakeValue
	}
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	baseConfig := AuctionConfig{
		MinStakeValue: args.MinStakeValue,
		NumNodes:      args.NumNodes,
		TotalSupply:   args.TotalSupply,
		MinStep:       args.MinStepValue,
		NodePrice:     big.NewInt(0).Set(args.MinStakeValue),
	}

	reg := &stakingAuctionSC{
		eei:            args.Eei,
		unBondPeriod:   args.UnBondPeriod,
		sigVerifier:    args.SigVerifier,
		baseConfig:     baseConfig,
		auctionEnabled: args.AuctionEnabled,
	}
	return reg, nil
}

// Execute calls one of the functions from the staking smart contract and runs the code according to the input
func (s *stakingAuctionSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	switch args.Function {
	case "_init":
		return s.init(args)
	case "stake":
		return s.stake(args)
	case "unStake":
		return s.unStake(args)
	case "unBond":
		return s.unBond(args)
	case "claim":
		return s.claim(args)
	case "slash":
		return s.slash(args)
	case "get":
		return s.get(args)
	case "setConfig":
		return s.setConfig(args)
	}

	return vmcommon.UserError
}

func (s *stakingAuctionSC) get(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		return vmcommon.UserError
	}

	value := s.eei.GetStorage(args.Arguments[0])
	s.eei.Finish(value)

	return vmcommon.Ok
}

func (s *stakingAuctionSC) setConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := s.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		log.Debug("setConfig function was not called by the owner address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 6 {
		log.Debug("setConfig function called with wrong number of arguments")
		return vmcommon.UserError
	}

	config := AuctionConfig{
		MinStakeValue: big.NewInt(0).SetBytes(args.Arguments[0]),
		NumNodes:      uint32(big.NewInt(0).SetBytes(args.Arguments[1]).Uint64()),
		TotalSupply:   big.NewInt(0).SetBytes(args.Arguments[2]),
		MinStep:       big.NewInt(0).SetBytes(args.Arguments[3]),
		NodePrice:     big.NewInt(0).SetBytes(args.Arguments[4]),
	}

	configData, err := json.Marshal(config)
	if err != nil {
		log.Debug("setConfig marshall config error")
		return vmcommon.UserError
	}

	// argument 5 is equal with epoch
	s.eei.SetStorage(args.Arguments[5], configData)

	return vmcommon.Ok
}

func (s *stakingAuctionSC) checkConfigCorrectness(config AuctionConfig) error {
	if config.MinStakeValue == nil {
		return vm.ErrConfigIncorrect
	}
	if config.NodePrice == nil {
		return vm.ErrConfigIncorrect
	}
	if config.TotalSupply == nil {
		return vm.ErrConfigIncorrect
	}
	if config.MinStep == nil {
		return vm.ErrConfigIncorrect
	}
	return nil
}

func (s *stakingAuctionSC) getConfig(epoch uint32) AuctionConfig {
	epochKey := big.NewInt(int64(epoch)).Bytes()
	configData := s.eei.GetStorage(epochKey)
	if len(configData) == 0 {
		return s.baseConfig
	}

	config := AuctionConfig{}
	err := json.Unmarshal(configData, &config)
	if err != nil {
		log.Debug("unmarshal error on getConfig function",
			"error", err.Error(),
		)
		return s.baseConfig
	}

	if s.checkConfigCorrectness(config) != nil {
		baseConfigData, err := json.Marshal(s.baseConfig)
		if err != nil {
			log.Debug("marshal error on getConfig function")
		}
		s.eei.SetStorage(epochKey, baseConfigData)
		return s.baseConfig
	}

	return config
}

func (s *stakingAuctionSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := s.eei.GetStorage([]byte(ownerKey))
	if ownerAddress != nil {
		log.Error("smart contract was already initialized")
		return vmcommon.UserError
	}

	s.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	return vmcommon.Ok
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
		data := s.eei.GetStorage(newKey)
		if len(data) > 0 {
			return nil, vm.ErrKeyAlreadyRegistered
		}
	}

	return newKeys, nil
}

func (s *stakingAuctionSC) saveBLSKeys(registrationData *AuctionData, args [][]byte) error {
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()
	if uint64(len(args)) < maxNodesToRun+1 {
		log.Debug("not enough arguments to process stake function")
		return vm.ErrNotEnoughArgumentsToStake
	}

	blsKeys := s.getVerifiedBLSKeysFromArgs(args)

	newKeys, err := s.getNewValidKeys(registrationData.BlsPubKeys, blsKeys)
	if err != nil {
		log.Debug("staking with already existing key is not allowed")
		return vm.ErrKeyAlreadyRegistered
	}

	for _, blsKey := range newKeys {
		err = s.saveStakedData(blsKey, &StakedData{})
		if err != nil {
			return err
		}

		registrationData.BlsPubKeys = append(registrationData.BlsPubKeys, blsKey)
	}

	return nil
}

func (s *stakingAuctionSC) updateStakeValue(registrationData *AuctionData, caller []byte) vmcommon.ReturnCode {
	if len(registrationData.BlsPubKeys) == 0 {
		log.Debug("not enough arguments to process stake function")
		return vmcommon.UserError
	}

	err := s.saveRegistrationData(caller, registrationData)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) getVerifiedBLSKeysFromArgs(args [][]byte) [][]byte {
	blsKeys := make([][]byte, 0)
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()

	for i := uint64(1); i < maxNodesToRun*2+1; i += 2 {
		blsKey := args[i]
		signedMessage := args[i+1]
		err := s.sigVerifier.Verify([]byte("stake"), signedMessage, blsKey)
		if err != nil {
			continue
		}

		blsKeys = append(blsKeys, blsKey)
	}

	return blsKeys
}

func (s *stakingAuctionSC) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	config := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())

	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		return vmcommon.UserError
	}

	registrationData.TotalStakeValue.Add(registrationData.TotalStakeValue, args.CallValue)
	if registrationData.TotalStakeValue.Cmp(config.MinStakeValue) < 0 || args.CallValue.Sign() <= 0 {
		return vmcommon.UserError
	}

	lenArgs := len(args.Arguments)
	if lenArgs == 0 {
		return s.updateStakeValue(registrationData, args.CallerAddr)
	}

	if !isNumArgsCorrectToStake(args.Arguments) {
		log.Debug("incorrect number of arguments to call stake", "numArgs", args.Arguments)
		return vmcommon.UserError
	}

	maxNodesToRun := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	err = s.saveBLSKeys(registrationData, args.Arguments)
	if err != nil {
		return vmcommon.UserError
	}

	registrationData.RewardAddress = args.CallerAddr
	registrationData.MaxStakePerNode = big.NewInt(0).Set(registrationData.TotalStakeValue)
	registrationData.Epoch = s.eei.BlockChainHook().CurrentEpoch()

	// do the optionals - rewardAddress and maxStakePerNode
	if uint64(lenArgs) > maxNodesToRun*2+1 {
		for i := maxNodesToRun*2 + 1; i < uint64(lenArgs); i++ {
			if len(args.Arguments[i]) == len(args.CallerAddr) {
				registrationData.RewardAddress = args.Arguments[i]
			} else {
				registrationData.MaxStakePerNode.SetBytes(args.Arguments[i])
			}
		}
	}

	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		return vmcommon.UserError
	}

	if !s.auctionEnabled {
		numQualified := big.NewInt(0).Div(registrationData.TotalStakeValue, config.MinStakeValue)
		s.activateStakingFor(numQualified.Uint64(), registrationData.BlsPubKeys, config.MinStakeValue)
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) activateStakingFor(numQualified uint64, blsKeys [][]byte, fixedStakeValue *big.Int) {
	numStaked := uint64(0)
	for i := uint64(0); numStaked < numQualified && i < uint64(len(blsKeys)); i++ {
		stakedData, err := s.getStakedData(blsKeys[i])
		if err != nil {
			continue
		}

		numStaked++
		if stakedData.Staked == true {
			continue
		}

		stakedData.Staked = true
		stakedData.StakeValue.Set(fixedStakeValue)
	}
}

func (s *stakingAuctionSC) getOrCreateRegistrationData(key []byte) (*AuctionData, error) {
	data := s.eei.GetStorage(key)
	registrationData := AuctionData{
		RewardAddress:   nil,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0),
		MaxStakePerNode: big.NewInt(0),
	}

	if data != nil {
		err := json.Unmarshal(data, &registrationData)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}
	}

	return &registrationData, nil
}

func (s *stakingAuctionSC) saveRegistrationData(key []byte, auction *AuctionData) error {
	data, err := json.Marshal(*auction)
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
	data := s.eei.GetStorage(key)
	stakedData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: nil,
		StakeValue:    big.NewInt(0),
	}

	if data != nil {
		err := json.Unmarshal(data, &stakedData)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}
	}

	return &stakedData, nil
}

func (s *stakingAuctionSC) saveStakedData(key []byte, staked *StakedData) error {
	data, err := json.Marshal(*staked)
	if err != nil {
		log.Debug("marshal error on staking SC stake function ",
			"error", err.Error(),
		)
		return err
	}

	s.eei.SetStorage(key, data)
	return nil
}

func (s *stakingAuctionSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		return vmcommon.UserError
	}

	blsKeys, err := getBLSPublicKeys(registrationData, args)
	if err != nil {
		return vmcommon.UserError
	}

	for _, blsKey := range blsKeys {
		stakedData, err := s.getStakedData(blsKey)
		if err != nil {
			log.Debug("bls key was not staked")
			continue
		}

		if !stakedData.Staked {
			log.Debug("bls key was already unstaked")
			continue
		}

		stakedData.Staked = false
		stakedData.UnStakedNonce = s.eei.BlockChainHook().CurrentNonce()
		err = s.saveStakedData(blsKey, stakedData)
		if err != nil {
			log.Debug("error while saving staked data")
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func getBLSPublicKeys(registrationData *AuctionData, args *vmcommon.ContractCallInput) ([][]byte, error) {
	blsKeys := registrationData.BlsPubKeys
	if len(args.Arguments) > 0 {
		for _, argKey := range args.Arguments {
			found := false
			for _, blsKey := range blsKeys {
				if bytes.Equal(argKey, blsKey) {
					found = true
					break
				}
			}

			if !found {
				log.Debug("bls key for validator not found")
				return nil, vm.ErrBLSPublicKeyMissmatch
			}
		}

		blsKeys = args.Arguments
	}

	return blsKeys, nil
}

func (s *stakingAuctionSC) deleteBLSStakedData(blsKey []byte) (*big.Int, error) {
	stakedData, err := s.getStakedData(blsKey)
	if err != nil {
		log.Debug("bls key was not staked")
		return nil, vm.ErrBLSKeyIsNotStaked
	}

	if stakedData.Staked || stakedData.UnStakedNonce < stakedData.RegisterNonce {
		log.Debug("unBond is not possible for address which is staked or is not in unbond period")
		return nil, vm.ErrStillInUnBoundPeriod
	}

	currentNonce := s.eei.BlockChainHook().CurrentNonce()
	if currentNonce-stakedData.UnStakedNonce < s.unBondPeriod {
		log.Debug("unBond is not possible for address because unbond period did not pass")
		return nil, vm.ErrStillInUnBoundPeriod
	}

	s.eei.SetStorage(blsKey, nil)
	config := s.getConfig(stakedData.UnStakedEpoch)

	if stakedData.StakeValue.Cmp(config.NodePrice) < 0 {
		// validator was slashed, so he can unBond only the remaining stake
		return stakedData.StakeValue, nil
	}

	return config.NodePrice, nil
}

func (s *stakingAuctionSC) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		return vmcommon.UserError
	}

	blsKeys, err := getBLSPublicKeys(registrationData, args)
	if err != nil {
		return vmcommon.UserError
	}

	totalUnBound := big.NewInt(0)
	for _, blsKey := range blsKeys {
		// returns what value is still under the selected bls key
		unBondValue, err := s.deleteBLSStakedData(blsKey)
		if err != nil {
			continue
		}

		_ = totalUnBound.Add(totalUnBound, unBondValue)
	}

	if registrationData.LockedStake.Cmp(totalUnBound) < 0 {
		totalUnBound.Set(registrationData.LockedStake)
	}

	_ = registrationData.LockedStake.Sub(registrationData.LockedStake, totalUnBound)
	_ = registrationData.TotalStakeValue.Sub(registrationData.TotalStakeValue, totalUnBound)

	err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, totalUnBound, nil)
	if err != nil {
		log.Debug("transfer error on unBond function")
		return vmcommon.UserError
	}

	zero := big.NewInt(0)
	if registrationData.LockedStake.Cmp(zero) == 0 && registrationData.TotalStakeValue.Cmp(zero) == 0 {
		s.eei.SetStorage(args.CallerAddr, nil)
	} else {
		err := s.saveRegistrationData(args.CallerAddr, registrationData)
		if err != nil {
			log.Debug("cannot save registration data change")
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) claim(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, err := s.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		return vmcommon.UserError
	}

	if len(registrationData.RewardAddress) == 0 {
		return vmcommon.UserError
	}

	zero := big.NewInt(0)
	claimable := big.NewInt(0).Sub(registrationData.TotalStakeValue, registrationData.LockedStake)
	if claimable.Cmp(zero) <= 0 {
		return vmcommon.UserError
	}

	registrationData.TotalStakeValue.Set(registrationData.LockedStake)
	err = s.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		log.Debug("cannot save registration data change")
		return vmcommon.UserError
	}

	err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, claimable, nil)
	if err != nil {
		log.Debug("transfer error on finalizeUnStake function",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) slash(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := s.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		log.Debug("slash function not called by the owners address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 2 {
		log.Debug("slash function called with wrong number of arguments")
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingAuctionSC) calculateNodePrice(bids []AuctionData) (*big.Int, error) {
	config := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())

	minNodePrice := big.NewInt(0).Set(config.MinStakeValue)
	maxNodePrice := big.NewInt(0).Div(config.TotalSupply, big.NewInt(int64(config.NumNodes)))
	numNodes := config.NumNodes

	for nodePrice := maxNodePrice; nodePrice.Cmp(minNodePrice) >= 0; nodePrice.Sub(nodePrice, config.MinStep) {
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
	config := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())
	stillNeeded := uint32(0)
	if config.NumNodes > alreadySelected {
		stillNeeded = config.NumNodes - alreadySelected
	}

	randomlySelected := s.selectRandomly(toBeSelectedRandom, stillNeeded)
	alreadySelected += uint32(len(randomlySelected))

	if config.NumNodes > alreadySelected {
		stillNeeded = config.NumNodes - alreadySelected
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

// IsInterfaceNil verifies if the underlying object is nil or not
func (s *stakingAuctionSC) IsInterfaceNil() bool {
	return s == nil
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
	if uint64(len(args)) < 2*maxNodesToRun+1 {
		return false
	}

	return true
}
