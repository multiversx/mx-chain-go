package genesis_test

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInitialBalance() *genesis.InitialBalance {
	return &genesis.InitialBalance{
		Address:        "0001",
		Supply:         "5",
		Balance:        "1",
		StakingBalance: "2",
		Delegation: &genesis.DelegationData{
			Address: "0002",
			Balance: "2",
		},
	}
}

func createSimpleInitialBalance(address string, balance int64) *genesis.InitialBalance {
	return &genesis.InitialBalance{
		Address:        address,
		Supply:         big.NewInt(balance).String(),
		Balance:        big.NewInt(balance).String(),
		StakingBalance: "0",
		Delegation: &genesis.DelegationData{
			Address: "",
			Balance: "0",
		},
	}
}

func createDelegatedInitialBalance(address string, delegated string, delegatedBalance int64) *genesis.InitialBalance {
	return &genesis.InitialBalance{
		Address:        address,
		Supply:         big.NewInt(delegatedBalance).String(),
		Balance:        "0",
		StakingBalance: "0",
		Delegation: &genesis.DelegationData{
			Address: delegated,
			Balance: big.NewInt(delegatedBalance).String(),
		},
	}
}

func createStakedInitialBalance(address string, stakedBalance int64) *genesis.InitialBalance {
	return &genesis.InitialBalance{
		Address:        address,
		Supply:         big.NewInt(stakedBalance).String(),
		Balance:        "0",
		StakingBalance: big.NewInt(stakedBalance).String(),
		Delegation: &genesis.DelegationData{
			Address: "",
			Balance: "0",
		},
	}
}

func TestNewGenesis_NilEntireBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis("./testdata/genesis_ok.json", nil)

	assert.Nil(t, g)
	assert.True(t, errors.Is(err, genesis.ErrNilEntireSupply))
}

func TestNewGenesis_ZeroEntireBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis("./testdata/genesis_ok.json", big.NewInt(0))

	assert.Nil(t, g)
	assert.True(t, errors.Is(err, genesis.ErrInvalidEntireSupply))
}

func TestNewGenesis_BadFilenameShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis("inexistent file", big.NewInt(1))

	assert.Nil(t, g)
	assert.NotNil(t, err)
}

func TestNewGenesis_BadJsonShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis("testdata/genesis_bad.json", big.NewInt(1))

	assert.Nil(t, g)
	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestNewGenesis_ShouldWork(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis("testdata/genesis_ok.json", big.NewInt(30))

	assert.NotNil(t, g)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(g.InitialBalances()))
}

//------- process

func TestGenesis_ProcessEmptyAddressShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Address = ""
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyAddress))
}

func TestGenesis_ProcessInvalidAddressShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Address = "invalid address"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestGenesis_ProcessInvalidSupplyStringShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Supply = "not-a-number"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidSupplyString))
}

func TestGenesis_ProcessInvalidBalanceStringShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Balance = "not-a-number"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidBalanceString))
}

func TestGenesis_ProcessInvalidStakingBalanceStringShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.StakingBalance = "not-a-number"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidStakingBalanceString))
}

func TestGenesis_ProcessInvalidDelegationBalanceStringShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Delegation.Balance = "not-a-number"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationBalanceString))
}

func TestGenesis_ProcessEmptyDelegationAddressButWithBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Delegation.Address = ""
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyDelegationAddress))
}

func TestGenesis_ProcessInvalidDelegationAddressShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Delegation.Address = "invalid address"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationAddress))
}

func TestGenesis_ProcessInvalidSupplyShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Supply = "-1"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidSupply))

	ib.Supply = "0"

	err = g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidSupply))
}

func TestGenesis_ProcessInvalidBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Balance = "-1"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidBalance))
}

func TestGenesis_ProcessInvalidStakingBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.StakingBalance = "-1"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidStakingBalance))
}

func TestGenesis_ProcessInvalidDelegationBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Delegation.Balance = "-1"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationBalance))
}

func TestGenesis_ProcessSupplyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Supply = "4"
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrSupplyMismatch))
}

func TestGenesis_ProcessDuplicatesShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib1 := createMockInitialBalance()
	ib2 := createMockInitialBalance()
	g.SetInitialBalances([]*genesis.InitialBalance{ib1, ib2})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrDuplicateAddress))
}

func TestGenesis_ProcessEntireSupplyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	g.SetInitialBalances([]*genesis.InitialBalance{ib})
	g.SetEntireSupply(big.NewInt(4))

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrEntireSupplyMismatch))
}

func TestGenesis_AddressIsSmartContractShouldErr(t *testing.T) {
	t.Parallel()

	addr := strings.Repeat("0", (core.NumInitCharactersForScAddress+1)*2)
	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	ib.Address = addr
	g.SetInitialBalances([]*genesis.InitialBalance{ib})
	g.SetEntireSupply(big.NewInt(4))

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrAddressIsSmartContract))
}

func TestGenesis_ProcessShouldWork(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ib := createMockInitialBalance()
	g.SetInitialBalances([]*genesis.InitialBalance{ib})
	g.SetEntireSupply(big.NewInt(5))

	err := g.Process()
	assert.Nil(t, err)
}

//------- StakedUpon / DelegatedUpon

func TestGenesis_StakedUpon(t *testing.T) {
	t.Parallel()

	addr := "0001"
	stakedUpon := int64(78)

	g := &genesis.Genesis{}
	ib := createStakedInitialBalance(addr, stakedUpon)
	g.SetEntireSupply(big.NewInt(stakedUpon))
	g.SetInitialBalances([]*genesis.InitialBalance{ib})

	err := g.Process()
	require.Nil(t, err)

	computedStakedUpon := g.StakedUpon(addr)
	assert.Equal(t, big.NewInt(stakedUpon), computedStakedUpon)

	computedStakedUpon = g.StakedUpon("not found")
	assert.Equal(t, big.NewInt(0), computedStakedUpon)
}

func TestGenesis_DelegatedUpon(t *testing.T) {
	t.Parallel()

	addr1 := "1000"
	addr2 := "2000"
	delegatedUpon := int64(78)

	g := &genesis.Genesis{}
	ib1 := createDelegatedInitialBalance("0001", addr1, delegatedUpon)
	ib2 := createDelegatedInitialBalance("0002", addr1, delegatedUpon)
	ib3 := createDelegatedInitialBalance("0003", addr2, delegatedUpon)

	g.SetEntireSupply(big.NewInt(3 * delegatedUpon))
	g.SetInitialBalances([]*genesis.InitialBalance{ib1, ib2, ib3})

	err := g.Process()
	require.Nil(t, err)

	computedDelegatedUpon := g.DelegatedUpon(addr1)
	assert.Equal(t, big.NewInt(2*delegatedUpon), computedDelegatedUpon)

	computedDelegatedUpon = g.DelegatedUpon(addr2)
	assert.Equal(t, big.NewInt(delegatedUpon), computedDelegatedUpon)

	computedDelegatedUpon = g.DelegatedUpon("not found")
	assert.Equal(t, big.NewInt(0), computedDelegatedUpon)
}

//------- InitialBalancesSplitOnAddresses

func TestGenesis_InitialBalancesSplitOnAddressesNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ibs, err := g.InitialBalancesSplitOnAddresses(
		nil,
		&mock.AddressConverterMock{},
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestGenesis_InitialBalancesSplitOnAddressesNilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ibs, err := g.InitialBalancesSplitOnAddresses(
		&mock.ShardCoordinatorMock{},
		nil,
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilAddressConverter, err)
}

func TestGenesis_InitialBalancesSplitOnAddressesAddressConvertFailsShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	balance := int64(1)
	ibs := []*genesis.InitialBalance{
		createSimpleInitialBalance("0001", balance),
	}
	g.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	g.SetInitialBalances(ibs)
	err := g.Process()
	require.Nil(t, err)

	expectedErr := errors.New("expected error")
	ibsSplit, err := g.InitialBalancesSplitOnAddresses(
		&mock.ShardCoordinatorMock{},
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, err error) {
				return nil, expectedErr
			},
		},
	)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, ibsSplit)
}

func TestGenesis_InitialBalancesSplitOnAddresses(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	balance := int64(1)
	ibs := []*genesis.InitialBalance{
		createSimpleInitialBalance("0001", balance),
		createSimpleInitialBalance("0002", balance),
		createSimpleInitialBalance("0000", balance),
		createSimpleInitialBalance("0101", balance),
	}

	g.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	g.SetInitialBalances(ibs)
	err := g.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	ibsSplit, err := g.InitialBalancesSplitOnAddresses(
		threeSharder,
		&mock.AddressConverterMock{},
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(ibsSplit))
	assert.Equal(t, 2, len(ibsSplit[1]))
}

//------- InitialBalancesSplitOnDelegationAddresses

func TestGenesis_InitialBalancesSplitOnDelegationAddressesNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ibs, err := g.InitialBalancesSplitOnDelegationAddresses(
		nil,
		&mock.AddressConverterMock{},
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestGenesis_InitialBalancesSplitOnDelegationAddressesNilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	ibs, err := g.InitialBalancesSplitOnDelegationAddresses(
		&mock.ShardCoordinatorMock{},
		nil,
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilAddressConverter, err)
}

func TestGenesis_InitialBalancesSplitOnDelegationAddressesAddressConvertFailsShouldErr(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	balance := int64(1)
	ibs := []*genesis.InitialBalance{
		createDelegatedInitialBalance("0001", "0002", balance),
	}
	g.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	g.SetInitialBalances(ibs)
	err := g.Process()
	require.Nil(t, err)

	expectedErr := errors.New("expected error")
	ibsSplit, err := g.InitialBalancesSplitOnDelegationAddresses(
		&mock.ShardCoordinatorMock{},
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, err error) {
				return nil, expectedErr
			},
		},
	)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, ibsSplit)
}

func TestGenesis_InitialBalancesSplitOnDelegation(t *testing.T) {
	t.Parallel()

	g := &genesis.Genesis{}
	balance := int64(1)
	ibs := []*genesis.InitialBalance{
		createSimpleInitialBalance("0001", balance),
		createDelegatedInitialBalance("0101", "0001", balance),
		createDelegatedInitialBalance("0201", "0000", balance),
		createDelegatedInitialBalance("0301", "0002", balance),
		createDelegatedInitialBalance("0401", "0101", balance),
	}

	g.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	g.SetInitialBalances(ibs)
	err := g.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	ibsSplit, err := g.InitialBalancesSplitOnDelegationAddresses(
		threeSharder,
		&mock.AddressConverterMock{},
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(ibsSplit))
	assert.Equal(t, 2, len(ibsSplit[1]))
}
