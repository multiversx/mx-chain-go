package metachain

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

// OwnerAuctionData holds necessary auction data for an owner
type OwnerAuctionData struct {
	numStakedNodes           int64
	numActiveNodes           int64
	numAuctionNodes          int64
	numQualifiedAuctionNodes int64
	totalTopUp               *big.Int
	topUpPerNode             *big.Int
	qualifiedTopUpPerNode    *big.Int
	auctionList              []state.ValidatorInfoHandler
}

type auctionConfig struct {
	step                  *big.Int
	minTopUp              *big.Int
	maxTopUp              *big.Int
	denominator           *big.Int
	maxNumberOfIterations uint64
}

type auctionListSelector struct {
	shardCoordinator     sharding.Coordinator
	stakingDataProvider  epochStart.StakingDataProvider
	nodesConfigProvider  epochStart.MaxNodesChangeConfigProvider
	auctionListDisplayer AuctionListDisplayHandler
	softAuctionConfig    *auctionConfig
}

// AuctionListSelectorArgs is a struct placeholder for all arguments required to create an auctionListSelector
type AuctionListSelectorArgs struct {
	ShardCoordinator             sharding.Coordinator
	StakingDataProvider          epochStart.StakingDataProvider
	MaxNodesChangeConfigProvider epochStart.MaxNodesChangeConfigProvider
	AuctionListDisplayHandler    AuctionListDisplayHandler
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
		shardCoordinator:     args.ShardCoordinator,
		stakingDataProvider:  args.StakingDataProvider,
		nodesConfigProvider:  args.MaxNodesChangeConfigProvider,
		auctionListDisplayer: args.AuctionListDisplayHandler,
		softAuctionConfig:    softAuctionConfig,
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

	if minTopUp.Cmp(maxTopUp) > 0 {
		return nil, fmt.Errorf("%w for min/max top up in soft auction config; min value: %s > max value: %s",
			process.ErrInvalidValue,
			softAuctionConfig.MinTopUp,
			softAuctionConfig.MaxTopUp,
		)
	}

	if denomination < 0 {
		return nil, fmt.Errorf("%w for denomination in soft auction config;expected number >= 0, got %d",
			process.ErrInvalidValue,
			denomination,
		)
	}

	if softAuctionConfig.MaxNumberOfIterations == 0 {
		return nil, fmt.Errorf("%w for max number of iterations in soft auction config;expected value > 0",
			process.ErrInvalidValue,
		)
	}

	denominationStr := "1" + strings.Repeat("0", denomination)
	denominator, ok := big.NewInt(0).SetString(denominationStr, 10)
	if !ok {
		return nil, fmt.Errorf("%w for denomination: %d",
			errCannotComputeDenominator,
			denomination,
		)
	}

	if minTopUp.Cmp(denominator) < 0 {
		return nil, fmt.Errorf("%w for min top up in auction config; expected value to be >= %s, got %s",
			process.ErrInvalidValue,
			denominator.String(),
			minTopUp.String(),
		)
	}

	if step.Cmp(denominator) < 0 {
		return nil, fmt.Errorf("%w for step in auction config; expected value to be >= %s, got %s",
			process.ErrInvalidValue,
			denominator.String(),
			step.String(),
		)
	}

	return &auctionConfig{
		step:                  step,
		minTopUp:              minTopUp,
		maxTopUp:              maxTopUp,
		denominator:           denominator,
		maxNumberOfIterations: softAuctionConfig.MaxNumberOfIterations,
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
	if check.IfNil(args.AuctionListDisplayHandler) {
		return errNilAuctionListDisplayHandler
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

	ownersData, auctionListSize := als.getAuctionData()
	if auctionListSize == 0 {
		log.Info("auctionListSelector.SelectNodesFromAuctionList: empty auction list; skip selection")
		return nil
	}

	currNodesConfig := als.nodesConfigProvider.GetCurrentNodesConfig()
	currNumOfValidators := als.stakingDataProvider.GetNumOfValidatorsInCurrentEpoch()
	numOfShuffledNodes, numForcedToStay := als.computeNumShuffledNodes(currNodesConfig)
	numOfValidatorsAfterShuffling, err := safeSub(currNumOfValidators, numOfShuffledNodes)
	if err != nil {
		log.Warn(fmt.Sprintf("auctionListSelector.SelectNodesFromAuctionList: %v when trying to compute numOfValidatorsAfterShuffling = %v - %v (currNumOfValidators - numOfShuffledNodes)",
			err,
			currNumOfValidators,
			numOfShuffledNodes,
		))
		numOfValidatorsAfterShuffling = 0
	}

	maxNumNodes := currNodesConfig.MaxNumNodes
	numValidatorsAfterShufflingWithForcedToStay := numOfValidatorsAfterShuffling + numForcedToStay
	availableSlots, err := safeSub(maxNumNodes, numValidatorsAfterShufflingWithForcedToStay)
	if availableSlots == 0 || err != nil {
		log.Info(fmt.Sprintf("auctionListSelector.SelectNodesFromAuctionList: %v or zero value when trying to compute availableSlots = %v - %v (maxNodes - numOfValidatorsAfterShuffling+numForcedToStay); skip selecting nodes from auction list",
			err,
			maxNumNodes,
			numValidatorsAfterShufflingWithForcedToStay,
		))
		return nil
	}

	log.Info("auctionListSelector.SelectNodesFromAuctionList",
		"max nodes", maxNumNodes,
		"current number of validators", currNumOfValidators,
		"num of nodes which will be shuffled out", numOfShuffledNodes,
		"num forced to stay", numForcedToStay,
		"num of validators after shuffling with forced to stay", numValidatorsAfterShufflingWithForcedToStay,
		"auction list size", auctionListSize,
		fmt.Sprintf("available slots (%v - %v)", maxNumNodes, numValidatorsAfterShufflingWithForcedToStay), availableSlots,
	)

	als.auctionListDisplayer.DisplayOwnersData(ownersData)
	numOfAvailableNodeSlots := core.MinUint32(auctionListSize, availableSlots)

	sw := core.NewStopWatch()
	sw.Start("auctionListSelector.sortAuctionList")
	defer func() {
		sw.Stop("auctionListSelector.sortAuctionList")
		log.Debug("time measurements", sw.GetMeasurements()...)
	}()

	return als.sortAuctionList(ownersData, numOfAvailableNodeSlots, validatorsInfoMap, randomness)
}

func (als *auctionListSelector) getAuctionData() (map[string]*OwnerAuctionData, uint32) {
	ownersData := make(map[string]*OwnerAuctionData)
	numOfNodesInAuction := uint32(0)

	for owner, ownerData := range als.stakingDataProvider.GetOwnersData() {
		if ownerData.Qualified && len(ownerData.AuctionList) > 0 {
			numAuctionNodes := len(ownerData.AuctionList)

			ownersData[owner] = &OwnerAuctionData{
				numActiveNodes:           ownerData.NumActiveNodes,
				numAuctionNodes:          int64(numAuctionNodes),
				numQualifiedAuctionNodes: int64(numAuctionNodes),
				numStakedNodes:           ownerData.NumStakedNodes,
				totalTopUp:               ownerData.TotalTopUp,
				topUpPerNode:             ownerData.TopUpPerNode,
				qualifiedTopUpPerNode:    ownerData.TopUpPerNode,
				auctionList:              make([]state.ValidatorInfoHandler, numAuctionNodes),
			}
			copy(ownersData[owner].auctionList, ownerData.AuctionList)
			numOfNodesInAuction += uint32(numAuctionNodes)
		}
	}

	return ownersData, numOfNodesInAuction
}

func isInAuction(validator state.ValidatorInfoHandler) bool {
	return validator.GetList() == string(common.AuctionList)
}

func (als *auctionListSelector) computeNumShuffledNodes(currNodesConfig config.MaxNodesChangeConfig) (uint32, uint32) {
	numNodesToShufflePerShard := currNodesConfig.NodesToShufflePerShard
	numTotalToShuffleOut := numNodesToShufflePerShard * (als.shardCoordinator.NumberOfShards() + 1)
	epochStats := als.stakingDataProvider.GetCurrentEpochValidatorStats()

	actuallyNumLeaving := uint32(0)
	forcedToStay := uint32(0)

	for shardID := uint32(0); shardID < als.shardCoordinator.NumberOfShards(); shardID++ {
		leavingInShard, forcedToStayInShard := computeActuallyNumLeaving(shardID, epochStats, numNodesToShufflePerShard)
		actuallyNumLeaving += leavingInShard
		forcedToStay += forcedToStayInShard
	}

	leavingInMeta, forcedToStayInMeta := computeActuallyNumLeaving(core.MetachainShardId, epochStats, numNodesToShufflePerShard)
	actuallyNumLeaving += leavingInMeta
	forcedToStay += forcedToStayInMeta

	finalShuffledOut, err := safeSub(numTotalToShuffleOut, actuallyNumLeaving)
	if err != nil {
		log.Error("auctionListSelector.computeNumShuffledNodes error computing finalShuffledOut, returning default values",
			"error", err, "numTotalToShuffleOut", numTotalToShuffleOut, "actuallyNumLeaving", actuallyNumLeaving)
		return numTotalToShuffleOut, 0
	}

	return finalShuffledOut, forcedToStay
}

func computeActuallyNumLeaving(shardID uint32, epochStats epochStart.ValidatorStatsInEpoch, numNodesToShuffledPerShard uint32) (uint32, uint32) {
	numLeavingInShard := uint32(epochStats.Leaving[shardID])
	numActiveInShard := uint32(epochStats.Waiting[shardID] + epochStats.Eligible[shardID])

	log.Debug("auctionListSelector.computeActuallyNumLeaving computing",
		"shardID", shardID, "numLeavingInShard", numLeavingInShard, "numActiveInShard", numActiveInShard)

	actuallyLeaving := uint32(0)
	forcedToStay := uint32(0)
	if numLeavingInShard <= numNodesToShuffledPerShard && numActiveInShard > numLeavingInShard {
		actuallyLeaving = numLeavingInShard
	}

	if numLeavingInShard > numNodesToShuffledPerShard {
		actuallyLeaving = numNodesToShuffledPerShard
		forcedToStay = numLeavingInShard - numNodesToShuffledPerShard
	}

	log.Debug("auctionListSelector.computeActuallyNumLeaving computed",
		"actuallyLeaving", actuallyLeaving, "forcedToStay", forcedToStay)

	return actuallyLeaving, forcedToStay
}

// TODO: Move this in elrond-go-core
func safeSub(a, b uint32) (uint32, error) {
	if a < b {
		return 0, epochStart.ErrUint32SubtractionOverflow
	}
	return a - b, nil
}

func (als *auctionListSelector) sortAuctionList(
	ownersData map[string]*OwnerAuctionData,
	numOfAvailableNodeSlots uint32,
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	randomness []byte,
) error {
	softAuctionNodesConfig := als.calcSoftAuctionNodesConfig(ownersData, numOfAvailableNodeSlots)
	selectedNodes := als.selectNodes(softAuctionNodesConfig, numOfAvailableNodeSlots, randomness)
	return markAuctionNodesAsSelected(selectedNodes, validatorsInfoMap)
}

func (als *auctionListSelector) calcSoftAuctionNodesConfig(
	data map[string]*OwnerAuctionData,
	numAvailableSlots uint32,
) map[string]*OwnerAuctionData {
	ownersData := copyOwnersData(data)
	minTopUp, maxTopUp := als.getMinMaxPossibleTopUp(ownersData)
	log.Debug("auctionListSelector: calc min and max possible top up",
		"min top up per node", getPrettyValue(minTopUp, als.softAuctionConfig.denominator),
		"max top up per node", getPrettyValue(maxTopUp, als.softAuctionConfig.denominator),
	)

	topUp := big.NewInt(0).SetBytes(minTopUp.Bytes())
	previousConfig := copyOwnersData(ownersData)
	iterationNumber := uint64(0)
	maxNumberOfIterationsReached := false

	for ; topUp.Cmp(maxTopUp) < 0 && !maxNumberOfIterationsReached; topUp.Add(topUp, als.softAuctionConfig.step) {
		previousConfig = copyOwnersData(ownersData)
		numNodesQualifyingForTopUp := calcNodesConfig(ownersData, topUp)

		if numNodesQualifyingForTopUp < int64(numAvailableSlots) {
			break
		}

		iterationNumber++
		maxNumberOfIterationsReached = iterationNumber >= als.softAuctionConfig.maxNumberOfIterations
	}

	log.Debug("auctionListSelector: found min required",
		"topUp", getPrettyValue(topUp, als.softAuctionConfig.denominator),
		"after num of iterations", iterationNumber,
	)
	return previousConfig
}

func (als *auctionListSelector) getMinMaxPossibleTopUp(ownersData map[string]*OwnerAuctionData) (*big.Int, *big.Int) {
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

func copyOwnersData(ownersData map[string]*OwnerAuctionData) map[string]*OwnerAuctionData {
	ret := make(map[string]*OwnerAuctionData)
	for owner, data := range ownersData {
		ret[owner] = &OwnerAuctionData{
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

func calcNodesConfig(ownersData map[string]*OwnerAuctionData, topUp *big.Int) int64 {
	numNodesQualifyingForTopUp := int64(0)

	for ownerPubKey, owner := range ownersData {
		activeNodes := big.NewInt(owner.numActiveNodes)
		topUpActiveNodes := big.NewInt(0).Mul(topUp, activeNodes)
		validatorTopUpForAuction := big.NewInt(0).Sub(owner.totalTopUp, topUpActiveNodes)
		if validatorTopUpForAuction.Cmp(topUp) < 0 {
			delete(ownersData, ownerPubKey)
			continue
		}

		qualifiedNodesBigInt := big.NewInt(0).Div(validatorTopUpForAuction, topUp)
		qualifiedNodes := qualifiedNodesBigInt.Int64()
		isNumQualifiedNodesOverflow := !qualifiedNodesBigInt.IsUint64()

		if qualifiedNodes > owner.numAuctionNodes || isNumQualifiedNodesOverflow {
			numNodesQualifyingForTopUp += owner.numAuctionNodes
		} else {
			numNodesQualifyingForTopUp += qualifiedNodes
			owner.numQualifiedAuctionNodes = qualifiedNodes

			ownerRemainingNodes := big.NewInt(owner.numActiveNodes + owner.numQualifiedAuctionNodes)
			owner.qualifiedTopUpPerNode = big.NewInt(0).Div(owner.totalTopUp, ownerRemainingNodes)
		}
	}

	return numNodesQualifyingForTopUp
}

func markAuctionNodesAsSelected(
	selectedNodes []state.ValidatorInfoHandler,
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
) error {
	for _, node := range selectedNodes {
		newNode := node.ShallowClone()
		newNode.SetPreviousList(node.GetList())
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
