package metachain

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const conversionBase = 10

type ownerStats struct {
	numEligible        int
	numStakedNodes     int
	topUpValue         *big.Int
	totalStaked        *big.Int
	eligibleBaseStake  *big.Int
	eligibleTopUpStake *big.Int
}

type stakingDataProvider struct {
	mutStakingData          sync.Mutex
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
	//TODO make vmcommon.VMExecutionHandler implement NilInterfaceChecker
	if check.IfNilReflect(systemVM) {
		return nil, epochStart.ErrNilSystemVmInstance
	}

	nodePrice, ok := big.NewInt(0).SetString(minNodePrice, 10)
	if !ok || nodePrice.Cmp(big.NewInt(0)) < 0 {
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

// PrepareStakingData prepares the staking data for the given map of node keys per shard
func (sdp *stakingDataProvider) PrepareStakingData(keys map[uint32][][]byte) error {
	sdp.Clean()

	for _, keysList := range keys {
		for _, blsKey := range keysList {
			err := sdp.prepareDataForBlsKey(blsKey)
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
		ownerStakePerNode := big.NewInt(0).Div(owner.totalStaked, big.NewInt(int64(owner.numStakedNodes)))
		ownerEligibleStake := big.NewInt(0).Mul(ownerStakePerNode, ownerEligibleNodes)

		owner.eligibleBaseStake = big.NewInt(0).Mul(ownerEligibleNodes, sdp.minNodePrice)
		owner.eligibleTopUpStake = big.NewInt(0).Sub(ownerEligibleStake, owner.eligibleBaseStake)

		totalEligibleStake.Add(totalEligibleStake, ownerEligibleStake)
		totalEligibleTopUpStake.Add(totalEligibleTopUpStake, owner.eligibleTopUpStake)
	}

	sdp.totalEligibleTopUpStake = totalEligibleTopUpStake
	sdp.totalEligibleStake = totalEligibleStake
}

// prepareDataForBlsKey will be called for each BLS key that took part in the consensus (no matter the shard ID) so the
// staking data can be recovered from the staking system smart contracts.
// The function will error if something went wrong. It does change the inner state of the called instance.
func (sdp *stakingDataProvider) prepareDataForBlsKey(blsKey []byte) error {
	owner, err := sdp.getBlsKeyOwnerAsHex(blsKey)
	if err != nil {
		log.Debug("error computing rewards for bls key", "step", "get owner from bls", "key", hex.EncodeToString(blsKey), "error", err)
		return err
	}

	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	ownerData, err := sdp.getValidatorData(owner)
	if err != nil {
		log.Debug("error computing rewards for bls key", "step", "get owner data", "key", hex.EncodeToString(blsKey), "error", err)
		return err
	}
	ownerData.numEligible++

	return nil
}

// GetTotalStakeEligibleNodes returns the total staked amount of the current epoch eligible nodes
func (sdp *stakingDataProvider) GetTotalStakeEligibleNodes() *big.Int {
	// TODO: implement this
	return big.NewInt(0)
}

func (sdp *stakingDataProvider) getBlsKeyOwnerAsHex(blsKey []byte) (string, error) {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.AuctionSCAddress,
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
	topUpValue, err := sdp.getTopUpValue(validatorAddress)
	if err != nil {
		return nil, err
	}

	totalStakedValue, err := sdp.getTotalStaked(validatorAddress)
	if err != nil {
		return nil, err
	}

	ownerBaseNodesStake := big.NewInt(0).Sub(totalStakedValue, topUpValue)
	numStakedNodes := big.NewInt(0).Div(ownerBaseNodesStake, sdp.minNodePrice)
	ownerData := &ownerStats{
		numEligible:    0,
		numStakedNodes: int(numStakedNodes.Int64()),
		topUpValue:     topUpValue,
		totalStaked:    totalStakedValue,
	}
	sdp.cache[validatorAddress] = ownerData

	return ownerData, nil
}

func (sdp *stakingDataProvider) getTopUpValue(validatorAddress string) (*big.Int, error) {
	validatorAddressBytes, err := hex.DecodeString(validatorAddress)
	if err != nil {
		return nil, err
	}

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  validatorAddressBytes,
			CallValue:   big.NewInt(0),
			GasProvided: math.MaxUint64,
		},
		RecipientAddr: vm.AuctionSCAddress,
		Function:      "getTopUp",
	}

	vmOutput, err := sdp.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("%w, error: %v", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode)
	}
	topUpBytes := vmOutput.ReturnData
	if len(topUpBytes) != 1 {
		return nil, fmt.Errorf("%w, getTopUp function should have returned exactly one value: the top up value", epochStart.ErrExecutingSystemScCode)
	}

	topUpValue, ok := big.NewInt(0).SetString(string(topUpBytes[0]), conversionBase)
	if !ok {
		return nil, fmt.Errorf("%w, error: topUp string returned is not a number", epochStart.ErrExecutingSystemScCode)
	}

	return topUpValue, nil
}

func (sdp *stakingDataProvider) getTotalStaked(validatorAddress string) (*big.Int, error) {
	validatorAddressBytes, err := hex.DecodeString(validatorAddress)
	if err != nil {
		return nil, err
	}

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  validatorAddressBytes,
			CallValue:   big.NewInt(0),
			GasProvided: math.MaxUint64,
		},
		RecipientAddr: vm.AuctionSCAddress,
		Function:      "getTotalStaked",
	}

	vmOutput, err := sdp.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("%w, error: %v", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode)
	}
	totalStakedBytes := vmOutput.ReturnData
	if len(totalStakedBytes) != 1 {
		return nil, fmt.Errorf("%w, getTotalStaked function should have returned exactly one value: the total staked value", epochStart.ErrExecutingSystemScCode)
	}

	totalStakedValue, ok := big.NewInt(0).SetString(string(totalStakedBytes[0]), conversionBase)
	if !ok {
		return nil, fmt.Errorf("%w, error: totalStaked string returned is not a number", epochStart.ErrExecutingSystemScCode)
	}

	return totalStakedValue, nil
}

// IsInterfaceNil return true if underlying object is nil
func (sdp *stakingDataProvider) IsInterfaceNil() bool {
	return sdp == nil
}
