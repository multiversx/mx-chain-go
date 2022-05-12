package metachain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/display"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

type auctionListSelector struct {
	currentNodesEnableConfig config.MaxNodesChangeConfig
	shardCoordinator         sharding.Coordinator
	stakingDataProvider      epochStart.StakingDataProvider
	maxNodesEnableConfig     []config.MaxNodesChangeConfig
}

type AuctionListSelectorArgs struct {
	ShardCoordinator     sharding.Coordinator
	StakingDataProvider  epochStart.StakingDataProvider
	EpochNotifier        process.EpochNotifier
	MaxNodesEnableConfig []config.MaxNodesChangeConfig
}

func NewAuctionListSelector(args AuctionListSelectorArgs) (*auctionListSelector, error) {
	asl := &auctionListSelector{
		shardCoordinator:    args.ShardCoordinator,
		stakingDataProvider: args.StakingDataProvider,
	}

	asl.maxNodesEnableConfig = make([]config.MaxNodesChangeConfig, len(args.MaxNodesEnableConfig))
	copy(asl.maxNodesEnableConfig, args.MaxNodesEnableConfig)
	args.EpochNotifier.RegisterNotifyHandler(asl)

	return asl, nil
}

func (als *auctionListSelector) selectNodesFromAuctionList(validatorsInfoMap state.ShardValidatorsInfoMapHandler, randomness []byte) error {
	auctionList, currNumOfValidators := getAuctionListAndNumOfValidators(validatorsInfoMap)
	numOfShuffledNodes := als.currentNodesEnableConfig.NodesToShufflePerShard * (als.shardCoordinator.NumberOfShards() + 1)

	numOfValidatorsAfterShuffling, err := safeSub(currNumOfValidators, numOfShuffledNodes)
	if err != nil {
		log.Warn(fmt.Sprintf("%v when trying to compute numOfValidatorsAfterShuffling = %v - %v (currNumOfValidators - numOfShuffledNodes)",
			err,
			currNumOfValidators,
			numOfShuffledNodes,
		))
		numOfValidatorsAfterShuffling = 0
	}

	availableSlots, err := safeSub(als.currentNodesEnableConfig.MaxNumNodes, numOfValidatorsAfterShuffling)
	if availableSlots == 0 || err != nil {
		log.Info(fmt.Sprintf("%v or zero value when trying to compute availableSlots = %v - %v (maxNodes - numOfValidatorsAfterShuffling); skip selecting nodes from auction list",
			err,
			als.currentNodesEnableConfig.MaxNumNodes,
			numOfValidatorsAfterShuffling,
		))
		return nil
	}

	auctionListSize := uint32(len(auctionList))
	log.Info("systemSCProcessor.selectNodesFromAuctionList",
		"max nodes", als.currentNodesEnableConfig.MaxNumNodes,
		"current number of validators", currNumOfValidators,
		"num of nodes which will be shuffled out", numOfShuffledNodes,
		"num of validators after shuffling", numOfValidatorsAfterShuffling,
		"auction list size", auctionListSize,
		fmt.Sprintf("available slots (%v -%v)", als.currentNodesEnableConfig.MaxNumNodes, numOfValidatorsAfterShuffling), availableSlots,
	)

	err = als.sortAuctionList(auctionList, randomness)
	if err != nil {
		return err
	}

	numOfAvailableNodeSlots := core.MinUint32(auctionListSize, availableSlots)
	als.displayAuctionList(auctionList, numOfAvailableNodeSlots)

	for i := uint32(0); i < numOfAvailableNodeSlots; i++ {
		newNode := auctionList[i]
		newNode.SetList(string(common.SelectedFromAuctionList))
		err = validatorsInfoMap.Replace(auctionList[i], newNode)
		if err != nil {
			return err
		}
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

func getAuctionListAndNumOfValidators(validatorsInfoMap state.ShardValidatorsInfoMapHandler) ([]state.ValidatorInfoHandler, uint32) {
	auctionList := make([]state.ValidatorInfoHandler, 0)
	numOfValidators := uint32(0)

	for _, validator := range validatorsInfoMap.GetAllValidatorsInfo() {
		if validator.GetList() == string(common.AuctionList) {
			auctionList = append(auctionList, validator)
			continue
		}
		if isValidator(validator) {
			numOfValidators++
		}
	}

	return auctionList, numOfValidators
}

func (als *auctionListSelector) sortAuctionList(auctionList []state.ValidatorInfoHandler, randomness []byte) error {
	if len(auctionList) == 0 {
		return nil
	}

	validatorTopUpMap, err := als.getValidatorTopUpMap(auctionList)
	if err != nil {
		return fmt.Errorf("%w: %v", epochStart.ErrSortAuctionList, err)
	}

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

	return nil
}

func (als *auctionListSelector) getValidatorTopUpMap(validators []state.ValidatorInfoHandler) (map[string]*big.Int, error) {
	ret := make(map[string]*big.Int, len(validators))

	for _, validator := range validators {
		pubKey := validator.GetPublicKey()
		topUp, err := als.stakingDataProvider.GetNodeStakedTopUp(pubKey)
		if err != nil {
			return nil, fmt.Errorf("%w when trying to get top up per node for %s", err, hex.EncodeToString(pubKey))
		}

		ret[string(pubKey)] = topUp
	}

	return ret, nil
}

func calcNormRand(randomness []byte, expectedLen int) []byte {
	rand := randomness
	randLen := len(rand)

	if expectedLen > randLen {
		repeatedCt := expectedLen/randLen + 1 // todo: fix possible div by 0
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
	if log.GetLevel() > logger.LogDebug {
		return
	}

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
			hex.EncodeToString([]byte(owner)),
			hex.EncodeToString(pubKey),
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
	log.Debug(message)
}

func (als *auctionListSelector) EpochConfirmed(epoch uint32, _ uint64) {
	for _, maxNodesConfig := range als.maxNodesEnableConfig {
		if epoch >= maxNodesConfig.EpochEnable {
			als.currentNodesEnableConfig = maxNodesConfig
		}
	}
}

func (als *auctionListSelector) IsInterfaceNil() bool {
	return als == nil
}
