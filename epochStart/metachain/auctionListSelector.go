package metachain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

const oneEGLD = 1000000000000000000 // with 18 decimals = 1 EGLD
const minEGLD = 1                   // with 18 decimals = 0.00...01 egld
const maxEGLD = 21000000            // without 18 decimals

type auctionListSelector struct {
	shardCoordinator    sharding.Coordinator
	stakingDataProvider epochStart.StakingDataProvider
	nodesConfigProvider epochStart.MaxNodesChangeConfigProvider
}

// AuctionListSelectorArgs is a struct placeholder for all arguments required to create a NewAuctionListSelector
type AuctionListSelectorArgs struct {
	ShardCoordinator             sharding.Coordinator
	StakingDataProvider          epochStart.StakingDataProvider
	MaxNodesChangeConfigProvider epochStart.MaxNodesChangeConfigProvider
}

// NewAuctionListSelector will create a new auctionListSelector, which handles selection of nodes from auction list based
// on their top up
func NewAuctionListSelector(args AuctionListSelectorArgs) (*auctionListSelector, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, epochStart.ErrNilShardCoordinator
	}
	if check.IfNil(args.StakingDataProvider) {
		return nil, epochStart.ErrNilStakingDataProvider
	}
	if check.IfNil(args.MaxNodesChangeConfigProvider) {
		return nil, epochStart.ErrNilMaxNodesChangeConfigProvider
	}

	asl := &auctionListSelector{
		shardCoordinator:    args.ShardCoordinator,
		stakingDataProvider: args.StakingDataProvider,
		nodesConfigProvider: args.MaxNodesChangeConfigProvider,
	}

	return asl, nil
}

// SelectNodesFromAuctionList will select nodes from validatorsInfoMap based on their top up. If two or more validators
// have the same top-up, then sorting will be done based on blsKey XOR randomness. Selected nodes will have their list set
// to common.SelectNodesFromAuctionList
// Depends that dat is filled in staking data provider
func (als *auctionListSelector) SelectNodesFromAuctionList(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	unqualifiedOwners map[string]struct{},
	randomness []byte,
) error {
	if len(randomness) == 0 {
		return process.ErrNilRandSeed
	}

	currNodesConfig := als.nodesConfigProvider.GetCurrentNodesConfig()
	numOfShuffledNodes := currNodesConfig.NodesToShufflePerShard * (als.shardCoordinator.NumberOfShards() + 1)

	auctionList, ownersData, currNumOfValidators, err := als.getAuctionListAndNumOfValidators(validatorsInfoMap, unqualifiedOwners)
	if err != nil {
		return err
	}

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

	auctionListSize := uint32(len(auctionList))
	log.Info("systemSCProcessor.SelectNodesFromAuctionList",
		"max nodes", maxNumNodes,
		"current number of validators", currNumOfValidators,
		"num of nodes which will be shuffled out", numOfShuffledNodes,
		"num of validators after shuffling", numOfValidatorsAfterShuffling,
		"auction list size", auctionListSize,
		fmt.Sprintf("available slots (%v -%v)", maxNumNodes, numOfValidatorsAfterShuffling), availableSlots,
	)

	if auctionListSize == 0 {
		log.Debug("auctionListSelector.SelectNodesFromAuctionList: empty auction list; skip selection")
		return nil
	}

	als.displayOwnersConfig(ownersData)
	numOfAvailableNodeSlots := core.MinUint32(auctionListSize, availableSlots)
	err = als.sortAuctionList(ownersData, numOfAvailableNodeSlots, validatorsInfoMap, randomness)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Move this in elrond-go-core
func safeSub(a, b uint32) (uint32, error) {
	if a < b {
		return 0, core.ErrSubtractionOverflow
	}
	return a - b, nil
}

func (als *auctionListSelector) getAuctionListAndNumOfValidators(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	unqualifiedOwners map[string]struct{},
) ([]state.ValidatorInfoHandler, map[string]*ownerData, uint32, error) {
	ownersData := make(map[string]*ownerData)
	auctionList := make([]state.ValidatorInfoHandler, 0)
	numOfValidators := uint32(0)

	for _, validator := range validatorsInfoMap.GetAllValidatorsInfo() {
		owner, err := als.stakingDataProvider.GetBlsKeyOwner(validator.GetPublicKey())
		if err != nil {
			return nil, nil, 0, err
		}

		if isInAuction(validator) {
			_, isUnqualified := unqualifiedOwners[owner]
			if isUnqualified {
				log.Debug("auctionListSelector: found node in auction with unqualified owner, do not add it to selection",
					"owner", owner,
					"bls key", string(validator.GetPublicKey()),
				)
				continue
			}

			err = als.addOwnerData(validator, ownersData)
			if err != nil {
				return nil, nil, 0, err
			}

			auctionList = append(auctionList, validator)
			continue
		}
		if isValidator(validator) {
			numOfValidators++
		}
	}

	return auctionList, ownersData, numOfValidators, nil
}

func isInAuction(validator state.ValidatorInfoHandler) bool {
	return validator.GetList() == string(common.AuctionList)
}

type ownerData struct {
	numActiveNodes           int64
	numAuctionNodes          int64
	numQualifiedAuctionNodes int64
	numStakedNodes           int64
	totalTopUp               *big.Int
	topUpPerNode             *big.Int
	qualifiedTopUpPerNode    *big.Int
	auctionList              []state.ValidatorInfoHandler
}

func (als *auctionListSelector) addOwnerData(
	validator state.ValidatorInfoHandler,
	ownersData map[string]*ownerData,
) error {
	owner, err := als.stakingDataProvider.GetBlsKeyOwner(validator.GetPublicKey())
	if err != nil {
		return err
	}

	stakedNodes, err := als.stakingDataProvider.GetNumStakedNodes([]byte(owner))
	if err != nil {
		return err
	}

	if stakedNodes == 0 {
		return fmt.Errorf("auctionListSelector.getOwnersDat: error: %w, owner: %s, node: %s",
			epochStart.ErrOwnerHasNoStakedNode,
			hex.EncodeToString([]byte(owner)),
			hex.EncodeToString(validator.GetPublicKey()),
		)
	}

	totalTopUp, err := als.stakingDataProvider.GetTotalTopUp([]byte(owner))
	if err != nil {
		return err
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
		ownersData[owner] = &ownerData{
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

func copyOwnersData(ownersData map[string]*ownerData) map[string]*ownerData {
	ret := make(map[string]*ownerData)
	for owner, data := range ownersData {
		ret[owner] = &ownerData{
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

func getMinMaxPossibleTopUp(ownersData map[string]*ownerData) (*big.Int, *big.Int) {
	min := big.NewInt(0).Mul(big.NewInt(oneEGLD), big.NewInt(maxEGLD))
	max := big.NewInt(0)

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

	minPossible := big.NewInt(minEGLD)
	if min.Cmp(minPossible) < 0 {
		min = minPossible
	}

	return min, max
}

func (als *auctionListSelector) calcSoftAuctionNodesConfig(
	ownersData map[string]*ownerData,
	numAvailableSlots uint32,
) (map[string]*ownerData, error) {
	minTopUp, maxTopUp := getMinMaxPossibleTopUp(ownersData) // TODO: What happens if min>max or MIN = MAX?
	log.Info("auctionListSelector: calc min and max possible top up",
		"min top up", minTopUp.String(),
		"max top up", maxTopUp.String(),
	)

	step := big.NewInt(10) // todo: granulate step if max- min < step???? + 10 egld for real
	previousConfig := copyOwnersData(ownersData)
	topUp := big.NewInt(0).SetBytes(minTopUp.Bytes())

	for ; topUp.Cmp(maxTopUp) < 0; topUp.Add(topUp, step) {
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
	displayRequiredTopUp(topUp, maxTopUp, minTopUp, step)
	return previousConfig, nil
}

func (als *auctionListSelector) selectNodes(
	ownersData map[string]*ownerData,
	numAvailableSlots uint32,
	randomness []byte,
) []state.ValidatorInfoHandler {
	selectedFromAuction := make([]state.ValidatorInfoHandler, 0)
	validatorTopUpMap := make(map[string]*big.Int)

	for _, owner := range ownersData {
		sortListByXORWithRand(owner.auctionList, randomness)
		addQualifiedValidatorsTopUpInMap(owner, validatorTopUpMap)
		selectedFromAuction = append(selectedFromAuction, owner.auctionList[:owner.numQualifiedAuctionNodes]...)
	}

	als.displayOwnersSelectedConfig(ownersData, randomness)
	als.sortValidators(selectedFromAuction, validatorTopUpMap, randomness)
	als.displayAuctionListV2(selectedFromAuction, ownersData, numAvailableSlots)

	return selectedFromAuction[:numAvailableSlots]
}

func addQualifiedValidatorsTopUpInMap(owner *ownerData, validatorTopUpMap map[string]*big.Int) {
	for i := int64(0); i < owner.numQualifiedAuctionNodes; i++ {
		validatorPubKey := owner.auctionList[i].GetPublicKey()
		validatorTopUpMap[string(validatorPubKey)] = big.NewInt(0).SetBytes(owner.qualifiedTopUpPerNode.Bytes())
	}
}

func sortListByXORWithRand(list []state.ValidatorInfoHandler, randomness []byte) {
	pubKeyLen := len(list[0].GetPublicKey())
	normRandomness := calcNormRand(randomness, pubKeyLen)

	sort.SliceStable(list, func(i, j int) bool {
		pubKey1 := list[i].GetPublicKey()
		pubKey2 := list[j].GetPublicKey()

		return compareByXORWithRandomness(pubKey1, pubKey2, normRandomness)
	})
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

func (als *auctionListSelector) sortAuctionList(
	ownersData map[string]*ownerData,
	numOfAvailableNodeSlots uint32,
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	randomness []byte,
) error {
	// TODO: Here add a stopwatch to measure execution time
	softAuctionNodesConfig, err := als.calcSoftAuctionNodesConfig(ownersData, numOfAvailableNodeSlots)
	if err != nil {
		return err
	}

	selectedNodes := als.selectNodes(softAuctionNodesConfig, numOfAvailableNodeSlots, randomness)
	err = markAuctionNodesAsSelected(selectedNodes, validatorsInfoMap)
	if err != nil {
		return err
	}

	return nil
}

func (als *auctionListSelector) sortValidators(
	auctionList []state.ValidatorInfoHandler,
	validatorTopUpMap map[string]*big.Int,
	randomness []byte,
) {
	pubKeyLen := len(auctionList[0].GetPublicKey())
	normRandomness := calcNormRand(randomness, pubKeyLen)
	sort.SliceStable(auctionList, func(i, j int) bool {
		pubKey1 := auctionList[i].GetPublicKey()
		pubKey2 := auctionList[j].GetPublicKey()

		nodeTopUpPubKey1 := validatorTopUpMap[string(pubKey1)]
		nodeTopUpPubKey2 := validatorTopUpMap[string(pubKey2)]

		if nodeTopUpPubKey1.Cmp(nodeTopUpPubKey2) == 0 {
			return compareByXORWithRandomness(pubKey1, pubKey2, normRandomness)
		}

		return nodeTopUpPubKey1.Cmp(nodeTopUpPubKey2) > 0
	})
}

func calcNormRand(randomness []byte, expectedLen int) []byte {
	rand := randomness
	randLen := len(rand)

	if expectedLen > randLen {
		repeatedCt := expectedLen/randLen + 1
		rand = bytes.Repeat(randomness, repeatedCt)
	}

	rand = rand[:expectedLen]
	return rand
}

func compareByXORWithRandomness(pubKey1, pubKey2, randomness []byte) bool {
	xorLen := len(randomness)

	key1Xor := make([]byte, xorLen)
	key2Xor := make([]byte, xorLen)

	for idx := 0; idx < xorLen; idx++ {
		key1Xor[idx] = pubKey1[idx] ^ randomness[idx]
		key2Xor[idx] = pubKey2[idx] ^ randomness[idx]
	}

	return bytes.Compare(key1Xor, key2Xor) == 1
}

// IsInterfaceNil checks if the underlying pointer is nil
func (als *auctionListSelector) IsInterfaceNil() bool {
	return als == nil
}
