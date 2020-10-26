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
	numEligible int
	topUpValue  *big.Int
}

type stakingDataProvider struct {
	mutCache sync.Mutex
	cache    map[string]*ownerStats
	systemVM vmcommon.VMExecutionHandler
}

// NewStakingDataProvider will create a new instance of a staking data provider able to aid in the final rewards
// computation as this will retrieve the staking data from the system VM
func NewStakingDataProvider(
	systemVM vmcommon.VMExecutionHandler,
) (*stakingDataProvider, error) {
	//TODO make vmcommon.VMExecutionHandler implement NilInterfaceChecker
	if check.IfNilReflect(systemVM) {
		return nil, epochStart.ErrNilSystemVmInstance
	}

	sdp := &stakingDataProvider{
		systemVM: systemVM,
	}
	sdp.Clean()

	return sdp, nil
}

// Clean will reset the inner state of the called instance
func (sdp *stakingDataProvider) Clean() {
	sdp.mutCache.Lock()
	sdp.cache = make(map[string]*ownerStats)
	sdp.mutCache.Unlock()
}

// PrepareDataForBlsKey will be called for each BLS key that took part in the consensus (no matter the shard ID) so the
// staking data can be recovered from the staking system smart contracts.
// The function will error if something went wrong. It does change the inner state of the called instance.
func (sdp *stakingDataProvider) PrepareDataForBlsKey(blsKey []byte) error {
	owner, err := sdp.getBlsKeyOwnerAsHex(blsKey)
	if err != nil {
		log.Debug("error computing rewards for bls key", "step", "get owner from bls", "key", hex.EncodeToString(blsKey), "error", err)
		return err
	}

	sdp.mutCache.Lock()
	defer sdp.mutCache.Unlock()

	ownerData, err := sdp.getValidatorData(owner)
	if err != nil {
		log.Debug("error computing rewards for bls key", "step", "get owner data", "key", hex.EncodeToString(blsKey), "error", err)
		return err
	}
	ownerData.numEligible++

	return nil
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

	ownerData := &ownerStats{
		numEligible: 0,
		topUpValue:  topUpValue,
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

// IsInterfaceNil return true if underlying object is nil
func (sdp *stakingDataProvider) IsInterfaceNil() bool {
	return sdp == nil
}
