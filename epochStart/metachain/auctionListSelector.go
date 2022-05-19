package metachain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/display"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

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

	auctionList, currNumOfValidators, err := als.getAuctionListAndNumOfValidators(validatorsInfoMap, unqualifiedOwners)
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

	numOfAvailableNodeSlots := core.MinUint32(auctionListSize, availableSlots)
	err = als.sortAuctionList(auctionList, numOfAvailableNodeSlots, validatorsInfoMap, randomness)
	if err != nil {
		return err
	}

	als.displayAuctionList(auctionList, numOfAvailableNodeSlots)
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
) ([]state.ValidatorInfoHandler, uint32, error) {
	auctionList := make([]state.ValidatorInfoHandler, 0)
	numOfValidators := uint32(0)

	for _, validator := range validatorsInfoMap.GetAllValidatorsInfo() {
		owner, err := als.stakingDataProvider.GetBlsKeyOwner(validator.GetPublicKey())
		if err != nil {
			return nil, 0, err
		}

		_, isUnqualified := unqualifiedOwners[owner]
		if isUnqualified {
			log.Debug("auctionListSelector: found unqualified owner, do not add validator in auction selection",
				"owner", hex.EncodeToString([]byte(owner)),
				"bls key", hex.EncodeToString(validator.GetPublicKey()),
			)
			continue
		}

		if validator.GetList() == string(common.AuctionList) {
			auctionList = append(auctionList, validator)
			continue
		}
		if isValidator(validator) {
			numOfValidators++
		}
	}

	return auctionList, numOfValidators, nil
}

type ownerData struct {
	activeNodes  int64
	auctionNodes int64
	stakedNodes  int64
	totalTopUp   *big.Int
	topUpPerNode *big.Int
	auctionList  []state.ValidatorInfoHandler
}

func (als *auctionListSelector) getOwnersData(auctionList []state.ValidatorInfoHandler) (map[string]*ownerData, error) {
	ownersData := make(map[string]*ownerData)

	for _, node := range auctionList {
		owner, err := als.stakingDataProvider.GetBlsKeyOwner(node.GetPublicKey())
		if err != nil {
			return nil, err
		}

		stakedNodes, err := als.stakingDataProvider.GetNumStakedNodes([]byte(owner))
		if err != nil {
			return nil, err
		}

		if stakedNodes == 0 {
			return nil, process.ErrNodeIsNotSynced
		}

		totalTopUp, err := als.stakingDataProvider.GetTotalTopUp([]byte(owner))
		if err != nil {
			return nil, err
		}

		data, exists := ownersData[owner]
		if exists {
			data.auctionNodes++
			data.activeNodes--
			data.auctionList = append(data.auctionList, node)
		} else {
			ownersData[owner] = &ownerData{
				auctionNodes: 1,
				activeNodes:  stakedNodes - 1,
				stakedNodes:  stakedNodes,
				totalTopUp:   big.NewInt(0).SetBytes(totalTopUp.Bytes()),
				topUpPerNode: big.NewInt(0).Div(totalTopUp, big.NewInt(stakedNodes)),
				auctionList:  []state.ValidatorInfoHandler{node},
			}
		}
	}

	return ownersData, nil
}

func copyOwnersData(ownersData map[string]*ownerData) map[string]*ownerData {
	ret := make(map[string]*ownerData)
	for owner, data := range ownersData {
		ret[owner] = &ownerData{
			activeNodes:  data.activeNodes,
			auctionNodes: data.auctionNodes,
			stakedNodes:  data.stakedNodes,
			totalTopUp:   data.totalTopUp,
			topUpPerNode: data.topUpPerNode,
			auctionList:  make([]state.ValidatorInfoHandler, len(data.auctionList)),
		}
		copy(ret[owner].auctionList, data.auctionList)
	}

	return ret
}

func (als *auctionListSelector) getMinRequiredTopUp(
	auctionList []state.ValidatorInfoHandler,
	numAvailableSlots uint32,
	randomness []byte,
) ([]state.ValidatorInfoHandler, *big.Int, error) {
	ownersData, err := als.getOwnersData(auctionList)
	if err != nil {
		return nil, nil, err
	}

	minTopUp := big.NewInt(1)       // pornim de la topup cel mai slab din lista initiala
	maxTopUp := big.NewInt(1000000) // todo: extract to const // max top up from auction list
	step := big.NewInt(100)

	for topUp := big.NewInt(0).SetBytes(minTopUp.Bytes()); topUp.Cmp(maxTopUp) < 0; topUp.Add(topUp, step) {
		numNodesQualifyingForTopUp := int64(0)
		previousConfig := copyOwnersData(ownersData)

		for ownerPubKey, owner := range ownersData {
			activeNodes := big.NewInt(owner.activeNodes)
			topUpActiveNodes := big.NewInt(0).Mul(topUp, activeNodes)
			validatorTopUpForAuction := big.NewInt(0).Sub(owner.totalTopUp, topUpActiveNodes)
			if validatorTopUpForAuction.Cmp(topUp) < 0 {
				delete(ownersData, ownerPubKey)
				continue
			}

			qualifiedNodes := big.NewInt(0).Div(validatorTopUpForAuction, topUp)
			qualifiedNodesInt := qualifiedNodes.Int64()
			if qualifiedNodesInt > owner.auctionNodes {
				numNodesQualifyingForTopUp += owner.auctionNodes
			} else {
				numNodesQualifyingForTopUp += qualifiedNodesInt

				owner.auctionNodes = qualifiedNodesInt

				ownerRemainingNodes := big.NewInt(owner.activeNodes + owner.auctionNodes)
				owner.topUpPerNode = big.NewInt(0).Div(owner.totalTopUp, ownerRemainingNodes)
			}
		}

		if numNodesQualifyingForTopUp < int64(numAvailableSlots) {

			if topUp.Cmp(minTopUp) == 0 {
				selectedNodes := als.selectNodes(previousConfig, uint32(len(auctionList)), randomness)

				return selectedNodes, big.NewInt(0), nil
			} else {
				selectedNodes := als.selectNodes(previousConfig, numAvailableSlots, randomness)
				return selectedNodes, topUp.Sub(topUp, step), nil
			}
		}

	}

	return nil, nil, errors.New("COULD NOT FIND TOPUP")
}

func (als *auctionListSelector) selectNodes(ownersData map[string]*ownerData, numAvailableSlots uint32, randomness []byte) []state.ValidatorInfoHandler {
	selectedFromAuction := make([]state.ValidatorInfoHandler, 0)
	validatorTopUpMap := make(map[string]*big.Int)

	for _, owner := range ownersData {
		sortListByXORWithRand(owner.auctionList, randomness)
		for i := int64(0); i < owner.auctionNodes; i++ {
			currNode := owner.auctionList[i]
			validatorTopUpMap[string(currNode.GetPublicKey())] = big.NewInt(0).SetBytes(owner.topUpPerNode.Bytes())
		}

		selectedFromAuction = append(selectedFromAuction, owner.auctionList[:owner.auctionNodes]...)
	}

	als.sortValidators(selectedFromAuction, validatorTopUpMap, randomness)

	selectedFromAuction = selectedFromAuction[:numAvailableSlots]

	return selectedFromAuction
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

func (als *auctionListSelector) sortAuctionList(
	auctionList []state.ValidatorInfoHandler,
	numOfAvailableNodeSlots uint32,
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	randomness []byte,
) error {
	if len(auctionList) == 0 {
		return nil
	}

	selectedNodes, minTopUp, err := als.getMinRequiredTopUp(auctionList, numOfAvailableNodeSlots, randomness)
	if err != nil {
		return err
	}

	for _, node := range selectedNodes {
		newNode := node
		newNode.SetList(string(common.SelectedFromAuctionList))
		err = validatorsInfoMap.Replace(node, newNode)
		if err != nil {
			return err
		}
	}

	_ = minTopUp
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

func (als *auctionListSelector) getValidatorTopUpMap(validators []state.ValidatorInfoHandler) (map[string]*big.Int, error) {
	ret := make(map[string]*big.Int, len(validators))

	for _, validator := range validators {
		pubKey := validator.GetPublicKey()
		topUp, err := als.stakingDataProvider.GetNodeStakedTopUp(pubKey)
		if err != nil {
			return nil, fmt.Errorf("%w when trying to get top up per node for %s", err, hex.EncodeToString(pubKey))
		}

		ret[string(pubKey)] = big.NewInt(0).SetBytes(topUp.Bytes())
	}

	return ret, nil
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

func (als *auctionListSelector) displayAuctionList(auctionList []state.ValidatorInfoHandler, numOfSelectedNodes uint32) {
	//if log.GetLevel() > logger.LogDebug {
	//	return
	//}

	tableHeader := []string{"Owner", "Registered key", "TopUp per node"}
	lines := make([]*display.LineData, 0, len(auctionList))
	horizontalLine := false
	for idx, validator := range auctionList {
		pubKey := validator.GetPublicKey()

		owner, err := als.stakingDataProvider.GetBlsKeyOwner(pubKey)
		log.LogIfError(err)

		topUp, err := als.stakingDataProvider.GetNodeStakedTopUp(pubKey)
		log.LogIfError(err)

		horizontalLine = uint32(idx) == numOfSelectedNodes-1
		line := display.NewLineData(horizontalLine, []string{
			(owner),
			string(pubKey),
			topUp.String(),
		})
		lines = append(lines, line)
	}

	table, err := display.CreateTableString(tableHeader, lines)
	if err != nil {
		log.Error("could not create table", "error", err)
		return
	}

	message := fmt.Sprintf("Auction list\n%s", table)
	log.Info(message)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (als *auctionListSelector) IsInterfaceNil() bool {
	return als == nil
}
