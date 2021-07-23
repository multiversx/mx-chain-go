package metachain

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type ownerStats struct {
	numEligible        int
	numStakedNodes     int64
	topUpValue         *big.Int
	totalStaked        *big.Int
	eligibleBaseStake  *big.Int
	eligibleTopUpStake *big.Int
	topUpPerNode       *big.Int
	blsKeys            [][]byte
}

type stakingDataProvider struct {
	mutStakingData          sync.RWMutex
	cache                   map[string]*ownerStats
	systemVM                vmcommon.VMExecutionHandler
	totalEligibleStake      *big.Int
	totalEligibleTopUpStake *big.Int
	minNodePrice            *big.Int
}

// NewStakingDataProvider will create a new instance of a staking data provider able to aid in the final rewards
// computation as this will retrieve the staking data from the system VM
func NewStakingDataProvider(
	systemVM vmcommon.VMExecutionHandler,
	minNodePrice string,
) (*stakingDataProvider, error) {
	if check.IfNil(systemVM) {
		return nil, epochStart.ErrNilSystemVmInstance
	}

	nodePrice, ok := big.NewInt(0).SetString(minNodePrice, 10)
	if !ok || nodePrice.Cmp(big.NewInt(0)) <= 0 {
		return nil, epochStart.ErrInvalidMinNodePrice
	}

	sdp := &stakingDataProvider{
		systemVM:                systemVM,
		cache:                   make(map[string]*ownerStats),
		minNodePrice:            nodePrice,
		totalEligibleStake:      big.NewInt(0),
		totalEligibleTopUpStake: big.NewInt(0),
	}

	return sdp, nil
}

// Clean will reset the inner state of the called instance
func (sdp *stakingDataProvider) Clean() {
	sdp.mutStakingData.Lock()
	sdp.cache = make(map[string]*ownerStats)
	sdp.totalEligibleStake.SetInt64(0)
	sdp.totalEligibleTopUpStake.SetInt64(0)
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
//current epoch eligible nodes
// This value is populated by a previous call to PrepareStakingData (done for epoch start)
func (sdp *stakingDataProvider) GetTotalTopUpStakeEligibleNodes() *big.Int {
	sdp.mutStakingData.RLock()
	defer sdp.mutStakingData.RUnlock()

	return big.NewInt(0).Set(sdp.totalEligibleTopUpStake)
}

// GetNodeStakedTopUp returns the owner of provided bls key staking stats for the current epoch
func (sdp *stakingDataProvider) GetNodeStakedTopUp(blsKey []byte) (*big.Int, error) {
	owner, err := sdp.getBlsKeyOwner(blsKey)
	if err != nil {
		log.Debug("GetOwnerStakingStats", "key", hex.EncodeToString(blsKey), "error", err)
		return nil, err
	}

	ownerInfo, ok := sdp.cache[owner]
	if !ok {
		return nil, epochStart.ErrOwnerDoesntHaveEligibleNodesInEpoch
	}

	return ownerInfo.topUpPerNode, nil
}

// PrepareStakingDataForRewards prepares the staking data for the given map of node keys per shard
func (sdp *stakingDataProvider) PrepareStakingDataForRewards(keys map[uint32][][]byte) error {
	sdp.Clean()

	for _, keysList := range keys {
		for _, blsKey := range keysList {
			err := sdp.loadDataForBlsKey(blsKey)
			if err != nil {
				return err
			}
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

		owner.topUpPerNode = big.NewInt(0).Div(owner.eligibleTopUpStake, ownerEligibleNodes)
	}

	sdp.totalEligibleTopUpStake = totalEligibleTopUpStake
	sdp.totalEligibleStake = totalEligibleStake
}

// FillValidatorInfo will fill the validator info for the bls key if it was not already filled
func (sdp *stakingDataProvider) FillValidatorInfo(blsKey []byte) error {
	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	_, err := sdp.getAndFillOwnerStatsFromSC(blsKey)
	return err
}

func (sdp *stakingDataProvider) getAndFillOwnerStatsFromSC(blsKey []byte) (*ownerStats, error) {
	owner, err := sdp.getBlsKeyOwner(blsKey)
	if err != nil {
		log.Debug("error fill owner stats", "step", "get owner from bls", "key", hex.EncodeToString(blsKey), "error", err)
		return nil, err
	}

	ownerData, err := sdp.getValidatorData(owner)
	if err != nil {
		log.Debug("error fill owner stats", "step", "get owner data", "key", hex.EncodeToString(blsKey), "owner", hex.EncodeToString([]byte(owner)), "error", err)
		return nil, err
	}

	return ownerData, nil
}

// loadDataForBlsKey will be called for each BLS key that took part in the consensus (no matter the shard ID) so the
// staking data can be recovered from the staking system smart contracts.
// The function will error if something went wrong. It does change the inner state of the called instance.
func (sdp *stakingDataProvider) loadDataForBlsKey(blsKey []byte) error {
	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	ownerData, err := sdp.getAndFillOwnerStatsFromSC(blsKey)
	if err != nil {
		log.Debug("error computing rewards for bls key", "step", "get owner data", "key", hex.EncodeToString(blsKey), "error", err)
		return err
	}
	ownerData.numEligible++

	return nil
}

func (sdp *stakingDataProvider) getBlsKeyOwner(blsKey []byte) (string, error) {
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

func (sdp *stakingDataProvider) getValidatorData(validatorAddress string) (*ownerStats, error) {
	ownerData, exists := sdp.cache[validatorAddress]
	if exists {
		return ownerData, nil
	}

	return sdp.getValidatorDataFromStakingSC(validatorAddress)
}

func (sdp *stakingDataProvider) getValidatorDataFromStakingSC(validatorAddress string) (*ownerStats, error) {
	topUpValue, totalStakedValue, numStakedWaiting, blsKeys, err := sdp.getValidatorInfoFromSC(validatorAddress)
	if err != nil {
		return nil, err
	}

	ownerData := &ownerStats{
		numEligible:        0,
		numStakedNodes:     numStakedWaiting.Int64(),
		topUpValue:         topUpValue,
		totalStaked:        totalStakedValue,
		eligibleBaseStake:  big.NewInt(0).Set(sdp.minNodePrice),
		eligibleTopUpStake: big.NewInt(0),
		topUpPerNode:       big.NewInt(0),
	}

	ownerData.blsKeys = make([][]byte, len(blsKeys))
	copy(ownerData.blsKeys, blsKeys)

	sdp.cache[validatorAddress] = ownerData

	return ownerData, nil
}

func (sdp *stakingDataProvider) getValidatorInfoFromSC(validatorAddress string) (*big.Int, *big.Int, *big.Int, [][]byte, error) {
	validatorAddressBytes := []byte(validatorAddress)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  vm.EndOfEpochAddress,
			CallValue:   big.NewInt(0),
			GasProvided: math.MaxUint64,
			Arguments:   [][]byte{validatorAddressBytes},
		},
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "getTotalStakedTopUpStakedBlsKeys",
	}

	vmOutput, err := sdp.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, nil, nil, nil, fmt.Errorf("%w, error: %v message: %s", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode, vmOutput.ReturnMessage)
	}

	if len(vmOutput.ReturnData) < 3 {
		return nil, nil, nil, nil, fmt.Errorf("%w, getTotalStakedTopUpStakedBlsKeys function should have at least three values", epochStart.ErrExecutingSystemScCode)
	}

	topUpValue := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
	totalStakedValue := big.NewInt(0).SetBytes(vmOutput.ReturnData[1])
	numStakedWaiting := big.NewInt(0).SetBytes(vmOutput.ReturnData[2])

	return topUpValue, totalStakedValue, numStakedWaiting, vmOutput.ReturnData[3:], nil
}

// ComputeUnQualifiedNodes will compute which nodes are not qualified - do not have enough tokens to be validators
func (sdp *stakingDataProvider) ComputeUnQualifiedNodes(validatorInfos map[uint32][]*state.ValidatorInfo) ([][]byte, map[string][][]byte, error) {
	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	mapOwnersKeys := make(map[string][][]byte)
	keysToUnStake := make([][]byte, 0)
	mapBLSKeyStatus := createMapBLSKeyStatus(validatorInfos)
	for ownerAddress, stakingInfo := range sdp.cache {
		maxQualified := big.NewInt(0).Div(stakingInfo.totalStaked, sdp.minNodePrice)
		if maxQualified.Int64() >= stakingInfo.numStakedNodes {
			continue
		}

		sortedKeys := arrangeBlsKeysByStatus(mapBLSKeyStatus, stakingInfo.blsKeys)

		numKeysToUnStake := stakingInfo.numStakedNodes - maxQualified.Int64()
		selectedKeys := selectKeysToUnStake(sortedKeys, numKeysToUnStake)
		if len(selectedKeys) == 0 {
			continue
		}

		keysToUnStake = append(keysToUnStake, selectedKeys...)

		mapOwnersKeys[ownerAddress] = make([][]byte, len(selectedKeys))
		copy(mapOwnersKeys[ownerAddress], selectedKeys)
	}

	return keysToUnStake, mapOwnersKeys, nil
}

func createMapBLSKeyStatus(validatorInfos map[uint32][]*state.ValidatorInfo) map[string]string {
	mapBLSKeyStatus := make(map[string]string)
	for _, validatorsInfoSlice := range validatorInfos {
		for _, validatorInfo := range validatorsInfoSlice {
			mapBLSKeyStatus[string(validatorInfo.PublicKey)] = validatorInfo.List
		}
	}

	return mapBLSKeyStatus
}

func selectKeysToUnStake(sortedKeys map[string][][]byte, numToSelect int64) [][]byte {
	selectedKeys := make([][]byte, 0)
	newKeys := sortedKeys[string(common.NewList)]
	if len(newKeys) > 0 {
		selectedKeys = append(selectedKeys, newKeys...)
	}

	if int64(len(selectedKeys)) >= numToSelect {
		return selectedKeys[:numToSelect]
	}

	waitingKeys := sortedKeys[string(common.WaitingList)]
	if len(waitingKeys) > 0 {
		selectedKeys = append(selectedKeys, waitingKeys...)
	}

	if int64(len(selectedKeys)) >= numToSelect {
		return selectedKeys[:numToSelect]
	}

	eligibleKeys := sortedKeys[string(common.EligibleList)]
	if len(eligibleKeys) > 0 {
		selectedKeys = append(selectedKeys, eligibleKeys...)
	}

	if int64(len(selectedKeys)) >= numToSelect {
		return selectedKeys[:numToSelect]
	}

	return selectedKeys
}

func arrangeBlsKeysByStatus(mapBlsKeyStatus map[string]string, blsKeys [][]byte) map[string][][]byte {
	sortedKeys := make(map[string][][]byte)
	for _, blsKey := range blsKeys {
		blsKeyStatus, ok := mapBlsKeyStatus[string(blsKey)]
		if !ok {
			sortedKeys[string(common.NewList)] = append(sortedKeys[string(common.NewList)], blsKey)
			continue
		}

		sortedKeys[blsKeyStatus] = append(sortedKeys[blsKeyStatus], blsKey)
	}

	return sortedKeys
}

// IsInterfaceNil return true if underlying object is nil
func (sdp *stakingDataProvider) IsInterfaceNil() bool {
	return sdp == nil
}
