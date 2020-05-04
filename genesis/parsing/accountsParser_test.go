package parsing_test

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInitialAccount() *data.InitialAccount {
	return &data.InitialAccount{
		Address:      "0001",
		Supply:       big.NewInt(5),
		Balance:      big.NewInt(1),
		StakingValue: big.NewInt(2),
		Delegation: &data.DelegationData{
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
	}
}

func createSimpleInitialAccount(address string, balance int64) *data.InitialAccount {
	return &data.InitialAccount{
		Address:      address,
		Supply:       big.NewInt(balance),
		Balance:      big.NewInt(balance),
		StakingValue: big.NewInt(0),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
}

func createDelegatedInitialAccount(address string, delegated string, delegatedBalance int64) *data.InitialAccount {
	return &data.InitialAccount{
		Address:      address,
		Supply:       big.NewInt(delegatedBalance),
		Balance:      big.NewInt(0),
		StakingValue: big.NewInt(0),
		Delegation: &data.DelegationData{
			Address: delegated,
			Value:   big.NewInt(delegatedBalance),
		},
	}
}

func createStakedInitialAccount(address string, stakedBalance int64) *data.InitialAccount {
	return &data.InitialAccount{
		Address:      address,
		Supply:       big.NewInt(stakedBalance),
		Balance:      big.NewInt(0),
		StakingValue: big.NewInt(stakedBalance),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
}

func TestNewAccountsParser_NilEntireBalanceShouldErr(t *testing.T) {
	t.Parallel()

	ap, err := parsing.NewAccountsParser(
		"./testdata/genesis_ok.json",
		nil,
		createMockHexPubkeyConverter(),
	)

	assert.True(t, check.IfNil(ap))
	assert.True(t, errors.Is(err, genesis.ErrNilEntireSupply))
}

func TestNewAccountsParser_ZeroEntireBalanceShouldErr(t *testing.T) {
	t.Parallel()

	ap, err := parsing.NewAccountsParser(
		"./testdata/genesis_ok.json",
		big.NewInt(0),
		createMockHexPubkeyConverter(),
	)

	assert.True(t, check.IfNil(ap))
	assert.True(t, errors.Is(err, genesis.ErrInvalidEntireSupply))
}

func TestNewAccountsParser_BadFilenameShouldErr(t *testing.T) {
	t.Parallel()

	ap, err := parsing.NewAccountsParser(
		"inexistent file",
		big.NewInt(1),
		createMockHexPubkeyConverter(),
	)

	assert.True(t, check.IfNil(ap))
	assert.NotNil(t, err)
}

func TestNewAccountsParser_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	ap, err := parsing.NewAccountsParser(
		"inexistent file",
		big.NewInt(1),
		nil,
	)

	assert.True(t, check.IfNil(ap))
	assert.Equal(t, genesis.ErrNilPubkeyConverter, err)
}

func TestNewAccountsParser_BadJsonShouldErr(t *testing.T) {
	t.Parallel()

	ap, err := parsing.NewAccountsParser(
		"testdata/genesis_bad.json",
		big.NewInt(1),
		createMockHexPubkeyConverter(),
	)

	assert.True(t, check.IfNil(ap))
	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestNewAccountsParser_ShouldWork(t *testing.T) {
	t.Parallel()

	ap, err := parsing.NewAccountsParser(
		"testdata/genesis_ok.json",
		big.NewInt(30),
		createMockHexPubkeyConverter(),
	)

	assert.False(t, check.IfNil(ap))
	assert.Nil(t, err)
	assert.Equal(t, 6, len(ap.InitialAccounts()))
}

//------- process

func TestAccountsParser_ProcessEmptyAddressShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = ""
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyAddress))
}

func TestAccountsParser_ProcessInvalidAddressShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = "invalid address"
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestAccountsParser_ProcessEmptyDelegationAddressButWithBalanceShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Address = ""
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyDelegationAddress))
}

func TestAccountsParser_ProcessInvalidDelegationAddressShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Address = "invalid address"
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationAddress))
}

func TestAccountsParser_ProcessInvalidSupplyShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Supply = big.NewInt(-1)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidSupply))

	ib.Supply = big.NewInt(0)

	err = ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidSupply))
}

func TestAccountsParser_ProcessInvalidBalanceShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Balance = big.NewInt(-1)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidBalance))
}

func TestAccountsParser_ProcessInvalidStakingBalanceShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.StakingValue = big.NewInt(-1)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidStakingBalance))
}

func TestAccountsParser_ProcessInvalidDelegationValueShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Value = big.NewInt(-1)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationValue))
}

func TestAccountsParser_ProcessSupplyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Supply = big.NewInt(4)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrSupplyMismatch))
}

func TestAccountsParser_ProcessDuplicatesShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib1 := createMockInitialAccount()
	ib2 := createMockInitialAccount()
	ap.SetInitialAccounts([]*data.InitialAccount{ib1, ib2})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrDuplicateAddress))
}

func TestAccountsParser_ProcessEntireSupplyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ap.SetInitialAccounts([]*data.InitialAccount{ib})
	ap.SetEntireSupply(big.NewInt(4))

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrEntireSupplyMismatch))
}

func TestAccountsParser_AddressIsSmartContractShouldErr(t *testing.T) {
	t.Parallel()

	addr := strings.Repeat("0", (core.NumInitCharactersForScAddress+1)*2)
	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = addr
	ap.SetInitialAccounts([]*data.InitialAccount{ib})
	ap.SetEntireSupply(big.NewInt(4))

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrAddressIsSmartContract))
}

func TestAccountsParser_ProcessShouldWork(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ap.SetInitialAccounts([]*data.InitialAccount{ib})
	ap.SetEntireSupply(big.NewInt(5))

	err := ap.Process()
	assert.Nil(t, err)
}

//------- StakedUpon / DelegatedUpon

func TestAccountsParser_StakedUpon(t *testing.T) {
	t.Parallel()

	addr := "0001"
	stakedUpon := int64(78)

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createStakedInitialAccount(addr, stakedUpon)
	ap.SetEntireSupply(big.NewInt(stakedUpon))
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	require.Nil(t, err)

	computedStakedUpon := ap.StakedUpon(addr)
	assert.Equal(t, big.NewInt(stakedUpon), computedStakedUpon)

	computedStakedUpon = ap.StakedUpon("not found")
	assert.Equal(t, big.NewInt(0), computedStakedUpon)
}

func TestAccountsParser_DelegatedUpon(t *testing.T) {
	t.Parallel()

	addr1 := "1000"
	addr2 := "2000"
	delegatedUpon := int64(78)

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib1 := createDelegatedInitialAccount("0001", addr1, delegatedUpon)
	ib2 := createDelegatedInitialAccount("0002", addr1, delegatedUpon)
	ib3 := createDelegatedInitialAccount("0003", addr2, delegatedUpon)

	ap.SetEntireSupply(big.NewInt(3 * delegatedUpon))
	ap.SetInitialAccounts([]*data.InitialAccount{ib1, ib2, ib3})

	err := ap.Process()
	require.Nil(t, err)

	computedDelegatedUpon := ap.DelegatedUpon(addr1)
	assert.Equal(t, big.NewInt(2*delegatedUpon), computedDelegatedUpon)

	computedDelegatedUpon = ap.DelegatedUpon(addr2)
	assert.Equal(t, big.NewInt(delegatedUpon), computedDelegatedUpon)

	computedDelegatedUpon = ap.DelegatedUpon("not found")
	assert.Equal(t, big.NewInt(0), computedDelegatedUpon)
}

//------- InitialAccountsSplitOnAddressesShards

func TestAccountsParser_InitialAccountsSplitOnAddressesShardsNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ibs, err := ap.InitialAccountsSplitOnAddressesShards(
		nil,
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestAccountsParser_InitialAccountsSplitOnAddressesShards(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*data.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createSimpleInitialAccount("0002", balance),
		createSimpleInitialAccount("0000", balance),
		createSimpleInitialAccount("0101", balance),
	}

	ap.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	ap.SetInitialAccounts(ibs)
	err := ap.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	ibsSplit, err := ap.InitialAccountsSplitOnAddressesShards(
		threeSharder,
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(ibsSplit))
	assert.Equal(t, 2, len(ibsSplit[1]))
}

//------- InitialAccountsSplitOnDelegationAddressesShards

func TestAccountsParser_InitialAccountsSplitOnDelegationAddressesShardsNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ibs, err := ap.InitialAccountsSplitOnDelegationAddressesShards(
		nil,
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestAccountsParser_InitialAccountsSplitOnDelegationAddressesShards(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*data.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createDelegatedInitialAccount("0101", "0001", balance),
		createDelegatedInitialAccount("0201", "0000", balance),
		createDelegatedInitialAccount("0301", "0002", balance),
		createDelegatedInitialAccount("0401", "0101", balance),
	}

	ap.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	ap.SetInitialAccounts(ibs)
	err := ap.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	ibsSplit, err := ap.InitialAccountsSplitOnDelegationAddressesShards(
		threeSharder,
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(ibsSplit))
	assert.Equal(t, 2, len(ibsSplit[1]))
}
