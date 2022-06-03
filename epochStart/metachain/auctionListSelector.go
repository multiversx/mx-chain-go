package metachain

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

type ownerAuctionData struct {
	numActiveNodes           int64
	numAuctionNodes          int64
	numQualifiedAuctionNodes int64
	numStakedNodes           int64
	totalTopUp               *big.Int
	topUpPerNode             *big.Int
	qualifiedTopUpPerNode    *big.Int
	auctionList              []state.ValidatorInfoHandler
}

type auctionConfig struct {
	step        *big.Int
	minTopUp    *big.Int
	maxTopUp    *big.Int
	denominator *big.Int
}

type auctionListSelector struct {
	shardCoordinator    sharding.Coordinator
	stakingDataProvider epochStart.StakingDataProvider
	nodesConfigProvider epochStart.MaxNodesChangeConfigProvider
	softAuctionConfig   *auctionConfig
}

// AuctionListSelectorArgs is a struct placeholder for all arguments required to create an auctionListSelector
type AuctionListSelectorArgs struct {
	ShardCoordinator             sharding.Coordinator
	StakingDataProvider          epochStart.StakingDataProvider
	MaxNodesChangeConfigProvider epochStart.MaxNodesChangeConfigProvider
	SoftAuctionConfig            config.SoftAuctionConfig
	Denomination                 int
}

// NewAuctionListSelector will create a new auctionListSelector, which handles selection of nodes from auction list based
// on their top up
func NewAuctionListSelector(args AuctionListSelectorArgs) (*auctionListSelector, error) {
	softAuctionConfig, err := getAuctionConfig(args.SoftAuctionConfig, args.Denomination)
	if err != nil {
		return nil, err
	}
	err = checkNilArgs(args)
	if err != nil {
		return nil, err
	}

	log.Debug("NewAuctionListSelector with config",
		"top up step", softAuctionConfig.step.String(),
		"min top up", softAuctionConfig.minTopUp.String(),
		"max top up", softAuctionConfig.maxTopUp.String(),
		"denomination", args.Denomination,
		"denominator for pretty values", softAuctionConfig.denominator.String(),
	)

	return &auctionListSelector{
		shardCoordinator:    args.ShardCoordinator,
		stakingDataProvider: args.StakingDataProvider,
		nodesConfigProvider: args.MaxNodesChangeConfigProvider,
		softAuctionConfig:   softAuctionConfig,
	}, nil
}

func getAuctionConfig(softAuctionConfig config.SoftAuctionConfig, denomination int) (*auctionConfig, error) {
	step, ok := big.NewInt(0).SetString(softAuctionConfig.TopUpStep, 10)
	if !ok || step.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w for step in soft auction config;expected number > 0, got %s",
			process.ErrInvalidValue,
			softAuctionConfig.TopUpStep,
		)
	}

	minTopUp, ok := big.NewInt(0).SetString(softAuctionConfig.MinTopUp, 10)
	if !ok || minTopUp.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w for min top up in soft auction config;expected number > 0, got %s",
			process.ErrInvalidValue,
			softAuctionConfig.MinTopUp,
		)
	}

	maxTopUp, ok := big.NewInt(0).SetString(softAuctionConfig.MaxTopUp, 10)
	if !ok || maxTopUp.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w for max top up in soft auction config;expected number > 0, got %s",
			process.ErrInvalidValue,
			softAuctionConfig.MaxTopUp,
		)
	}

	if denomination < 0 {
		return nil, fmt.Errorf("%w for denomination soft auction config;expected number >= 0, got %d",
			process.ErrInvalidValue,
			denomination,
		)
	}

	return &auctionConfig{
		step:        step,
		minTopUp:    minTopUp,
		maxTopUp:    maxTopUp,
		denominator: big.NewInt(int64(math.Pow10(denomination))),
	}, nil
}

func checkNilArgs(args AuctionListSelectorArgs) error {
	if check.IfNil(args.ShardCoordinator) {
		return epochStart.ErrNilShardCoordinator
	}
	if check.IfNil(args.StakingDataProvider) {
		return epochStart.ErrNilStakingDataProvider
	}
	if check.IfNil(args.MaxNodesChangeConfigProvider) {
		return epochStart.ErrNilMaxNodesChangeConfigProvider
	}

	return nil
}

// SelectNodesFromAuctionList will select nodes from validatorsInfoMap based on their top up. If two or more validators
// have the same top-up, then sorting will be done based on blsKey XOR randomness. Selected nodes will have their list set
// to common.SelectNodesFromAuctionList
func (als *auctionListSelector) SelectNodesFromAuctionList(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	randomness []byte,
) error {
	if len(randomness) == 0 {
		return process.ErrNilRandSeed
	}

	ownersData, auctionListSize, err := als.getAuctionData()
	currNumOfValidators := als.stakingDataProvider.GetNumOfValidatorsInCurrentEpoch()
	if err != nil {
		return err
	}
	if auctionListSize == 0 {
		log.Info("auctionListSelector.SelectNodesFromAuctionList: empty auction list; skip selection")
		return nil
	}

	currNodesConfig := als.nodesConfigProvider.GetCurrentNodesConfig()
	numOfShuffledNodes := currNodesConfig.NodesToShufflePerShard * (als.shardCoordinator.NumberOfShards() + 1)
	numOfValidatorsAfterShuffling, err := safeSub(currNumOfValidators, numOfShuffledNodes)
	if err != nil {
		log.Warn(fmt.Sprintf("%v when trying to compute numOfValidatorsAfterShuffling = %v - %v (currNumOfValidators - numOfShuffledNodes)",
			err,
			currNumOfValidators,
			numOfShuffledNodes,
		))
		numOfValidatorsAfterShuffling = 0
	}

	maxNumNodes := currNodesConfig.MaxNumNodes
	availableSlots, err := safeSub(maxNumNodes, numOfValidatorsAfterShuffling)
	if availableSlots == 0 || err != nil {
		log.Info(fmt.Sprintf("%v or zero value when trying to compute availableSlots = %v - %v (maxNodes - numOfValidatorsAfterShuffling); skip selecting nodes from auction list",
			err,
			maxNumNodes,
			numOfValidatorsAfterShuffling,
		))
		return nil
	}

	log.Info("auctionListSelector.SelectNodesFromAuctionList",
		"max nodes", maxNumNodes,
		"current number of validators", currNumOfValidators,
		"num of nodes which will be shuffled out", numOfShuffledNodes,
		"num of validators after shuffling", numOfValidatorsAfterShuffling,
		"auction list size", auctionListSize,
		fmt.Sprintf("available slots (%v - %v)", maxNumNodes, numOfValidatorsAfterShuffling), availableSlots,
	)

	als.displayOwnersData(ownersData)
	numOfAvailableNodeSlots := core.MinUint32(auctionListSize, availableSlots)

	sw := core.NewStopWatch()
	sw.Start("auctionListSelector.sortAuctionList")
	defer func() {
		sw.Stop("auctionListSelector.sortAuctionList")
		log.Debug("time measurements", sw.GetMeasurements()...)
	}()

	return als.sortAuctionList(ownersData, numOfAvailableNodeSlots, validatorsInfoMap, randomness)
}

func (als *auctionListSelector) getAuctionData() (map[string]*ownerAuctionData, uint32, error) {
	ownersData := make(map[string]*ownerAuctionData)
	numOfNodesInAuction := uint32(0)

	for owner, ownerData := range als.stakingDataProvider.GetOwnersStats() {
		if ownerData.Qualified && ownerData.NumAuctionNodes > 0 {
			ownersData[owner] = &ownerAuctionData{
				numActiveNodes:           ownerData.NumActiveNodes,
				numAuctionNodes:          ownerData.NumAuctionNodes,
				numQualifiedAuctionNodes: ownerData.NumAuctionNodes,
				numStakedNodes:           ownerData.NumStakedNodes,
				totalTopUp:               ownerData.TotalTopUp,
				topUpPerNode:             ownerData.TopUpPerNode,
				qualifiedTopUpPerNode:    ownerData.TopUpPerNode,
				auctionList:              make([]state.ValidatorInfoHandler, len(ownerData.AuctionList)),
			}
			copy(ownersData[owner].auctionList, ownerData.AuctionList)
			numOfNodesInAuction += uint32(ownerData.NumAuctionNodes)
		}
	}

	return ownersData, numOfNodesInAuction, nil
}

func isInAuction(validator state.ValidatorInfoHandler) bool {
	return validator.GetList() == string(common.AuctionList)
}

func (als *auctionListSelector) addOwnerData(
	owner string,
	validator state.ValidatorInfoHandler,
	ownersData map[string]*ownerAuctionData,
) error {
	ownerPubKey := []byte(owner)
	validatorPubKey := validator.GetPublicKey()
	stakedNodes, err := als.stakingDataProvider.GetNumStakedNodes(ownerPubKey)
	if err != nil {
		return fmt.Errorf("auctionListSelector.addOwnerData: error getting num staked nodes: %w, owner: %s, node: %s",
			err,
			hex.EncodeToString(ownerPubKey),
			hex.EncodeToString(validatorPubKey),
		)
	}
	if stakedNodes == 0 {
		return fmt.Errorf("auctionListSelector.addOwnerData error: %w, owner: %s, node: %s",
			epochStart.ErrOwnerHasNoStakedNode,
			hex.EncodeToString(ownerPubKey),
			hex.EncodeToString(validatorPubKey),
		)
	}

	totalTopUp, err := als.stakingDataProvider.GetTotalTopUp(ownerPubKey)
	if err != nil {
		return fmt.Errorf("auctionListSelector.addOwnerData: error getting total top up: %w, owner: %s, node: %s",
			err,
			hex.EncodeToString(ownerPubKey),
			hex.EncodeToString(validatorPubKey),
		)
	}

	data, exists := ownersData[owner]
	if exists {
		data.numAuctionNodes++
		data.numQualifiedAuctionNodes++
		data.numActiveNodes--
		data.auctionList = append(data.auctionList, validator)
	} else {
		stakedNodesBigInt := big.NewInt(stakedNodes)
		topUpPerNode := big.NewInt(0).Div(totalTopUp, stakedNodesBigInt)
		ownersData[owner] = &ownerAuctionData{
			numAuctionNodes:          1,
			numQualifiedAuctionNodes: 1,
			numActiveNodes:           stakedNodes - 1,
			numStakedNodes:           stakedNodes,
			totalTopUp:               big.NewInt(0).SetBytes(totalTopUp.Bytes()),
			topUpPerNode:             topUpPerNode,
			qualifiedTopUpPerNode:    topUpPerNode,
			auctionList:              []state.ValidatorInfoHandler{validator},
		}
	}

	return nil
}

// TODO: Move this in elrond-go-core
func safeSub(a, b uint32) (uint32, error) {
	if a < b {
		return 0, epochStart.ErrUint32SubtractionOverflow
	}
	return a - b, nil
}

func (als *auctionListSelector) sortAuctionList(
	ownersData map[string]*ownerAuctionData,
	numOfAvailableNodeSlots uint32,
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	randomness []byte,
) error {
	softAuctionNodesConfig := als.calcSoftAuctionNodesConfig(ownersData, numOfAvailableNodeSlots)
	selectedNodes := als.selectNodes(softAuctionNodesConfig, numOfAvailableNodeSlots, randomness)
	return markAuctionNodesAsSelected(selectedNodes, validatorsInfoMap)
}

func (als *auctionListSelector) calcSoftAuctionNodesConfig(
	data map[string]*ownerAuctionData,
	numAvailableSlots uint32,
) map[string]*ownerAuctionData {
	ownersData := copyOwnersData(data)
	minTopUp, maxTopUp := als.getMinMaxPossibleTopUp(ownersData)
	log.Debug("auctionListSelector: calc min and max possible top up",
		"min top up per node", minTopUp.String(),
		"max top up per node", maxTopUp.String(),
	)

	topUp := big.NewInt(0).SetBytes(minTopUp.Bytes())
	previousConfig := copyOwnersData(ownersData)
	for ; topUp.Cmp(maxTopUp) < 0; topUp.Add(topUp, als.softAuctionConfig.step) {
		numNodesQualifyingForTopUp := int64(0)
		previousConfig = copyOwnersData(ownersData)

		for ownerPubKey, owner := range ownersData {
			activeNodes := big.NewInt(owner.numActiveNodes)
			topUpActiveNodes := big.NewInt(0).Mul(topUp, activeNodes)
			validatorTopUpForAuction := big.NewInt(0).Sub(owner.totalTopUp, topUpActiveNodes)
			if validatorTopUpForAuction.Cmp(topUp) < 0 {
				delete(ownersData, ownerPubKey)
				continue
			}

			qualifiedNodes := big.NewInt(0).Div(validatorTopUpForAuction, topUp).Int64()
			if qualifiedNodes > owner.numAuctionNodes {
				numNodesQualifyingForTopUp += owner.numAuctionNodes
			} else {
				numNodesQualifyingForTopUp += qualifiedNodes
				owner.numQualifiedAuctionNodes = qualifiedNodes

				ownerRemainingNodes := big.NewInt(owner.numActiveNodes + owner.numQualifiedAuctionNodes)
				owner.qualifiedTopUpPerNode = big.NewInt(0).Div(owner.totalTopUp, ownerRemainingNodes)
			}
		}

		if numNodesQualifyingForTopUp < int64(numAvailableSlots) {
			break
		}
	}

	als.displayMinRequiredTopUp(topUp, minTopUp)
	return previousConfig
}

func (als *auctionListSelector) getMinMaxPossibleTopUp(ownersData map[string]*ownerAuctionData) (*big.Int, *big.Int) {
	min := big.NewInt(0).SetBytes(als.softAuctionConfig.maxTopUp.Bytes())
	max := big.NewInt(0).SetBytes(als.softAuctionConfig.minTopUp.Bytes())

	for _, owner := range ownersData {
		if owner.topUpPerNode.Cmp(min) < 0 {
			min = big.NewInt(0).SetBytes(owner.topUpPerNode.Bytes())
		}

		ownerNumNodesWithOnlyOneAuctionNode := big.NewInt(owner.numActiveNodes + 1)
		maxPossibleTopUpForOwner := big.NewInt(0).Div(owner.totalTopUp, ownerNumNodesWithOnlyOneAuctionNode)
		if maxPossibleTopUpForOwner.Cmp(max) > 0 {
			max = big.NewInt(0).SetBytes(maxPossibleTopUpForOwner.Bytes())
		}
	}

	if min.Cmp(als.softAuctionConfig.minTopUp) < 0 {
		min = als.softAuctionConfig.minTopUp
	}

	return min, max
}

func copyOwnersData(ownersData map[string]*ownerAuctionData) map[string]*ownerAuctionData {
	ret := make(map[string]*ownerAuctionData)
	for owner, data := range ownersData {
		ret[owner] = &ownerAuctionData{
			numActiveNodes:           data.numActiveNodes,
			numAuctionNodes:          data.numAuctionNodes,
			numQualifiedAuctionNodes: data.numQualifiedAuctionNodes,
			numStakedNodes:           data.numStakedNodes,
			totalTopUp:               data.totalTopUp,
			topUpPerNode:             data.topUpPerNode,
			qualifiedTopUpPerNode:    data.qualifiedTopUpPerNode,
			auctionList:              make([]state.ValidatorInfoHandler, len(data.auctionList)),
		}
		copy(ret[owner].auctionList, data.auctionList)
	}

	return ret
}

func markAuctionNodesAsSelected(
	selectedNodes []state.ValidatorInfoHandler,
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
) error {
	for _, node := range selectedNodes {
		newNode := node
		newNode.SetList(string(common.SelectedFromAuctionList))

		err := validatorsInfoMap.Replace(node, newNode)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (als *auctionListSelector) IsInterfaceNil() bool {
	return als == nil
}
