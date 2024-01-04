package metachain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stakingV4Step1EnableEpoch = 444
const stakingV4Step2EnableEpoch = 445

func createStakingDataProviderArgs() StakingDataProviderArgs {
	return StakingDataProviderArgs{
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{},
		SystemVM:            &mock.VMExecutionHandlerStub{},
		MinNodePrice:        "2500",
	}
}

func TestNewStakingDataProvider_NilInputPointersShouldErr(t *testing.T) {
	t.Parallel()

	t.Run("nil system vm", func(t *testing.T) {
		args := createStakingDataProviderArgs()
		args.SystemVM = nil
		sdp, err := NewStakingDataProvider(args)
		assert.True(t, check.IfNil(sdp))
		assert.Equal(t, epochStart.ErrNilSystemVmInstance, err)
	})

	t.Run("nil epoch notifier", func(t *testing.T) {
		args := createStakingDataProviderArgs()
		args.EnableEpochsHandler = nil
		sdp, err := NewStakingDataProvider(args)
		assert.True(t, check.IfNil(sdp))
		assert.Equal(t, vm.ErrNilEnableEpochsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		args := createStakingDataProviderArgs()
		sdp, err := NewStakingDataProvider(args)
		assert.False(t, check.IfNil(sdp))
		assert.Nil(t, err)
	})
}

func TestStakingDataProvider_PrepareDataForBlsKeyGetBlsKeyOwnerErrorsShouldErr(t *testing.T) {
	t.Parallel()

	numCall := 0
	expectedErr := errors.New("expected error")
	args := createStakingDataProviderArgs()
	args.SystemVM = &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			numCall++
			if numCall == 1 {
				return nil, expectedErr
			}
			if numCall == 2 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.UserError,
				}, nil
			}
			if numCall == 3 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.Ok,
				}, nil
			}

			return nil, nil
		},
	}
	sdp, _ := NewStakingDataProvider(args)

	err := sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	assert.Equal(t, expectedErr, err)

	err = sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), vmcommon.UserError.String()))

	err = sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "returned exactly one value: the owner address"))
}

func TestStakingDataProvider_PrepareDataForBlsKeyLoadOwnerDataErrorsShouldErr(t *testing.T) {
	t.Parallel()

	numCall := 0
	owner := []byte("owner")
	expectedErr := errors.New("expected error")
	args := createStakingDataProviderArgs()
	args.SystemVM = &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			if input.Function == "getOwner" {
				return &vmcommon.VMOutput{
					ReturnData: [][]byte{[]byte(hex.EncodeToString(owner))},
				}, nil
			}

			numCall++
			if numCall == 1 {
				return nil, expectedErr
			}
			if numCall == 2 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.UserError,
				}, nil
			}
			if numCall == 3 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.Ok,
				}, nil
			}
			return nil, nil
		},
	}
	sdp, _ := NewStakingDataProvider(args)

	err := sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	assert.Equal(t, expectedErr, err)

	err = sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), vmcommon.UserError.String()))

	err = sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "getTotalStakedTopUpStakedBlsKeys function should have at least three values"))
}

func TestStakingDataProvider_PrepareDataForBlsKeyFromSCShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	basePrice := big.NewInt(100000)
	stakeVal := big.NewInt(0).Add(topUpVal, basePrice)
	numRunContractCalls := 0

	sdp := createStakingDataProviderWithMockArgs(t, owner, topUpVal, stakeVal, &numRunContractCalls)

	err := sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	assert.Nil(t, err)
	assert.Equal(t, 2, numRunContractCalls)
	ownerData := sdp.GetFromCache(owner)
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.totalTopUp)
	assert.Equal(t, 1, ownerData.numEligible)
}

func TestStakingDataProvider_PrepareDataForBlsKeyCachedResponseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	basePrice := big.NewInt(100000)
	stakeVal := big.NewInt(0).Add(topUpVal, basePrice)
	numRunContractCalls := 0

	sdp := createStakingDataProviderWithMockArgs(t, owner, topUpVal, stakeVal, &numRunContractCalls)

	err := sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	assert.Nil(t, err)

	err = sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: []byte("bls key2")})
	assert.Nil(t, err)

	assert.Equal(t, 3, numRunContractCalls)
	ownerData := sdp.GetFromCache(owner)
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.totalTopUp)
	assert.Equal(t, 2, ownerData.numEligible)
}

func TestStakingDataProvider_PrepareDataForBlsKeyWithRealSystemVmShouldWork(t *testing.T) {
	t.Parallel()

	owner := append([]byte("owner"), bytes.Repeat([]byte{1}, 27)...)
	topUpVal := big.NewInt(828743)
	blsKey := []byte("bls key")

	sdp := createStakingDataProviderWithRealArgs(t, owner, blsKey, topUpVal)
	err := sdp.loadDataForBlsKey(&state.ValidatorInfo{PublicKey: blsKey})
	assert.Nil(t, err)
	ownerData := sdp.GetFromCache(owner)
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.totalTopUp)
	assert.Equal(t, 1, ownerData.numEligible)
}

func TestStakingDataProvider_ComputeUnQualifiedNodes(t *testing.T) {
	nbShards := uint32(3)
	nbEligible := make(map[uint32]uint32)
	nbWaiting := make(map[uint32]uint32)
	nbLeaving := make(map[uint32]uint32)
	nbInactive := make(map[uint32]uint32)

	nbEligible[core.MetachainShardId] = 7
	nbWaiting[core.MetachainShardId] = 1
	nbLeaving[core.MetachainShardId] = 0
	nbInactive[core.MetachainShardId] = 0

	nbEligible[0] = 6
	nbWaiting[0] = 2
	nbLeaving[0] = 1
	nbInactive[0] = 0

	nbEligible[1] = 7
	nbWaiting[1] = 1
	nbLeaving[1] = 1
	nbInactive[1] = 1

	nbEligible[2] = 7
	nbWaiting[2] = 2
	nbLeaving[2] = 0
	nbInactive[2] = 0

	valInfo := createValidatorsInfo(nbShards, nbEligible, nbWaiting, nbLeaving, nbInactive)
	sdp := createStakingDataProviderAndUpdateCache(t, valInfo, big.NewInt(0))
	require.NotNil(t, sdp)

	keysToUnStake, ownersWithNotEnoughFunds, err := sdp.ComputeUnQualifiedNodes(valInfo)
	require.Nil(t, err)
	require.Zero(t, len(keysToUnStake))
	require.Zero(t, len(ownersWithNotEnoughFunds))
}

func TestStakingDataProvider_ComputeUnQualifiedNodesWithStakingV4ReceivedNewListNode(t *testing.T) {
	v0 := &state.ValidatorInfo{
		PublicKey:     []byte("blsKey0"),
		List:          string(common.EligibleList),
		RewardAddress: []byte("address0"),
	}
	v1 := &state.ValidatorInfo{
		PublicKey:     []byte("blsKey1"),
		List:          string(common.NewList),
		RewardAddress: []byte("address0"),
	}
	v2 := &state.ValidatorInfo{
		PublicKey:     []byte("blsKey2"),
		List:          string(common.AuctionList),
		RewardAddress: []byte("address1"),
	}

	valInfo := state.NewShardValidatorsInfoMap()
	_ = valInfo.Add(v0)
	_ = valInfo.Add(v1)
	_ = valInfo.Add(v2)

	sdp := createStakingDataProviderAndUpdateCache(t, valInfo, big.NewInt(0))
	sdp.enableEpochsHandler = &testscommon.EnableEpochsHandlerStub{IsStakingV4Step2FlagEnabledField: true}

	keysToUnStake, ownersWithNotEnoughFunds, err := sdp.ComputeUnQualifiedNodes(valInfo)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), epochStart.ErrReceivedNewListNodeInStakingV4.Error()))
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(v1.PublicKey)))
	require.Empty(t, keysToUnStake)
	require.Empty(t, ownersWithNotEnoughFunds)
}

func TestStakingDataProvider_ComputeUnQualifiedNodesWithOwnerNotEnoughFunds(t *testing.T) {
	nbShards := uint32(3)
	nbEligible := make(map[uint32]uint32)
	nbWaiting := make(map[uint32]uint32)
	nbLeaving := make(map[uint32]uint32)
	nbInactive := make(map[uint32]uint32)

	nbEligible[core.MetachainShardId] = 1
	nbWaiting[core.MetachainShardId] = 1
	nbLeaving[core.MetachainShardId] = 0
	nbInactive[core.MetachainShardId] = 0

	nbEligible[0] = 1
	nbWaiting[0] = 1
	nbLeaving[0] = 1
	nbInactive[0] = 1

	valInfo := createValidatorsInfo(nbShards, nbEligible, nbWaiting, nbLeaving, nbInactive)
	sdp := createStakingDataProviderAndUpdateCache(t, valInfo, big.NewInt(200))
	require.NotNil(t, sdp)

	keyOfAddressWaitingShard0 := "addresswaiting00"
	sdp.cache[keyOfAddressWaitingShard0].blsKeys = append(sdp.cache[keyOfAddressWaitingShard0].blsKeys, []byte("newKey"))
	sdp.cache[keyOfAddressWaitingShard0].totalStaked = big.NewInt(400)

	for id, val := range sdp.cache {
		fmt.Printf("id: %s, num bls keys: %d\n", id, len(val.blsKeys))
	}

	keysToUnStake, ownersWithNotEnoughFunds, err := sdp.ComputeUnQualifiedNodes(valInfo)
	require.Nil(t, err)
	require.Equal(t, 1, len(keysToUnStake))
	require.Equal(t, 1, len(ownersWithNotEnoughFunds))
}

func TestStakingDataProvider_ComputeUnQualifiedNodesWithOwnerNotEnoughFundsWithStakingV4(t *testing.T) {
	owner := "address0"
	v0 := &state.ValidatorInfo{
		PublicKey:     []byte("blsKey0"),
		List:          string(common.EligibleList),
		RewardAddress: []byte(owner),
	}
	v1 := &state.ValidatorInfo{
		PublicKey:     []byte("blsKey1"),
		List:          string(common.AuctionList),
		RewardAddress: []byte(owner),
	}

	valInfo := state.NewShardValidatorsInfoMap()
	_ = valInfo.Add(v0)
	_ = valInfo.Add(v1)

	sdp := createStakingDataProviderAndUpdateCache(t, valInfo, big.NewInt(0))
	sdp.enableEpochsHandler = &testscommon.EnableEpochsHandlerStub{IsStakingV4Step2FlagEnabledField: true}

	sdp.cache[owner].blsKeys = append(sdp.cache[owner].blsKeys, []byte("newKey"))
	sdp.cache[owner].totalStaked = big.NewInt(2500)
	sdp.cache[owner].numStakedNodes++

	keysToUnStake, ownersWithNotEnoughFunds, err := sdp.ComputeUnQualifiedNodes(valInfo)
	require.Nil(t, err)

	expectedUnStakedKeys := [][]byte{[]byte("blsKey1"), []byte("newKey")}
	expectedOwnerWithNotEnoughFunds := map[string][][]byte{owner: expectedUnStakedKeys}
	require.Equal(t, expectedUnStakedKeys, keysToUnStake)
	require.Equal(t, expectedOwnerWithNotEnoughFunds, ownersWithNotEnoughFunds)
}

func TestStakingDataProvider_GetTotalStakeEligibleNodes(t *testing.T) {
	t.Parallel()

	owner := append([]byte("owner"), bytes.Repeat([]byte{1}, 27)...)
	topUpVal := big.NewInt(828743)
	blsKey := []byte("bls key")

	sdp := createStakingDataProviderWithRealArgs(t, owner, blsKey, topUpVal)

	expectedRes := big.NewInt(37)
	sdp.totalEligibleStake = expectedRes

	res := sdp.GetTotalStakeEligibleNodes()
	require.Equal(t, 0, expectedRes.Cmp(res))
	require.False(t, res == expectedRes, "GetTotalStakeEligibleNodes should have returned a new pointer")
}

func TestStakingDataProvider_GetTotalTopUpStakeEligibleNodes(t *testing.T) {
	t.Parallel()

	owner := append([]byte("owner"), bytes.Repeat([]byte{1}, 27)...)
	topUpVal := big.NewInt(828743)
	blsKey := []byte("bls key")

	sdp := createStakingDataProviderWithRealArgs(t, owner, blsKey, topUpVal)

	expectedRes := big.NewInt(37)
	sdp.totalEligibleTopUpStake = expectedRes

	res := sdp.GetTotalTopUpStakeEligibleNodes()
	require.Equal(t, 0, expectedRes.Cmp(res))
	require.False(t, res == expectedRes, "GetTotalTopUpStakeEligibleNodes should have returned a new pointer")
}

func TestStakingDataProvider_GetNodeStakedTopUpOwnerNotInCacheShouldErr(t *testing.T) {
	t.Parallel()

	owner := append([]byte("owner"), bytes.Repeat([]byte{1}, 27)...)
	topUpVal := big.NewInt(828743)
	blsKey := []byte("bls key")

	sdp := createStakingDataProviderWithRealArgs(t, owner, blsKey, topUpVal)

	expectedRes := big.NewInt(37)
	sdp.totalEligibleTopUpStake = expectedRes

	res, err := sdp.GetNodeStakedTopUp(blsKey)
	require.Equal(t, epochStart.ErrOwnerDoesntHaveEligibleNodesInEpoch, err)
	require.Nil(t, res)
}

func TestStakingDataProvider_GetNodeStakedTopUpScCallError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected")

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	basePrice := big.NewInt(100000)
	stakeVal := big.NewInt(0).Add(topUpVal, basePrice)
	numRunContractCalls := 0

	sdp := createStakingDataProviderWithMockArgs(t, owner, topUpVal, stakeVal, &numRunContractCalls)
	sdp.systemVM = &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(_ *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			return nil, expectedErr
		},
	}

	res, err := sdp.GetNodeStakedTopUp(owner)
	require.Equal(t, expectedErr, err)
	require.Nil(t, res)
}

func TestStakingDataProvider_GetNodeStakedTopUpShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	basePrice := big.NewInt(100000)
	stakeVal := big.NewInt(0).Add(topUpVal, basePrice)
	numRunContractCalls := 0

	sdp := createStakingDataProviderWithMockArgs(t, owner, topUpVal, stakeVal, &numRunContractCalls)

	expectedOwnerStats := &ownerStats{
		eligibleTopUpPerNode: big.NewInt(37),
	}
	sdp.SetInCache(owner, expectedOwnerStats)

	res, err := sdp.GetNodeStakedTopUp(owner)
	require.NoError(t, err)
	require.Equal(t, expectedOwnerStats.eligibleTopUpPerNode, res)
}

func TestStakingDataProvider_PrepareStakingDataForRewards(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	basePrice := big.NewInt(100000)
	stakeVal := big.NewInt(0).Add(topUpVal, basePrice)
	numRunContractCalls := 0

	sdp := createStakingDataProviderWithMockArgs(t, owner, topUpVal, stakeVal, &numRunContractCalls)

	validatorsMap := state.NewShardValidatorsInfoMap()
	_ = validatorsMap.Add(&state.ValidatorInfo{PublicKey: owner, ShardId: 0})
	err := sdp.PrepareStakingData(validatorsMap)
	require.NoError(t, err)
}

func TestStakingDataProvider_FillValidatorInfo(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	basePrice := big.NewInt(100000)
	stakeVal := big.NewInt(0).Add(topUpVal, basePrice)
	numRunContractCalls := 0

	sdp := createStakingDataProviderWithMockArgs(t, owner, topUpVal, stakeVal, &numRunContractCalls)

	err := sdp.FillValidatorInfo(&state.ValidatorInfo{PublicKey: []byte("bls key")})
	require.NoError(t, err)
}

func TestCheckAndFillOwnerValidatorAuctionData(t *testing.T) {
	t.Parallel()

	t.Run("validator not in auction, expect no error, no owner data update", func(t *testing.T) {
		t.Parallel()
		args := createStakingDataProviderArgs()
		sdp, _ := NewStakingDataProvider(args)

		ownerData := &ownerStats{}
		err := sdp.checkAndFillOwnerValidatorAuctionData([]byte("owner"), ownerData, &state.ValidatorInfo{List: string(common.NewList)})
		require.Nil(t, err)
		require.Equal(t, &ownerStats{}, ownerData)
	})

	t.Run("validator in auction, but no staked node, expect error", func(t *testing.T) {
		t.Parallel()
		args := createStakingDataProviderArgs()
		sdp, _ := NewStakingDataProvider(args)

		owner := []byte("owner")
		ownerData := &ownerStats{numStakedNodes: 0}
		validator := &state.ValidatorInfo{PublicKey: []byte("validatorPubKey"), List: string(common.AuctionList)}

		err := sdp.checkAndFillOwnerValidatorAuctionData(owner, ownerData, validator)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), epochStart.ErrOwnerHasNoStakedNode.Error()))
		require.True(t, strings.Contains(err.Error(), hex.EncodeToString(owner)))
		require.True(t, strings.Contains(err.Error(), hex.EncodeToString(validator.PublicKey)))
		require.Equal(t, &ownerStats{numStakedNodes: 0}, ownerData)
	})

	t.Run("validator in auction, staking v4 not enabled yet, expect error", func(t *testing.T) {
		t.Parallel()
		args := createStakingDataProviderArgs()
		sdp, _ := NewStakingDataProvider(args)

		owner := []byte("owner")
		ownerData := &ownerStats{numStakedNodes: 1}
		validator := &state.ValidatorInfo{PublicKey: []byte("validatorPubKey"), List: string(common.AuctionList)}

		err := sdp.checkAndFillOwnerValidatorAuctionData(owner, ownerData, validator)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), epochStart.ErrReceivedAuctionValidatorsBeforeStakingV4.Error()))
		require.True(t, strings.Contains(err.Error(), hex.EncodeToString(owner)))
		require.True(t, strings.Contains(err.Error(), hex.EncodeToString(validator.PublicKey)))
		require.Equal(t, &ownerStats{numStakedNodes: 1}, ownerData)
	})

	t.Run("should update owner's data", func(t *testing.T) {
		t.Parallel()
		args := createStakingDataProviderArgs()
		sdp, _ := NewStakingDataProvider(args)
		sdp.enableEpochsHandler = &testscommon.EnableEpochsHandlerStub{IsStakingV4StartedField: true}

		owner := []byte("owner")
		ownerData := &ownerStats{numStakedNodes: 3, numActiveNodes: 3}
		validator := &state.ValidatorInfo{PublicKey: []byte("validatorPubKey"), List: string(common.AuctionList)}

		err := sdp.checkAndFillOwnerValidatorAuctionData(owner, ownerData, validator)
		require.Nil(t, err)
		require.Equal(t, &ownerStats{
			numStakedNodes: 3,
			numActiveNodes: 2,
			auctionList:    []state.ValidatorInfoHandler{validator},
		}, ownerData)
	})
}

func TestSelectKeysToUnStake(t *testing.T) {
	t.Parallel()

	t.Run("no validator removed", func(t *testing.T) {
		t.Parallel()
		args := createStakingDataProviderArgs()
		sdp, _ := NewStakingDataProvider(args)
		sdp.enableEpochsHandler = &testscommon.EnableEpochsHandlerStub{IsStakingV4Step2FlagEnabledField: true}

		sortedKeys := map[string][][]byte{
			string(common.AuctionList): {[]byte("pk0")},
		}
		unStakedKeys, removedValidators := sdp.selectKeysToUnStake(sortedKeys, 2)
		require.Equal(t, [][]byte{[]byte("pk0")}, unStakedKeys)
		require.Equal(t, 0, removedValidators)
	})

	t.Run("overflow from waiting", func(t *testing.T) {
		t.Parallel()
		args := createStakingDataProviderArgs()
		sdp, _ := NewStakingDataProvider(args)
		sdp.enableEpochsHandler = &testscommon.EnableEpochsHandlerStub{IsStakingV4Step2FlagEnabledField: true}

		sortedKeys := map[string][][]byte{
			string(common.AuctionList):  {[]byte("pk0")},
			string(common.EligibleList): {[]byte("pk2")},
			string(common.WaitingList):  {[]byte("pk3"), []byte("pk4"), []byte("pk5")},
		}
		unStakedKeys, removedValidators := sdp.selectKeysToUnStake(sortedKeys, 2)
		require.Equal(t, [][]byte{[]byte("pk0"), []byte("pk3")}, unStakedKeys)
		require.Equal(t, 1, removedValidators)
	})

	t.Run("overflow from eligible", func(t *testing.T) {
		t.Parallel()
		args := createStakingDataProviderArgs()
		sdp, _ := NewStakingDataProvider(args)
		sdp.enableEpochsHandler = &testscommon.EnableEpochsHandlerStub{IsStakingV4Step2FlagEnabledField: true}

		sortedKeys := map[string][][]byte{
			string(common.AuctionList):  {[]byte("pk0")},
			string(common.EligibleList): {[]byte("pk1"), []byte("pk2")},
			string(common.WaitingList):  {[]byte("pk4"), []byte("pk5")},
		}
		unStakedKeys, removedValidators := sdp.selectKeysToUnStake(sortedKeys, 4)
		require.Equal(t, [][]byte{[]byte("pk0"), []byte("pk4"), []byte("pk5"), []byte("pk1")}, unStakedKeys)
		require.Equal(t, 3, removedValidators)
	})

	t.Run("no overflow", func(t *testing.T) {
		t.Parallel()
		args := createStakingDataProviderArgs()
		sdp, _ := NewStakingDataProvider(args)
		sdp.enableEpochsHandler = &testscommon.EnableEpochsHandlerStub{IsStakingV4Step2FlagEnabledField: true}

		sortedKeys := map[string][][]byte{
			string(common.AuctionList):  {[]byte("pk0")},
			string(common.EligibleList): {[]byte("pk1")},
			string(common.WaitingList):  {[]byte("pk2")},
		}
		unStakedKeys, removedValidators := sdp.selectKeysToUnStake(sortedKeys, 3)
		require.Equal(t, [][]byte{[]byte("pk0"), []byte("pk2"), []byte("pk1")}, unStakedKeys)
		require.Equal(t, 2, removedValidators)
	})
}

func createStakingDataProviderWithMockArgs(
	t *testing.T,
	owner []byte,
	topUpVal *big.Int,
	stakingVal *big.Int,
	numRunContractCalls *int,
) *stakingDataProvider {
	args := createStakingDataProviderArgs()
	args.SystemVM = &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			*numRunContractCalls++
			switch input.Function {
			case "getOwner":
				assert.Equal(t, vm.ValidatorSCAddress, input.VMInput.CallerAddr)
				assert.Equal(t, vm.StakingSCAddress, input.RecipientAddr)

				return &vmcommon.VMOutput{
					ReturnData: [][]byte{owner},
				}, nil
			case "getTotalStakedTopUpStakedBlsKeys":
				assert.Equal(t, 1, len(input.Arguments))
				assert.Equal(t, owner, input.Arguments[0])
				assert.Equal(t, vm.ValidatorSCAddress, input.RecipientAddr)

				return &vmcommon.VMOutput{
					ReturnData: [][]byte{topUpVal.Bytes(), stakingVal.Bytes(), big.NewInt(3).Bytes()},
				}, nil

			}

			return nil, errors.New("unexpected call")
		},
	}
	sdp, err := NewStakingDataProvider(args)
	require.Nil(t, err)

	return sdp
}

func createStakingDataProviderWithRealArgs(t *testing.T, owner []byte, blsKey []byte, topUpVal *big.Int) *stakingDataProvider {
	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch: 1000,
	}, testscommon.CreateMemUnit())
	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1000000,
	})
	s, _ := NewSystemSCProcessor(args)
	require.NotNil(t, s)

	doStake(t, s.systemVM, s.userAccountsDB, owner, big.NewInt(0).Add(big.NewInt(1000), topUpVal), blsKey)

	argsStakingDataProvider := createStakingDataProviderArgs()
	argsStakingDataProvider.SystemVM = s.systemVM
	sdp, _ := NewStakingDataProvider(argsStakingDataProvider)

	return sdp
}

func saveOutputAccounts(t *testing.T, accountsDB state.AccountsAdapter, vmOutput *vmcommon.VMOutput) {
	for _, outputAccount := range vmOutput.OutputAccounts {
		account, errLoad := accountsDB.LoadAccount(outputAccount.Address)
		if errLoad != nil {
			log.Error(errLoad.Error())
		}
		require.Nil(t, errLoad)

		userAccount, _ := account.(state.UserAccountHandler)
		for _, storeUpdate := range outputAccount.StorageUpdates {
			_ = userAccount.SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
		}

		err := accountsDB.SaveAccount(account)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		require.Nil(t, err)
	}

	_, err := accountsDB.Commit()
	require.Nil(t, err)
}

func createStakingDataProviderAndUpdateCache(t *testing.T, validatorsInfo state.ShardValidatorsInfoMapHandler, topUpValue *big.Int) *stakingDataProvider {
	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch: 1,
	}, testscommon.CreateMemUnit())
	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1,
	})

	argsStakingDataProvider := createStakingDataProviderArgs()
	argsStakingDataProvider.SystemVM = args.SystemVM
	sdp, _ := NewStakingDataProvider(argsStakingDataProvider)
	args.StakingDataProvider = sdp
	s, _ := NewSystemSCProcessor(args)
	require.NotNil(t, s)

	for _, valInfo := range validatorsInfo.GetAllValidatorsInfo() {
		stake := big.NewInt(0).Add(big.NewInt(2500), topUpValue)
		if valInfo.GetList() != string(common.LeavingList) && valInfo.GetList() != string(common.InactiveList) {
			doStake(t, s.systemVM, s.userAccountsDB, valInfo.GetRewardAddress(), stake, valInfo.GetPublicKey())
		}
		updateCache(sdp, valInfo.GetRewardAddress(), valInfo.GetPublicKey(), valInfo.GetList(), stake)

	}

	return sdp
}

func updateCache(sdp *stakingDataProvider, ownerAddress []byte, blsKey []byte, list string, stake *big.Int) {
	owner := sdp.cache[string(ownerAddress)]

	if owner == nil {
		owner = &ownerStats{
			numEligible:          0,
			numStakedNodes:       0,
			totalTopUp:           big.NewInt(0),
			totalStaked:          big.NewInt(0),
			eligibleBaseStake:    big.NewInt(0),
			eligibleTopUpStake:   big.NewInt(0),
			eligibleTopUpPerNode: big.NewInt(0),
			blsKeys:              nil,
		}
	}

	owner.blsKeys = append(owner.blsKeys, blsKey)
	if list != string(common.LeavingList) && list != string(common.InactiveList) {
		if list == string(common.EligibleList) {
			owner.numEligible++
		}
		owner.numStakedNodes++
		owner.totalStaked = owner.totalStaked.Add(owner.totalStaked, stake)
	}

	sdp.cache[string(ownerAddress)] = owner
}

func createValidatorsInfo(nbShards uint32, nbEligible, nbWaiting, nbLeaving, nbInactive map[uint32]uint32) state.ShardValidatorsInfoMapHandler {
	validatorsInfo := state.NewShardValidatorsInfoMap()
	shardMap := shardsMap(nbShards)

	for shardID := range shardMap {
		valInfoList := make([]state.ValidatorInfoHandler, 0)
		for eligible := uint32(0); eligible < nbEligible[shardID]; eligible++ {
			vInfo := &state.ValidatorInfo{
				PublicKey:     []byte(fmt.Sprintf("blsKey%s%d%d", common.EligibleList, shardID, eligible)),
				ShardId:       shardID,
				List:          string(common.EligibleList),
				RewardAddress: []byte(fmt.Sprintf("address%s%d%d", common.EligibleList, shardID, eligible)),
			}
			valInfoList = append(valInfoList, vInfo)
		}
		for waiting := uint32(0); waiting < nbWaiting[shardID]; waiting++ {
			vInfo := &state.ValidatorInfo{
				PublicKey:     []byte(fmt.Sprintf("blsKey%s%d%d", common.WaitingList, shardID, waiting)),
				ShardId:       shardID,
				List:          string(common.WaitingList),
				RewardAddress: []byte(fmt.Sprintf("address%s%d%d", common.WaitingList, shardID, waiting)),
			}
			valInfoList = append(valInfoList, vInfo)
		}
		for leaving := uint32(0); leaving < nbLeaving[shardID]; leaving++ {
			vInfo := &state.ValidatorInfo{
				PublicKey:     []byte(fmt.Sprintf("blsKey%s%d%d", common.LeavingList, shardID, leaving)),
				ShardId:       shardID,
				List:          string(common.LeavingList),
				RewardAddress: []byte(fmt.Sprintf("address%s%d%d", common.LeavingList, shardID, leaving)),
			}
			valInfoList = append(valInfoList, vInfo)
		}

		for inactive := uint32(0); inactive < nbInactive[shardID]; inactive++ {
			vInfo := &state.ValidatorInfo{
				PublicKey:     []byte(fmt.Sprintf("blsKey%s%d%d", common.InactiveList, shardID, inactive)),
				ShardId:       shardID,
				List:          string(common.InactiveList),
				RewardAddress: []byte(fmt.Sprintf("address%s%d%d", common.InactiveList, shardID, inactive)),
			}
			valInfoList = append(valInfoList, vInfo)
		}
		_ = validatorsInfo.SetValidatorsInShard(shardID, valInfoList)
	}
	return validatorsInfo
}

func shardsMap(nbShards uint32) map[uint32]struct{} {
	shardMap := make(map[uint32]struct{})
	shardMap[core.MetachainShardId] = struct{}{}
	for i := uint32(0); i < nbShards; i++ {
		shardMap[i] = struct{}{}
	}
	return shardMap
}
