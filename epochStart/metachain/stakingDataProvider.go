package metachain

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type ownerStats struct {
	numEligible          int
	numStakedNodes       int64
	numActiveNodes       int64
	totalTopUp           *big.Int
	topUpPerNode         *big.Int
	totalStaked          *big.Int
	eligibleBaseStake    *big.Int
	eligibleTopUpStake   *big.Int
	eligibleTopUpPerNode *big.Int
	blsKeys              [][]byte
	auctionList          []state.ValidatorInfoHandler
	qualified            bool
}

type ownerInfoSC struct {
	topUpValue       *big.Int
	totalStakedValue *big.Int
	numStakedWaiting *big.Int
	blsKeys          [][]byte
}

type stakingDataProvider struct {
	mutStakingData             sync.RWMutex
	cache                      map[string]*ownerStats
	systemVM                   vmcommon.VMExecutionHandler
	totalEligibleStake         *big.Int
	totalEligibleTopUpStake    *big.Int
	minNodePrice               *big.Int
	numOfValidatorsInCurrEpoch uint32
	enableEpochsHandler        common.EnableEpochsHandler
	validatorStatsInEpoch      epochStart.ValidatorStatsInEpoch
}

// StakingDataProviderArgs is a struct placeholder for all arguments required to create a NewStakingDataProvider
type StakingDataProviderArgs struct {
	EnableEpochsHandler common.EnableEpochsHandler
	SystemVM            vmcommon.VMExecutionHandler
	MinNodePrice        string
}

// NewStakingDataProvider will create a new instance of a staking data provider able to aid in the final rewards
// computation as this will retrieve the staking data from the system VM
func NewStakingDataProvider(args StakingDataProviderArgs) (*stakingDataProvider, error) {
	if check.IfNil(args.SystemVM) {
		return nil, epochStart.ErrNilSystemVmInstance
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, epochStart.ErrNilEnableEpochsHandler
	}

	nodePrice, ok := big.NewInt(0).SetString(args.MinNodePrice, 10)
	if !ok || nodePrice.Cmp(big.NewInt(0)) <= 0 {
		return nil, epochStart.ErrInvalidMinNodePrice
	}

	sdp := &stakingDataProvider{
		systemVM:                args.SystemVM,
		cache:                   make(map[string]*ownerStats),
		minNodePrice:            nodePrice,
		totalEligibleStake:      big.NewInt(0),
		totalEligibleTopUpStake: big.NewInt(0),
		enableEpochsHandler:     args.EnableEpochsHandler,
		validatorStatsInEpoch: epochStart.ValidatorStatsInEpoch{
			Eligible: make(map[uint32]int),
			Waiting:  make(map[uint32]int),
			Leaving:  make(map[uint32]int),
		},
	}

	return sdp, nil
}

// Clean will reset the inner state of the called instance
func (sdp *stakingDataProvider) Clean() {
	sdp.mutStakingData.Lock()
	sdp.cache = make(map[string]*ownerStats)
	sdp.totalEligibleStake.SetInt64(0)
	sdp.totalEligibleTopUpStake.SetInt64(0)
	sdp.numOfValidatorsInCurrEpoch = 0
	sdp.validatorStatsInEpoch = epochStart.ValidatorStatsInEpoch{
		Eligible: make(map[uint32]int),
		Waiting:  make(map[uint32]int),
		Leaving:  make(map[uint32]int),
	}
	sdp.mutStakingData.Unlock()
}

// GetTotalStakeEligibleNodes returns the total stake backing the current epoch eligible nodes
// This value is populated by a previous call to PrepareStakingData (done for epoch start)
func (sdp *stakingDataProvider) GetTotalStakeEligibleNodes() *big.Int {
	sdp.mutStakingData.RLock()
	defer sdp.mutStakingData.RUnlock()

	return big.NewInt(0).Set(sdp.totalEligibleStake)
}

// GetTotalTopUpStakeEligibleNodes returns the stake in excess of the minimum stake required, that is backing the
// current epoch eligible nodes
// This value is populated by a previous call to PrepareStakingData (done for epoch start)
func (sdp *stakingDataProvider) GetTotalTopUpStakeEligibleNodes() *big.Int {
	sdp.mutStakingData.RLock()
	defer sdp.mutStakingData.RUnlock()

	return big.NewInt(0).Set(sdp.totalEligibleTopUpStake)
}

// GetNodeStakedTopUp returns the owner of provided bls key staking stats for the current epoch
func (sdp *stakingDataProvider) GetNodeStakedTopUp(blsKey []byte) (*big.Int, error) {
	owner, err := sdp.GetBlsKeyOwner(blsKey)
	if err != nil {
		log.Debug("GetOwnerStakingStats", "key", hex.EncodeToString(blsKey), "error", err)
		return nil, err
	}

	ownerInfo, ok := sdp.cache[owner]
	if !ok {
		return nil, epochStart.ErrOwnerDoesntHaveEligibleNodesInEpoch
	}

	return ownerInfo.eligibleTopUpPerNode, nil
}

// PrepareStakingData prepares the staking data for the given map of node keys per shard
func (sdp *stakingDataProvider) PrepareStakingData(validatorsMap state.ShardValidatorsInfoMapHandler) error {
	sdp.Clean()

	for _, validator := range validatorsMap.GetAllValidatorsInfo() {
		err := sdp.loadDataForBlsKey(validator)
		if err != nil {
			return err
		}
	}

	sdp.processStakingData()

	return nil
}

func (sdp *stakingDataProvider) processStakingData() {
	totalEligibleStake := big.NewInt(0)
	totalEligibleTopUpStake := big.NewInt(0)

	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	for _, owner := range sdp.cache {
		ownerEligibleNodes := big.NewInt(int64(owner.numEligible))
		ownerStakePerNode := big.NewInt(0)
		if owner.numStakedNodes == 0 {
			ownerStakePerNode.Set(sdp.minNodePrice)
		} else {
			ownerStakePerNode.Div(owner.totalStaked, big.NewInt(owner.numStakedNodes))
		}

		ownerEligibleStake := big.NewInt(0).Mul(ownerStakePerNode, ownerEligibleNodes)
		owner.eligibleBaseStake = big.NewInt(0).Mul(ownerEligibleNodes, sdp.minNodePrice)
		owner.eligibleTopUpStake = big.NewInt(0).Sub(ownerEligibleStake, owner.eligibleBaseStake)

		totalEligibleStake.Add(totalEligibleStake, ownerEligibleStake)
		totalEligibleTopUpStake.Add(totalEligibleTopUpStake, owner.eligibleTopUpStake)

		owner.eligibleTopUpPerNode = big.NewInt(0).Div(owner.eligibleTopUpStake, ownerEligibleNodes)
	}

	sdp.totalEligibleTopUpStake = totalEligibleTopUpStake
	sdp.totalEligibleStake = totalEligibleStake
}

// FillValidatorInfo will fill the validator info for the bls key if it was not already filled
func (sdp *stakingDataProvider) FillValidatorInfo(validator state.ValidatorInfoHandler) error {
	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	_, err := sdp.getAndFillOwnerStats(validator)
	return err
}

func (sdp *stakingDataProvider) getAndFillOwnerStats(validator state.ValidatorInfoHandler) (*ownerStats, error) {
	blsKey := validator.GetPublicKey()
	owner, err := sdp.GetBlsKeyOwner(blsKey)
	if err != nil {
		log.Debug("error fill owner stats", "step", "get owner from bls", "key", hex.EncodeToString(blsKey), "error", err)
		return nil, err
	}

	ownerData, err := sdp.fillOwnerData(owner, validator)
	if err != nil {
		log.Debug("error fill owner stats", "step", "get owner data", "key", hex.EncodeToString(blsKey), "owner", hex.EncodeToString([]byte(owner)), "error", err)
		return nil, err
	}

	if isValidator(validator) {
		sdp.numOfValidatorsInCurrEpoch++
	}

	sdp.updateEpochStats(validator)
	return ownerData, nil
}

// loadDataForBlsKey will be called for each BLS key that took part in the consensus (no matter the shard ID) so the
// staking data can be recovered from the staking system smart contracts.
// The function will error if something went wrong. It does change the inner state of the called instance.
func (sdp *stakingDataProvider) loadDataForBlsKey(validator state.ValidatorInfoHandler) error {
	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	ownerData, err := sdp.getAndFillOwnerStats(validator)
	if err != nil {
		log.Debug("error computing rewards for bls key",
			"step", "get owner data",
			"key", hex.EncodeToString(validator.GetPublicKey()),
			"error", err)
		return err
	}
	ownerData.numEligible++

	return nil
}

// GetOwnersData returns all owner stats
func (sdp *stakingDataProvider) GetOwnersData() map[string]*epochStart.OwnerData {
	sdp.mutStakingData.RLock()
	defer sdp.mutStakingData.RUnlock()

	ret := make(map[string]*epochStart.OwnerData)
	for owner, ownerData := range sdp.cache {
		ret[owner] = &epochStart.OwnerData{
			NumActiveNodes: ownerData.numActiveNodes,
			NumStakedNodes: ownerData.numStakedNodes,
			TotalTopUp:     big.NewInt(0).SetBytes(ownerData.totalTopUp.Bytes()),
			TopUpPerNode:   big.NewInt(0).SetBytes(ownerData.topUpPerNode.Bytes()),
			AuctionList:    make([]state.ValidatorInfoHandler, len(ownerData.auctionList)),
			Qualified:      ownerData.qualified,
		}
		copy(ret[owner].AuctionList, ownerData.auctionList)
	}

	return ret
}

// GetBlsKeyOwner returns the owner's public key of the provided bls key
func (sdp *stakingDataProvider) GetBlsKeyOwner(blsKey []byte) (string, error) {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.ValidatorSCAddress,
			Arguments:  [][]byte{blsKey},
			CallValue:  big.NewInt(0),
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "getOwner",
	}

	vmOutput, err := sdp.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return "", err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return "", fmt.Errorf("%w, error: %v", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode)
	}
	data := vmOutput.ReturnData
	if len(data) != 1 {
		return "", fmt.Errorf("%w, getOwner function should have returned exactly one value: the owner address", epochStart.ErrExecutingSystemScCode)
	}

	return string(data[0]), nil
}

func (sdp *stakingDataProvider) fillOwnerData(owner string, validator state.ValidatorInfoHandler) (*ownerStats, error) {
	var err error
	ownerData, exists := sdp.cache[owner]
	if exists {
		updateOwnerData(ownerData, validator)
	} else {
		ownerData, err = sdp.getAndFillOwnerDataFromSC(owner, validator)
		if err != nil {
			return nil, err
		}
		sdp.cache[owner] = ownerData
	}

	return ownerData, nil
}

func updateOwnerData(ownerData *ownerStats, validator state.ValidatorInfoHandler) {
	if isInAuction(validator) {
		ownerData.numActiveNodes--
		ownerData.auctionList = append(ownerData.auctionList, validator.ShallowClone())
	}
}

func (sdp *stakingDataProvider) getAndFillOwnerDataFromSC(owner string, validator state.ValidatorInfoHandler) (*ownerStats, error) {
	ownerInfo, err := sdp.getOwnerInfoFromSC(owner)
	if err != nil {
		return nil, err
	}

	topUpPerNode := big.NewInt(0)
	numStakedNodes := ownerInfo.numStakedWaiting.Int64()
	if numStakedNodes == 0 {
		log.Debug("stakingDataProvider.fillOwnerData",
			"message", epochStart.ErrOwnerHasNoStakedNode,
			"owner", hex.EncodeToString([]byte(owner)),
			"validator", hex.EncodeToString(validator.GetPublicKey()),
		)
	} else {
		topUpPerNode = big.NewInt(0).Div(ownerInfo.topUpValue, ownerInfo.numStakedWaiting)
	}

	ownerData := &ownerStats{
		numEligible:          0,
		numStakedNodes:       numStakedNodes,
		numActiveNodes:       numStakedNodes,
		totalTopUp:           ownerInfo.topUpValue,
		topUpPerNode:         topUpPerNode,
		totalStaked:          ownerInfo.totalStakedValue,
		eligibleBaseStake:    big.NewInt(0).Set(sdp.minNodePrice),
		eligibleTopUpStake:   big.NewInt(0),
		eligibleTopUpPerNode: big.NewInt(0),
		qualified:            true,
	}
	err = sdp.checkAndFillOwnerValidatorAuctionData([]byte(owner), ownerData, validator)
	if err != nil {
		return nil, err
	}

	ownerData.blsKeys = make([][]byte, len(ownerInfo.blsKeys))
	copy(ownerData.blsKeys, ownerInfo.blsKeys)

	return ownerData, nil
}

func (sdp *stakingDataProvider) checkAndFillOwnerValidatorAuctionData(
	ownerPubKey []byte,
	ownerData *ownerStats,
	validator state.ValidatorInfoHandler,
) error {
	validatorInAuction := isInAuction(validator)
	if !validatorInAuction {
		return nil
	}
	if ownerData.numStakedNodes == 0 {
		return fmt.Errorf("stakingDataProvider.checkAndFillOwnerValidatorAuctionData for validator in auction error: %w, owner: %s, node: %s",
			epochStart.ErrOwnerHasNoStakedNode,
			hex.EncodeToString(ownerPubKey),
			hex.EncodeToString(validator.GetPublicKey()),
		)
	}
	if !sdp.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) {
		return fmt.Errorf("stakingDataProvider.checkAndFillOwnerValidatorAuctionData for validator in auction error: %w, owner: %s, node: %s",
			epochStart.ErrReceivedAuctionValidatorsBeforeStakingV4,
			hex.EncodeToString(ownerPubKey),
			hex.EncodeToString(validator.GetPublicKey()),
		)
	}

	ownerData.numActiveNodes -= 1
	ownerData.auctionList = []state.ValidatorInfoHandler{validator}

	return nil
}

func (sdp *stakingDataProvider) getOwnerInfoFromSC(owner string) (*ownerInfoSC, error) {
	ownerAddressBytes := []byte(owner)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  vm.EndOfEpochAddress,
			CallValue:   big.NewInt(0),
			GasProvided: math.MaxInt64,
			Arguments:   [][]byte{ownerAddressBytes},
		},
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "getTotalStakedTopUpStakedBlsKeys",
	}

	vmOutput, err := sdp.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("%w, error: %v message: %s", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode, vmOutput.ReturnMessage)
	}

	if len(vmOutput.ReturnData) < 3 {
		return nil, fmt.Errorf("%w, getTotalStakedTopUpStakedBlsKeys function should have at least three values", epochStart.ErrExecutingSystemScCode)
	}

	topUpValue := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
	totalStakedValue := big.NewInt(0).SetBytes(vmOutput.ReturnData[1])
	numStakedWaiting := big.NewInt(0).SetBytes(vmOutput.ReturnData[2])

	return &ownerInfoSC{
		topUpValue:       topUpValue,
		totalStakedValue: totalStakedValue,
		numStakedWaiting: numStakedWaiting,
		blsKeys:          vmOutput.ReturnData[3:],
	}, nil
}

// ComputeUnQualifiedNodes will compute which nodes are not qualified - do not have enough tokens to be validators
func (sdp *stakingDataProvider) ComputeUnQualifiedNodes(validatorsInfo state.ShardValidatorsInfoMapHandler) ([][]byte, map[string][][]byte, error) {
	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	mapOwnersKeys := make(map[string][][]byte)
	keysToUnStake := make([][]byte, 0)
	mapBLSKeyStatus, err := sdp.createMapBLSKeyStatus(validatorsInfo)
	if err != nil {
		return nil, nil, err
	}

	for ownerAddress, stakingInfo := range sdp.cache {
		maxQualified := big.NewInt(0).Div(stakingInfo.totalStaked, sdp.minNodePrice)
		if maxQualified.Int64() >= stakingInfo.numStakedNodes {
			continue
		}

		sortedKeys := sdp.arrangeBlsKeysByStatus(mapBLSKeyStatus, stakingInfo.blsKeys)

		numKeysToUnStake := stakingInfo.numStakedNodes - maxQualified.Int64()
		selectedKeys, numRemovedValidators := sdp.selectKeysToUnStake(sortedKeys, numKeysToUnStake)
		if len(selectedKeys) == 0 {
			continue
		}

		keysToUnStake = append(keysToUnStake, selectedKeys...)

		mapOwnersKeys[ownerAddress] = make([][]byte, len(selectedKeys))
		copy(mapOwnersKeys[ownerAddress], selectedKeys)

		stakingInfo.qualified = false
		sdp.numOfValidatorsInCurrEpoch -= uint32(numRemovedValidators)
	}

	return keysToUnStake, mapOwnersKeys, nil
}

func (sdp *stakingDataProvider) createMapBLSKeyStatus(validatorsInfo state.ShardValidatorsInfoMapHandler) (map[string]string, error) {
	mapBLSKeyStatus := make(map[string]string)
	for _, validator := range validatorsInfo.GetAllValidatorsInfo() {
		list := validator.GetList()
		pubKey := validator.GetPublicKey()

		if sdp.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step2Flag) && list == string(common.NewList) {
			return nil, fmt.Errorf("%w, bls key = %s",
				epochStart.ErrReceivedNewListNodeInStakingV4,
				hex.EncodeToString(pubKey),
			)
		}

		mapBLSKeyStatus[string(pubKey)] = list
	}

	return mapBLSKeyStatus, nil
}

func (sdp *stakingDataProvider) selectKeysToUnStake(sortedKeys map[string][][]byte, numToSelect int64) ([][]byte, int) {
	selectedKeys := make([][]byte, 0)
	newNodesList := sdp.getNewNodesList()

	newKeys := sortedKeys[newNodesList]
	if len(newKeys) > 0 {
		selectedKeys = append(selectedKeys, newKeys...)
	}

	if int64(len(selectedKeys)) >= numToSelect {
		return selectedKeys[:numToSelect], 0
	}

	waitingKeys := sortedKeys[string(common.WaitingList)]
	if len(waitingKeys) > 0 {
		selectedKeys = append(selectedKeys, waitingKeys...)
	}

	if int64(len(selectedKeys)) >= numToSelect {
		overFlowKeys := len(selectedKeys) - int(numToSelect)
		removedWaiting := len(waitingKeys) - overFlowKeys
		return selectedKeys[:numToSelect], removedWaiting
	}

	eligibleKeys := sortedKeys[string(common.EligibleList)]
	if len(eligibleKeys) > 0 {
		selectedKeys = append(selectedKeys, eligibleKeys...)
	}

	if int64(len(selectedKeys)) >= numToSelect {
		overFlowKeys := len(selectedKeys) - int(numToSelect)
		removedEligible := len(eligibleKeys) - overFlowKeys
		return selectedKeys[:numToSelect], removedEligible + len(waitingKeys)
	}

	return selectedKeys, len(eligibleKeys) + len(waitingKeys)
}

func (sdp *stakingDataProvider) arrangeBlsKeysByStatus(mapBlsKeyStatus map[string]string, blsKeys [][]byte) map[string][][]byte {
	sortedKeys := make(map[string][][]byte)
	newNodesList := sdp.getNewNodesList()

	for _, blsKey := range blsKeys {
		blsKeyStatus, found := mapBlsKeyStatus[string(blsKey)]
		if !found {
			sortedKeys[newNodesList] = append(sortedKeys[newNodesList], blsKey)
			continue
		}

		sortedKeys[blsKeyStatus] = append(sortedKeys[blsKeyStatus], blsKey)
	}

	return sortedKeys
}

func (sdp *stakingDataProvider) getNewNodesList() string {
	newNodesList := string(common.NewList)
	if sdp.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step2Flag) {
		newNodesList = string(common.AuctionList)
	}

	return newNodesList
}

// GetNumOfValidatorsInCurrentEpoch returns the number of validators(eligible + waiting) in current epoch
func (sdp *stakingDataProvider) GetNumOfValidatorsInCurrentEpoch() uint32 {
	sdp.mutStakingData.RLock()
	defer sdp.mutStakingData.RUnlock()

	return sdp.numOfValidatorsInCurrEpoch
}

func (sdp *stakingDataProvider) updateEpochStats(validator state.ValidatorInfoHandler) {
	validatorCurrentList := common.PeerType(validator.GetList())
	shardID := validator.GetShardId()

	if validatorCurrentList == common.EligibleList {
		sdp.validatorStatsInEpoch.Eligible[shardID]++
		return
	}

	if validatorCurrentList == common.WaitingList {
		sdp.validatorStatsInEpoch.Waiting[shardID]++
		return
	}

	validatorPreviousList := common.PeerType(validator.GetPreviousList())
	if sdp.isValidatorLeaving(validatorCurrentList, validatorPreviousList) {
		sdp.validatorStatsInEpoch.Leaving[shardID]++
	}
}

func (sdp *stakingDataProvider) isValidatorLeaving(validatorCurrentList, validatorPreviousList common.PeerType) bool {
	if validatorCurrentList != common.LeavingList {
		return false
	}

	// If no previous list is set, means that staking v4 is not activated or node is leaving right before activation
	// and this node will be considered as eligible by the nodes coordinator with old code.
	// Otherwise, it will have it set, and we should check its previous list in the current epoch
	return len(validatorPreviousList) == 0 || validatorPreviousList == common.EligibleList || validatorPreviousList == common.WaitingList
}

// GetCurrentEpochValidatorStats returns the current epoch validator stats
func (sdp *stakingDataProvider) GetCurrentEpochValidatorStats() epochStart.ValidatorStatsInEpoch {
	sdp.mutStakingData.RLock()
	defer sdp.mutStakingData.RUnlock()

	return sdp.validatorStatsInEpoch
}

// IsInterfaceNil return true if underlying object is nil
func (sdp *stakingDataProvider) IsInterfaceNil() bool {
	return sdp == nil
}
