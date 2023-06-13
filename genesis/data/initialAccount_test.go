package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInitialAccount() *InitialAccount {
	return &InitialAccount{
		Address:      "address",
		Supply:       big.NewInt(1142),
		Balance:      big.NewInt(2242),
		StakingValue: big.NewInt(3342),
		Delegation: &DelegationData{
			Address: "delegation address",
			Value:   big.NewInt(4442),
		},
	}
}

func TestInitialAccount_UnmarshalJSONInvalidValue(t *testing.T) {
	t.Parallel()

	acc := &InitialAccount{}
	err := acc.UnmarshalJSON([]byte("invalid data"))
	assert.Error(t, err)
}

func TestInitialAccount_UnmarshalJSON_MarshalUnmarshalEmptyStruct(t *testing.T) {
	t.Parallel()

	input := &InitialAccount{
		Address:      "",
		Supply:       nil,
		Balance:      nil,
		StakingValue: nil,
		Delegation:   nil,
	}
	expected := &InitialAccount{
		Address:      "",
		Supply:       big.NewInt(0),
		Balance:      big.NewInt(0),
		StakingValue: big.NewInt(0),
		Delegation: &DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}

	buff, err := json.Marshal(input)
	assert.Nil(t, err)

	recovered := &InitialAccount{}
	err = json.Unmarshal(buff, recovered)

	assert.Nil(t, err)
	assert.Equal(t, expected, recovered)
}

func TestInitialAccount_MarshalUnmarshalValidStruct(t *testing.T) {
	t.Parallel()

	address := "address"
	supply := int64(1142)
	balance := int64(2242)
	stakingValue := int64(3342)
	delegationAddress := "delegation address"
	delegationValue := int64(4442)
	input := &InitialAccount{
		Address:      address,
		Supply:       big.NewInt(supply),
		Balance:      big.NewInt(balance),
		StakingValue: big.NewInt(stakingValue),
		Delegation: &DelegationData{
			Address: delegationAddress,
			Value:   big.NewInt(delegationValue),
		},
	}

	buff, err := json.Marshal(input)
	assert.Nil(t, err)

	recovered := &InitialAccount{}
	err = json.Unmarshal(buff, recovered)

	assert.Nil(t, err)
	assert.Equal(t, input, recovered)
}

func TestInitialAccount_UnmarshalNotAValidSupplyShouldErr(t *testing.T) {
	t.Parallel()

	input := createMockInitialAccount()

	buff, _ := json.Marshal(input)
	buff = bytes.Replace(buff, []byte(input.Supply.String()), []byte("not a number"), 1)

	recovered := &InitialAccount{}
	err := json.Unmarshal(buff, recovered)

	assert.True(t, errors.Is(err, genesis.ErrInvalidSupplyString))
}

func TestInitialAccount_UnmarshalNotAValidBalanceShouldErr(t *testing.T) {
	t.Parallel()

	input := createMockInitialAccount()

	buff, _ := json.Marshal(input)
	buff = bytes.Replace(buff, []byte(input.Balance.String()), []byte("not a number"), 1)

	recovered := &InitialAccount{}
	err := json.Unmarshal(buff, recovered)

	assert.True(t, errors.Is(err, genesis.ErrInvalidBalanceString))
}

func TestInitialAccount_UnmarshalNotAValidStakingValueShouldErr(t *testing.T) {
	t.Parallel()

	input := createMockInitialAccount()

	buff, _ := json.Marshal(input)
	buff = bytes.Replace(buff, []byte(input.StakingValue.String()), []byte("not a number"), 1)

	recovered := &InitialAccount{}
	err := json.Unmarshal(buff, recovered)

	assert.True(t, errors.Is(err, genesis.ErrInvalidStakingBalanceString))
}

func TestInitialAccount_AddressBytes(t *testing.T) {
	t.Parallel()

	ia := &InitialAccount{}
	addrBytes := []byte("address bytes")
	ia.SetAddressBytes(addrBytes)
	recoverdAddrBytes := ia.AddressBytes()

	assert.Equal(t, addrBytes, recoverdAddrBytes)
}

func TestInitialAccount_Clone(t *testing.T) {
	t.Parallel()

	ia := &InitialAccount{
		Address:      "address",
		Supply:       big.NewInt(45),
		Balance:      big.NewInt(56),
		StakingValue: big.NewInt(78),
		addressBytes: []byte("address bytes"),
		Delegation: &DelegationData{
			Address:      "delegation address",
			Value:        big.NewInt(910),
			addressBytes: []byte("delegation address bytes"),
		},
	}

	iaCloned := ia.Clone()

	assert.Equal(t, ia, iaCloned)
	assert.False(t, ia == iaCloned) //pointer testing
	assert.False(t, ia.Supply == iaCloned.GetSupply())
	assert.False(t, ia.Balance == iaCloned.GetBalanceValue())
	assert.False(t, ia.StakingValue == iaCloned.GetStakingValue())
	assert.False(t, ia.Delegation == iaCloned.GetDelegationHandler())
}

func TestInitialAccount_Getters(t *testing.T) {
	t.Parallel()

	accountAddr := "account address"
	supply := big.NewInt(56)
	balance := big.NewInt(67)
	staking := big.NewInt(78)
	dd := &DelegationData{}
	ia := &InitialAccount{
		Address:      accountAddr,
		Supply:       supply,
		Balance:      balance,
		StakingValue: staking,
		Delegation:   dd,
	}

	require.False(t, check.IfNil(ia))
	require.False(t, check.IfNil(ia.GetDelegationHandler()))
	assert.Equal(t, accountAddr, ia.GetAddress())
	assert.Equal(t, supply, ia.GetSupply())
	assert.Equal(t, balance, ia.GetBalanceValue())
	assert.Equal(t, staking, ia.GetStakingValue())
	assert.Equal(t, dd, ia.GetDelegationHandler())
}
