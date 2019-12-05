package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func createABid(totalStakeValue uint64, numBlsKeys uint32, maxStakePerNode uint64) AuctionData {
	data := AuctionData{
		RewardAddress:   []byte("addr"),
		StartNonce:      0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0).SetUint64(totalStakeValue),
		BlockedStake:    big.NewInt(100000000),
		MaxStakePerNode: big.NewInt(0).SetUint64(maxStakePerNode),
	}

	keys := make([][]byte, 0)
	for i := uint32(0); i < numBlsKeys; i++ {
		keys = append(keys, []byte(fmt.Sprintf("%d", rand.Uint32())))
	}
	data.BlsPubKeys = keys

	return data
}

func TestAuctionSC_calculateNodePrice_Case1(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(10)

	eei := &mock.SystemEIStub{}

	kg := &mock.KeyGenMock{}

	expectedNodePrice := big.NewInt(20000000)

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	bids := []AuctionData{
		createABid(100000000, 100, 30000000),
		createABid(20000000, 100, 20000000),
		createABid(20000000, 100, 20000000),
		createABid(20000000, 100, 20000000),
		createABid(20000000, 100, 20000000),
		createABid(20000000, 100, 20000000),
	}

	nodePrice, err := stakingAuctionSC.calculateNodePrice(bids)
	assert.Equal(t, expectedNodePrice, nodePrice)
	assert.Nil(t, err)
}

func TestAuctionSC_calculateNodePrice_Case2(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(5)
	eei := &mock.SystemEIStub{}
	kg := &mock.KeyGenMock{}

	expectedNodePrice := big.NewInt(20000000)

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	bids := []AuctionData{
		createABid(100000000, 1, 30000000),
		createABid(50000000, 100, 25000000),
		createABid(30000000, 100, 15000000),
		createABid(40000000, 100, 20000000),
	}

	nodePrice, err := stakingAuctionSC.calculateNodePrice(bids)
	assert.Equal(t, expectedNodePrice, nodePrice)
	assert.Nil(t, err)
}

func TestAuctionSC_calculateNodePrice_Case3(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(5)
	eei := &mock.SystemEIStub{}
	kg := &mock.KeyGenMock{}

	expectedNodePrice := big.NewInt(12500000)

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	bids := []AuctionData{
		createABid(25000000, 2, 12500000),
		createABid(30000000, 3, 10000000),
		createABid(40000000, 2, 20000000),
		createABid(50000000, 2, 25000000),
	}

	nodePrice, err := stakingAuctionSC.calculateNodePrice(bids)
	assert.Equal(t, expectedNodePrice, nodePrice)
	assert.Nil(t, err)
}

func TestAuctionSC_calculateNodePrice_Case4ShouldErr(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(10)
	eei := &mock.SystemEIStub{}
	kg := &mock.KeyGenMock{}

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	bids := []AuctionData{
		createABid(25000000, 2, 12500000),
		createABid(30000000, 3, 10000000),
		createABid(40000000, 2, 20000000),
		createABid(50000000, 2, 25000000),
	}

	nodePrice, err := stakingAuctionSC.calculateNodePrice(bids)
	assert.Nil(t, nodePrice)
	assert.Equal(t, vm.ErrNotEnoughQualifiedNodes, err)
}

func TestAuctionSC_selection_Case1(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(5)
	eei := &mock.SystemEIStub{}
	kg := &mock.KeyGenMock{}

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	bids := []AuctionData{
		createABid(25000000, 2, 12500000),
		createABid(30000000, 3, 10000000),
		createABid(40000000, 2, 20000000),
		createABid(50000000, 2, 25000000),
	}

	data := stakingAuctionSC.selection(bids)
	assert.Equal(t, int(numNodes), len(data))
}

func TestStakingAuctionSC_ExecuteStakeWithoutArgumentsShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(5)

	arguments := CreateVmContractCallInput()
	auctionData := createABid(25000000, 2, 12500000)
	auctionDataBytes, _ := json.Marshal(&auctionData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Compare(key, arguments.CallerAddr) == 0 {
			return auctionDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		if bytes.Compare(key, arguments.CallerAddr) == 0 {
			var auctionData AuctionData
			_ = json.Unmarshal(value, &auctionData)
			assert.Equal(t, big.NewInt(26000000), auctionData.TotalStakeValue)
		}
	}

	kg := &mock.KeyGenMock{}

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	arguments.Function = "stake"
	arguments.CallValue = big.NewInt(1000000)

	errCode := stakingAuctionSC.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingAuctionSC_ExecuteStakeAddedNewPubKeysShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(5)

	arguments := CreateVmContractCallInput()
	auctionData := createABid(25000000, 2, 12500000)
	auctionDataBytes, _ := json.Marshal(&auctionData)

	key1 := []byte("Key1")
	key2 := []byte("Key2")
	rewardAddr := []byte("tralala2")
	maxStakePerNoce := big.NewInt(500)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Compare(key, arguments.CallerAddr) == 0 {
			return auctionDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		if bytes.Compare(key, arguments.CallerAddr) == 0 {
			var auctionData AuctionData
			_ = json.Unmarshal(value, &auctionData)
			assert.Equal(t, big.NewInt(26000000), auctionData.TotalStakeValue)
			assert.True(t, bytes.Compare(auctionData.BlsPubKeys[2], key1) == 0)
			assert.True(t, bytes.Compare(auctionData.BlsPubKeys[3], key2) == 0)
			assert.True(t, bytes.Compare(auctionData.RewardAddress, rewardAddr) == 0)
			assert.Equal(t, maxStakePerNoce, auctionData.MaxStakePerNode)
		}
	}

	kg := &mock.KeyGenMock{}

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	arguments.Function = "stake"
	arguments.CallValue = big.NewInt(1000000)
	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), key1, key2, maxStakePerNoce.Bytes(), rewardAddr}

	errCode := stakingAuctionSC.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingAuctionSC_ExecuteStakeUnStakeOneBlsPubKey(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(5)

	arguments := CreateVmContractCallInput()
	auctionData := createABid(25000000, 2, 12500000)
	auctionDataBytes, _ := json.Marshal(&auctionData)

	stakedData := StakedData{
		StartNonce:    0,
		Staked:        true,
		UnStakedNonce: 1,
		UnStakedEpoch: 0,
		RewardAddress: []byte("tralala1"),
	}
	stakedDataBytes, _ := json.Marshal(&stakedData)

	eei := &mock.SystemEIStub{}
	kg := &mock.KeyGenMock{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Compare(key, arguments.CallerAddr) == 0 {
			return auctionDataBytes
		}
		if bytes.Compare(key, auctionData.BlsPubKeys[0]) == 0 {
			return stakedDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		var stakedData StakedData
		_ = json.Unmarshal(value, &stakedData)

		assert.Equal(t, false, stakedData.Staked)
	}

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{auctionData.BlsPubKeys[0]}
	errCode := stakingAuctionSC.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingAuctionSC_ExecuteUnBound(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(10000000)
	totalSupply := big.NewInt(100000000000)
	minStep := big.NewInt(100000)
	unBoundPeriod := uint64(100000)
	numNodes := uint32(5)
	arguments := CreateVmContractCallInput()

	auctionData := createABid(25000000, 2, 12500000)
	auctionDataBytes, _ := json.Marshal(&auctionData)

	stakedData := StakedData{
		StartNonce:    0,
		Staked:        false,
		UnStakedNonce: 1,
		UnStakedEpoch: 0,
		RewardAddress: []byte("tralala1"),
	}
	stakedDataBytes, _ := json.Marshal(&stakedData)

	eei := &mock.SystemEIStub{}
	kg := &mock.KeyGenMock{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Compare(arguments.CallerAddr, key) == 0 {
			return auctionDataBytes
		}
		if bytes.Compare(key, auctionData.BlsPubKeys[0]) == 0 {
			return stakedDataBytes
		}

		return nil
	}

	eei.SetStorageCalled = func(key []byte, value []byte) {
		if bytes.Contains(key, arguments.CallerAddr) {
			var regData AuctionData
			_ = json.Unmarshal(value, &regData)

			assert.Equal(t, big.NewInt(90000000), regData.BlockedStake)
			assert.Equal(t, big.NewInt(15000000), regData.TotalStakeValue)
		}
	}

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	arguments.Function = "unBound"
	arguments.Arguments = [][]byte{auctionData.BlsPubKeys[0]}
	errCode := stakingAuctionSC.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}
