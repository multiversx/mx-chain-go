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

type rewardsStakingProvider struct {
	mutCache sync.Mutex
	cache    map[string]*ownerStats
	systemVM vmcommon.VMExecutionHandler
}

// NewRewardsStakingProvider will create a new instance of a staking data provider able to aid in the final rewards
// computation as this will retrieve the staking data from the system VM
func NewRewardsStakingProvider(
	systemVM vmcommon.VMExecutionHandler,
) (*rewardsStakingProvider, error) {
	//TODO make vmcommon.VMExecutionHandler implement NilInterfaceChecker
	if check.IfNilReflect(systemVM) {
		return nil, epochStart.ErrNilSystemVmInstance
	}

	rsp := &rewardsStakingProvider{
		systemVM: systemVM,
	}
	rsp.Clean()

	return rsp, nil
}

// Clean will reset the inner state of the called instance
func (rsp *rewardsStakingProvider) Clean() {
	rsp.mutCache.Lock()
	rsp.cache = make(map[string]*ownerStats)
	rsp.mutCache.Unlock()
}

// ComputeRewardsForBlsKey will be called for each BLS key that took part in the consensus (no matter the shard ID)
// The function will error if something went wrong. It does change the inner state of the called instance.
func (rsp *rewardsStakingProvider) ComputeRewardsForBlsKey(blsKey []byte) error {
	owner, err := rsp.getBlsKeyOwnerAsHex(blsKey)
	if err != nil {
		log.Debug("error computing rewards for bls key", "step", "get owner from bls", "key", hex.EncodeToString(blsKey), "error", err)
		return err
	}

	rsp.mutCache.Lock()
	defer rsp.mutCache.Unlock()

	ownerData, err := rsp.getOwnerData(owner)
	if err != nil {
		log.Debug("error computing rewards for bls key", "step", "get owner data", "key", hex.EncodeToString(blsKey), "error", err)
		return err
	}
	ownerData.numEligible++

	return nil
}

func (rsp *rewardsStakingProvider) getBlsKeyOwnerAsHex(blsKey []byte) (string, error) {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.AuctionSCAddress,
			Arguments:  [][]byte{blsKey},
			CallValue:  big.NewInt(0),
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "getOwner",
	}

	vmOutput, err := rsp.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return "", err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return "", fmt.Errorf("%w, error: %v", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode)
	}
	data := vmOutput.ReturnData
	if len(data) < 1 {
		return "", fmt.Errorf("%w, error: missing owner address", epochStart.ErrExecutingSystemScCode)
	}

	return string(data[0]), nil
}

func (rsp *rewardsStakingProvider) getOwnerData(owner string) (*ownerStats, error) {
	ownerData, exists := rsp.cache[owner]
	if exists {
		return ownerData, nil
	}

	return rsp.loadOwnerData(owner)
}

func (rsp *rewardsStakingProvider) loadOwnerData(owner string) (*ownerStats, error) {
	topUpValue, err := rsp.loadTopUpValue(owner)
	if err != nil {
		return nil, err
	}

	ownerData := &ownerStats{
		numEligible: 0,
		topUpValue:  topUpValue,
	}
	rsp.cache[owner] = ownerData

	return ownerData, nil
}

func (rsp *rewardsStakingProvider) loadTopUpValue(owner string) (*big.Int, error) {
	ownerBytes, err := hex.DecodeString(owner)
	if err != nil {
		return nil, err
	}

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  ownerBytes,
			CallValue:   big.NewInt(0),
			GasProvided: math.MaxUint64,
		},
		RecipientAddr: vm.AuctionSCAddress,
		Function:      "getTopUp",
	}

	vmOutput, err := rsp.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("%w, error: %v", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode)
	}
	topUpBytes := vmOutput.ReturnData
	if len(topUpBytes) < 1 {
		return nil, fmt.Errorf("%w, error: missing top up value data bytes", epochStart.ErrExecutingSystemScCode)
	}

	topUpValue, ok := big.NewInt(0).SetString(string(topUpBytes[0]), conversionBase)
	if !ok {
		return nil, fmt.Errorf("%w, error: topUp string returned is not a number", epochStart.ErrExecutingSystemScCode)
	}

	return topUpValue, nil
}

// IsInterfaceNil return true if underlying object is nil
func (rsp *rewardsStakingProvider) IsInterfaceNil() bool {
	return rsp == nil
}
