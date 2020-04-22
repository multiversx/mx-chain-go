package genesis_test

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInitialAccount() *genesis.InitialAccount {
	return &genesis.InitialAccount{
		Address:      "0001",
		Supply:       big.NewInt(5),
		Balance:      big.NewInt(1),
		StakingValue: big.NewInt(2),
		Delegation: &genesis.DelegationData{
			Address: "0002",
			Value:   big.NewInt(2),
		},
	}
}

func createMockHexPubkeyConverter() *mock.PubkeyConverterStub {
	return &mock.PubkeyConverterStub{
		DecodeCalled: func(humanReadable string) ([]byte, error) {
			return hex.DecodeString(humanReadable)
		},
		CreateAddressFromBytesCalled: func(pkBytes []byte) (state.AddressContainer, error) {
			return mock.NewAddressMock(pkBytes), nil
		},
	}
}

func createSimpleInitialAccount(address string, balance int64) *genesis.InitialAccount {
	return &genesis.InitialAccount{
		Address:      address,
		Supply:       big.NewInt(balance),
		Balance:      big.NewInt(balance),
		StakingValue: big.NewInt(0),
		Delegation: &genesis.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
}

func createDelegatedInitialAccount(address string, delegated string, delegatedBalance int64) *genesis.InitialAccount {
	return &genesis.InitialAccount{
		Address:      address,
		Supply:       big.NewInt(delegatedBalance),
		Balance:      big.NewInt(0),
		StakingValue: big.NewInt(0),
		Delegation: &genesis.DelegationData{
			Address: delegated,
			Value:   big.NewInt(delegatedBalance),
		},
	}
}

func createStakedInitialAccount(address string, stakedBalance int64) *genesis.InitialAccount {
	return &genesis.InitialAccount{
		Address:      address,
		Supply:       big.NewInt(stakedBalance),
		Balance:      big.NewInt(0),
		StakingValue: big.NewInt(stakedBalance),
		Delegation: &genesis.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
}

func TestNewGenesis_NilEntireBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis(
		"./testdata/genesis_ok.json",
		nil,
		createMockHexPubkeyConverter(),
	)

	assert.True(t, check.IfNil(g))
	assert.True(t, errors.Is(err, genesis.ErrNilEntireSupply))
}

func TestNewGenesis_ZeroEntireBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis(
		"./testdata/genesis_ok.json",
		big.NewInt(0),
		createMockHexPubkeyConverter(),
	)

	assert.True(t, check.IfNil(g))
	assert.True(t, errors.Is(err, genesis.ErrInvalidEntireSupply))
}

func TestNewGenesis_BadFilenameShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis(
		"inexistent file",
		big.NewInt(1),
		createMockHexPubkeyConverter(),
	)

	assert.True(t, check.IfNil(g))
	assert.NotNil(t, err)
}

func TestNewGenesis_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis(
		"inexistent file",
		big.NewInt(1),
		nil,
	)

	assert.True(t, check.IfNil(g))
	assert.Equal(t, genesis.ErrNilPubkeyConverter, err)
}

func TestNewGenesis_BadJsonShouldErr(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis(
		"testdata/genesis_bad.json",
		big.NewInt(1),
		createMockHexPubkeyConverter(),
	)

	assert.True(t, check.IfNil(g))
	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestNewGenesis_ShouldWork(t *testing.T) {
	t.Parallel()

	g, err := genesis.NewGenesis(
		"testdata/genesis_ok.json",
		big.NewInt(30),
		createMockHexPubkeyConverter(),
	)

	assert.False(t, check.IfNil(g))
	assert.Nil(t, err)
	assert.Equal(t, 6, len(g.InitialAccounts()))
}

//------- process

func TestGenesis_ProcessEmptyAddressShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = ""
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyAddress))
}

func TestGenesis_ProcessInvalidAddressShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = "invalid address"
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestGenesis_ProcessEmptyDelegationAddressButWithBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Address = ""
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyDelegationAddress))
}

func TestGenesis_ProcessInvalidDelegationAddressShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Address = "invalid address"
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationAddress))
}

func TestGenesis_ProcessInvalidSupplyShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Supply = big.NewInt(-1)
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidSupply))

	ib.Supply = big.NewInt(0)

	err = g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidSupply))
}

func TestGenesis_ProcessInvalidBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Balance = big.NewInt(-1)
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidBalance))
}

func TestGenesis_ProcessInvalidStakingBalanceShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.StakingValue = big.NewInt(-1)
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidStakingBalance))
}

func TestGenesis_ProcessInvalidDelegationValueShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Value = big.NewInt(-1)
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationValue))
}

func TestGenesis_ProcessSupplyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Supply = big.NewInt(4)
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrSupplyMismatch))
}

func TestGenesis_ProcessDuplicatesShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib1 := createMockInitialAccount()
	ib2 := createMockInitialAccount()
	g.SetInitialAccounts([]*genesis.InitialAccount{ib1, ib2})

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrDuplicateAddress))
}

func TestGenesis_ProcessEntireSupplyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})
	g.SetEntireSupply(big.NewInt(4))

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrEntireSupplyMismatch))
}

func TestGenesis_AddressIsSmartContractShouldErr(t *testing.T) {
	t.Parallel()

	addr := strings.Repeat("0", (core.NumInitCharactersForScAddress+1)*2)
	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = addr
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})
	g.SetEntireSupply(big.NewInt(4))

	err := g.Process()
	assert.True(t, errors.Is(err, genesis.ErrAddressIsSmartContract))
}

func TestGenesis_ProcessShouldWork(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})
	g.SetEntireSupply(big.NewInt(5))

	err := g.Process()
	assert.Nil(t, err)
}

//------- StakedUpon / DelegatedUpon

func TestGenesis_StakedUpon(t *testing.T) {
	t.Parallel()

	addr := "0001"
	stakedUpon := int64(78)

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib := createStakedInitialAccount(addr, stakedUpon)
	g.SetEntireSupply(big.NewInt(stakedUpon))
	g.SetInitialAccounts([]*genesis.InitialAccount{ib})

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

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ib1 := createDelegatedInitialAccount("0001", addr1, delegatedUpon)
	ib2 := createDelegatedInitialAccount("0002", addr1, delegatedUpon)
	ib3 := createDelegatedInitialAccount("0003", addr2, delegatedUpon)

	g.SetEntireSupply(big.NewInt(3 * delegatedUpon))
	g.SetInitialAccounts([]*genesis.InitialAccount{ib1, ib2, ib3})

	err := g.Process()
	require.Nil(t, err)

	computedDelegatedUpon := g.DelegatedUpon(addr1)
	assert.Equal(t, big.NewInt(2*delegatedUpon), computedDelegatedUpon)

	computedDelegatedUpon = g.DelegatedUpon(addr2)
	assert.Equal(t, big.NewInt(delegatedUpon), computedDelegatedUpon)

	computedDelegatedUpon = g.DelegatedUpon("not found")
	assert.Equal(t, big.NewInt(0), computedDelegatedUpon)
}

//------- InitialAccountsSplitOnAddressesShards

func TestGenesis_InitialAccountsSplitOnAddressesShardsNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ibs, err := g.InitialAccountsSplitOnAddressesShards(
		nil,
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestGenesis_InitialAccountsSplitOnAddressesShardsShardsAddressConvertFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	g := genesis.NewTestGenesis(
		&mock.PubkeyConverterStub{
			CreateAddressFromBytesCalled: func(pubKey []byte) (container state.AddressContainer, err error) {
				return nil, expectedErr
			},
		},
	)
	balance := int64(1)
	ibs := []*genesis.InitialAccount{
		createSimpleInitialAccount("0001", balance),
	}
	g.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	g.SetInitialAccounts(ibs)
	err := g.Process()
	require.Nil(t, err)

	ibsSplit, err := g.InitialAccountsSplitOnAddressesShards(
		&mock.ShardCoordinatorMock{},
	)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, ibsSplit)
}

func TestGenesis_InitialAccountsSplitOnAddressesShards(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*genesis.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createSimpleInitialAccount("0002", balance),
		createSimpleInitialAccount("0000", balance),
		createSimpleInitialAccount("0101", balance),
	}

	g.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	g.SetInitialAccounts(ibs)
	err := g.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	ibsSplit, err := g.InitialAccountsSplitOnAddressesShards(
		threeSharder,
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(ibsSplit))
	assert.Equal(t, 2, len(ibsSplit[1]))
}

//------- InitialAccountsSplitOnDelegationAddressesShards

func TestGenesis_InitialAccountsSplitOnDelegationAddressesShardsNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	ibs, err := g.InitialAccountsSplitOnDelegationAddressesShards(
		nil,
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestGenesis_InitialAccountsSplitOnDelegationAddressesShardsPubkeyConverterFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	g := genesis.NewTestGenesis(
		&mock.PubkeyConverterStub{
			CreateAddressFromBytesCalled: func(pubKey []byte) (container state.AddressContainer, err error) {
				return nil, expectedErr
			},
		},
	)
	balance := int64(1)
	ibs := []*genesis.InitialAccount{
		createDelegatedInitialAccount("0001", "0002", balance),
	}
	g.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	g.SetInitialAccounts(ibs)
	err := g.Process()
	require.Nil(t, err)

	ibsSplit, err := g.InitialAccountsSplitOnDelegationAddressesShards(
		&mock.ShardCoordinatorMock{},
	)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, ibsSplit)
}

func TestGenesis_InitialAccountsSplitOnDelegationAddressesShards(t *testing.T) {
	t.Parallel()

	g := genesis.NewTestGenesis(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*genesis.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createDelegatedInitialAccount("0101", "0001", balance),
		createDelegatedInitialAccount("0201", "0000", balance),
		createDelegatedInitialAccount("0301", "0002", balance),
		createDelegatedInitialAccount("0401", "0101", balance),
	}

	g.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	g.SetInitialAccounts(ibs)
	err := g.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	ibsSplit, err := g.InitialAccountsSplitOnDelegationAddressesShards(
		threeSharder,
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(ibsSplit))
	assert.Equal(t, 2, len(ibsSplit[1]))
}
