package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"math/big"
	"math/rand"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// AuctionData represents what is saved for each validator / bid
type AuctionData struct {
	RewardAddress   []byte   `json:"RewardAddress"`
	RegisterNonce   uint64   `json:"RegisterNonce"`
	Epoch           uint32   `json:"Epoch"`
	BlsPubKeys      [][]byte `json:"BlsPubKeys"`
	TotalStakeValue *big.Int `json:"StakeValue"`
	BlockedStake    *big.Int `json:"BlockedStake"`
	MaxStakePerNode *big.Int `json:"MaxStakePerNode"`
}

// StakedData represents the data which is saved for the selected nodes
type StakedData struct {
	RegisterNonce uint64 `json:"RegisterNonce"`
	Staked        bool   `json:"Staked"`
	UnStakedNonce uint64 `json:"UnStakedNonce"`
	UnStakedEpoch uint32 `json:"UnStakedEpoch"`
	RewardAddress []byte `json:"RewardAddress"`
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
	eei           vm.SystemEI
	unBoundPeriod uint64
	kg            crypto.KeyGenerator
	baseConfig    AuctionConfig
}

// NewStakingAuctionSmartContract creates an auction smart contract
func NewStakingAuctionSmartContract(
	minStakeValue *big.Int,
	minStepValue *big.Int,
	totalSupply *big.Int,
	unBoundPeriod uint64,
	numNodes uint32,
	eei vm.SystemEI,
	kg crypto.KeyGenerator,
) (*stakingAuctionSC, error) {
	if minStakeValue == nil {
		return nil, vm.ErrNilInitialStakeValue
	}
	if minStakeValue.Cmp(big.NewInt(0)) < 1 {
		return nil, vm.ErrNegativeInitialStakeValue
	}
	if check.IfNil(eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	baseConfig := AuctionConfig{
		MinStakeValue: minStakeValue,
		NumNodes:      numNodes,
		TotalSupply:   totalSupply,
		MinStep:       minStepValue,
		NodePrice:     big.NewInt(0).Set(minStakeValue),
	}

	reg := &stakingAuctionSC{
		eei:           eei,
		unBoundPeriod: unBoundPeriod,
		kg:            kg,
		baseConfig:    baseConfig,
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
	case "unBound":
		return s.unBound(args)
	case "claim":
		return s.claim(args)
	case "slash":
		return s.slash(args)
	case "get":
		return s.get(args)
	case "setConfig":
		return s.setConfig(args)
	}

	//TODO: integrated into the protocol the calling the functions select and setConfig at end-of-epoch

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

func (s *stakingAuctionSC) getConfig(epoch uint32) AuctionConfig {
	epochKey := big.NewInt(int64(epoch)).Bytes()
	configData := s.eei.GetStorage(epochKey)
	if len(configData) == 0 {
		return s.baseConfig
	}

	config := AuctionConfig{}
	err := json.Unmarshal(configData, &config)
	if err != nil {
		log.Debug("unmarshal error on staking SC stake function",
			"error", err.Error(),
		)
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

func (s *stakingAuctionSC) verifyIfKeysExist(registeredKeys [][]byte, arguments [][]byte, maxNumNodes uint64) ([][]byte, error) {
	registeredKeysMap := make(map[string]struct{})

	for _, blsKey := range registeredKeys {
		registeredKeysMap[string(blsKey)] = struct{}{}
	}

	newKeys := make([][]byte, 0)
	keysFromArgument := arguments[1:]
	for i := uint64(0); i < uint64(len(keysFromArgument)) && i < maxNumNodes+1; i++ {
		if _, ok := registeredKeysMap[string(arguments[i])]; ok {
			continue
		}

		newKeys = append(newKeys, arguments[i])
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

	newKeys, err := s.verifyIfKeysExist(registrationData.BlsPubKeys, args, maxNodesToRun)
	if err != nil {
		log.Debug("staking with already existing key is not allowed")
		return vm.ErrKeyAlreadyRegistered
	}

	for _, blsKey := range newKeys {
		_, err := s.kg.PublicKeyFromByteArray(blsKey)
		if err != nil {
			log.Debug("bls key is not valid")
			return vm.ErrBLSKeyIsNotValid
		}

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

func (s *stakingAuctionSC) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	config := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())

	registrationData, err := s.getRegistrationData(args.CallerAddr)
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

	maxNodesToRun := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	err = s.saveBLSKeys(registrationData, args.Arguments)
	if err != nil {
		return vmcommon.UserError
	}

	registrationData.RewardAddress = args.CallerAddr
	registrationData.MaxStakePerNode = big.NewInt(0).Set(registrationData.TotalStakeValue)
	registrationData.Epoch = s.eei.BlockChainHook().CurrentEpoch()

	// do the optionals - rewardAddress and maxStakePerNode
	if uint64(lenArgs) > maxNodesToRun+1 {
		for i := maxNodesToRun + 1; i < uint64(lenArgs); i++ {
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

	return vmcommon.Ok
}

func (s *stakingAuctionSC) getRegistrationData(key []byte) (*AuctionData, error) {
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
	registrationData, err := s.getRegistrationData(args.CallerAddr)
	if err != nil {
		return vmcommon.UserError
	}

	blsKeys, err := getBLSPublicKeys(registrationData, args)
	if err != nil {
		return vmcommon.UserError
	}

	for _, blsKey := range blsKeys {
		stakedData, err := s.getStakedData(blsKey)
		if err != nil || len(stakedData.RewardAddress) == 0 {
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
	if err != nil || len(stakedData.RewardAddress) == 0 {
		log.Debug("bls key was not staked")
		return nil, vm.ErrBLSKeyIsNotStaked
	}

	if stakedData.Staked || stakedData.UnStakedNonce < stakedData.RegisterNonce {
		log.Debug("unBound is not possible for address which is staked or is not in unbound period")
		return nil, vm.ErrStillInUnBoundPeriod
	}

	currentNonce := s.eei.BlockChainHook().CurrentNonce()
	if currentNonce-stakedData.UnStakedNonce < s.unBoundPeriod {
		log.Debug("unBound is not possible for address because unbound period did not pass")
		return nil, vm.ErrStillInUnBoundPeriod
	}

	s.eei.SetStorage(blsKey, nil)
	config := s.getConfig(stakedData.UnStakedEpoch)

	return config.NodePrice, nil
}

func (s *stakingAuctionSC) unBound(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, err := s.getRegistrationData(args.CallerAddr)
	if err != nil {
		return vmcommon.UserError
	}

	blsKeys, err := getBLSPublicKeys(registrationData, args)
	if err != nil {
		return vmcommon.UserError
	}

	totalUnBound := big.NewInt(0)
	for _, blsKey := range blsKeys {
		nodePrice, err := s.deleteBLSStakedData(blsKey)
		if err != nil {
			continue
		}

		err = s.eei.Transfer(args.CallerAddr, args.RecipientAddr, nodePrice, nil)
		if err != nil {
			log.Debug("transfer error on unBound function",
				"error", err.Error(),
			)
			return vmcommon.UserError
		}

		_ = totalUnBound.Add(totalUnBound, nodePrice)
	}

	if registrationData.BlockedStake.Cmp(totalUnBound) < 0 {
		log.Debug("too much to unbound, not enough total stake")
		return vmcommon.UserError
	}

	_ = registrationData.BlockedStake.Sub(registrationData.BlockedStake, totalUnBound)
	_ = registrationData.TotalStakeValue.Sub(registrationData.TotalStakeValue, totalUnBound)

	zero := big.NewInt(0)
	if registrationData.BlockedStake.Cmp(zero) == 0 && registrationData.TotalStakeValue.Cmp(zero) == 0 {
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
	registrationData, err := s.getRegistrationData(args.CallerAddr)
	if err != nil {
		return vmcommon.UserError
	}

	if len(registrationData.RewardAddress) == 0 {
		return vmcommon.UserError
	}

	zero := big.NewInt(0)
	claimable := big.NewInt(0).Sub(registrationData.TotalStakeValue, registrationData.BlockedStake)
	if claimable.Cmp(zero) <= 0 {
		return vmcommon.UserError
	}

	registrationData.TotalStakeValue.Set(registrationData.BlockedStake)
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
