package genesis

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createMockInitialAccount() *InitialAccount {
	return &InitialAccount{
		Address:      "addres",
		Supply:       big.NewInt(1142),
		Balance:      big.NewInt(2242),
		StakingValue: big.NewInt(3342),
		Delegation: &DelegationData{
			Address: "delegation address",
			Value:   big.NewInt(4442),
		},
	}
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

	assert.True(t, errors.Is(err, ErrInvalidSupplyString))
}

func TestInitialAccount_UnmarshalNotAValidBalanceShouldErr(t *testing.T) {
	t.Parallel()

	input := createMockInitialAccount()

	buff, _ := json.Marshal(input)
	buff = bytes.Replace(buff, []byte(input.Balance.String()), []byte("not a number"), 1)

	recovered := &InitialAccount{}
	err := json.Unmarshal(buff, recovered)

	assert.True(t, errors.Is(err, ErrInvalidBalanceString))
}

func TestInitialAccount_UnmarshalNotAValidStakingValueShouldErr(t *testing.T) {
	t.Parallel()

	input := createMockInitialAccount()

	buff, _ := json.Marshal(input)
	buff = bytes.Replace(buff, []byte(input.StakingValue.String()), []byte("not a number"), 1)

	recovered := &InitialAccount{}
	err := json.Unmarshal(buff, recovered)

	assert.True(t, errors.Is(err, ErrInvalidStakingBalanceString))
}
