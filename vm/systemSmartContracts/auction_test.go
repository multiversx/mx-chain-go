package systemSmartContracts

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func createABid(totalStakeValue uint64, numBlsKeys uint32, maxStakePerNode uint64) AuctionData {
	data := AuctionData{
		RewardAddress:   []byte("addr"),
		StartNonce:      0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0).SetUint64(totalStakeValue),
		BlockedStake:    big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0).SetUint64(maxStakePerNode),
	}

	keys := make([][]byte, 0)
	for i := uint32(0); i < numBlsKeys; i++ {
		keys = append(keys, []byte(fmt.Sprintf("%d", i)))
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

	//expectedNodePrice := big.NewInt(12500000)

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(minStakeValue, minStep, totalSupply, unBoundPeriod, numNodes, eei, kg)

	bids := []AuctionData{
		createABid(25000000, 2, 12500000),
		createABid(30000000, 3, 10000000),
		createABid(40000000, 2, 20000000),
		createABid(50000000, 2, 25000000),
	}

	data := stakingAuctionSC.selection(bids)
	fmt.Println(data)
}
