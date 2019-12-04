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
	StartNonce      uint64   `json:"StartNonce"`
	Epoch           uint32   `json:"Epoch"`
	BlsPubKeys      [][]byte `json:"BlsPubKeys"`
	TotalStakeValue *big.Int `json:"StakeValue"`
	BlockedStake    *big.Int `json:"BlockedStake"`
	MaxStakePerNode *big.Int `json:"MaxStakePerNode"`
}

// StakedData represents the data which is saved for the selected nodes
type StakedData struct {
	StartNonce    uint64 `json:"StartNonce"`
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
		log.Debug("setConfig function called by not the owners address")
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
		if len(registrationData.BlsPubKeys) > 0 {
			err := s.saveRegistrationData(args.CallerAddr, registrationData)
			if err != nil {
				return vmcommon.UserError
			}

			return vmcommon.Ok
		}

		return vmcommon.UserError
	}

	if lenArgs < 1 {
		log.Debug("not enough arguments to process stake function")
		return vmcommon.UserError
	}

	maxNodesToRun := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	if uint64(lenArgs) < maxNodesToRun+1 {
		log.Debug("not enough arguments to process stake function")
		return vmcommon.UserError
	}

	for i := uint64(1); i < maxNodesToRun+1; i++ {
		_, err := s.kg.PublicKeyFromByteArray(args.Arguments[i])
		if err != nil {
			log.Debug("bls key is not valid")
			return vmcommon.UserError
		}

		registrationData.BlsPubKeys = append(registrationData.BlsPubKeys, args.Arguments[i])
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
		StartNonce:      0,
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
		StartNonce:    0,
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
			return vmcommon.UserError
		}

		if !stakedData.Staked {
			log.Debug("bls key was already unstaked")
			return vmcommon.UserError
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
				log.Debug("not allowed to unbound for bls key which is not under validator")
				return nil, vm.ErrBLSPublicKeyMissmatch
			}
		}

		blsKeys = args.Arguments
	}

	return blsKeys, nil
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
		stakedData, err := s.getStakedData(blsKey)
		if err != nil || len(stakedData.RewardAddress) == 0 {
			log.Debug("bls key was not staked")
			return vmcommon.UserError
		}

		if stakedData.Staked || stakedData.UnStakedNonce <= stakedData.StartNonce {
			log.Debug("unBound is not possible for address which is staked or is not in unbound period")
			return vmcommon.UserError
		}

		currentNonce := s.eei.BlockChainHook().CurrentNonce()
		if currentNonce-stakedData.UnStakedNonce < s.unBoundPeriod {
			log.Debug("unBound is not possible for address because unbound period did not pass")
			return vmcommon.UserError
		}

		s.eei.SetStorage(blsKey, nil)
		config := s.getConfig(stakedData.UnStakedEpoch)

		ownerAddress := s.eei.GetStorage([]byte(ownerKey))
		err = s.eei.Transfer(args.CallerAddr, ownerAddress, config.NodePrice, nil)
		if err != nil {
			log.Debug("transfer error on finalizeUnStake function",
				"error", err.Error(),
			)
			return vmcommon.UserError
		}

		_ = totalUnBound.Add(totalUnBound, config.NodePrice)
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

	ownerAddress := s.eei.GetStorage([]byte(ownerKey))
	err = s.eei.Transfer(args.CallerAddr, ownerAddress, claimable, nil)
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
		log.Debug("slash function called by not the owners address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 2 {
		log.Debug("slash function called by wrong number of arguments")
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
		qualifiedNodes := s.calcNumQualifiedNodes(nodePrice, bids)
		if qualifiedNodes >= numNodes {
			return nodePrice, nil
		}
	}

	return nil, vm.ErrNotEnoughQualifiedNodes
}

func (s *stakingAuctionSC) calcNumQualifiedNodes(nodePrice *big.Int, bids []AuctionData) uint32 {
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

func (s *stakingAuctionSC) selection(bids []AuctionData) [][]byte {
	nodePrice, err := s.calculateNodePrice(bids)
	if err != nil {
		return nil
	}

	totalQualifyingStake := big.NewFloat(0).SetInt(calcTotalQualifyingStake(nodePrice, bids))

	toBeSelectedRandom := make(map[string]float64)
	finalSelectedNodes := make([][]byte, 0)
	for _, validator := range bids {
		if validator.MaxStakePerNode.Cmp(nodePrice) < 0 {
			continue
		}

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
		if allocatedNodes-float64(numAllocatedNodes) > 0.99 {
			numAllocatedNodes += 1
		} else {
			selectorProp := allocatedNodes - float64(numAllocatedNodes)
			toBeSelectedRandom[string(validator.BlsPubKeys[numAllocatedNodes])] = selectorProp
		}

		if numAllocatedNodes > 0 {
			finalSelectedNodes = append(finalSelectedNodes, validator.BlsPubKeys[:numAllocatedNodes]...)
		}
	}

	config := s.getConfig(s.eei.BlockChainHook().CurrentEpoch())
	stillNeeded := uint32(0)
	if config.NumNodes > uint32(len(finalSelectedNodes)) {
		stillNeeded = config.NumNodes - uint32(len(finalSelectedNodes))
	}

	randomlySelected := s.selectRandomly(toBeSelectedRandom, stillNeeded)
	finalSelectedNodes = append(finalSelectedNodes, randomlySelected...)

	return finalSelectedNodes
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
	randomSeed := big.NewInt(0).SetBytes(random[:8])
	rand.Seed(randomSeed.Int64())

	rand.Shuffle(len(expandedList), func(i, j int) { expandedList[i], expandedList[j] = expandedList[j], expandedList[i] })

	selectedKeys := make(map[string]struct{})
	selected := uint32(0)
	for i := 0; selected < numNeeded || i < len(expandedList); i++ {
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

// IsInterfaceNil verifies if the underlying object is nil or not
func (s *stakingAuctionSC) IsInterfaceNil() bool {
	return s == nil
}
