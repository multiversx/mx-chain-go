package data

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-go/genesis"
)

// InitialAccount provides information about one entry in the genesis file
type InitialAccount struct {
	Address      string          `json:"address"`
	Supply       *big.Int        `json:"supply"`
	Balance      *big.Int        `json:"balance"`
	StakingValue *big.Int        `json:"stakingvalue"`
	Delegation   *DelegationData `json:"delegation"`
	addressBytes []byte
}

// MarshalJSON is the function called when trying to serialize the object using the JSON marshaler
func (ia *InitialAccount) MarshalJSON() ([]byte, error) {
	supply := ia.Supply
	if supply == nil {
		supply = big.NewInt(0)
	}

	balance := ia.Balance
	if balance == nil {
		balance = big.NewInt(0)
	}

	stakingValue := ia.StakingValue
	if stakingValue == nil {
		stakingValue = big.NewInt(0)
	}

	delegation := ia.Delegation
	if delegation == nil {
		delegation = &DelegationData{}
	}

	s := struct {
		Address      string          `json:"address"`
		Supply       string          `json:"supply"`
		Balance      string          `json:"balance"`
		StakingValue string          `json:"stakingvalue"`
		Delegation   *DelegationData `json:"delegation"`
	}{
		Address:      ia.Address,
		Supply:       supply.String(),
		Balance:      balance.String(),
		StakingValue: stakingValue.String(),
		Delegation:   delegation,
	}

	return json.Marshal(&s)
}

// UnmarshalJSON is the function called when trying to de-serialize the object using the JSON marshaler
func (ia *InitialAccount) UnmarshalJSON(data []byte) error {
	s := struct {
		Address      string          `json:"address"`
		Supply       string          `json:"supply"`
		Balance      string          `json:"balance"`
		StakingValue string          `json:"stakingvalue"`
		Delegation   *DelegationData `json:"delegation"`
	}{}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	ia.Supply, ok = big.NewInt(0).SetString(s.Supply, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidSupplyString,
			s.Supply,
			s.Address,
		)
	}

	ia.Balance, ok = big.NewInt(0).SetString(s.Balance, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidBalanceString,
			s.Balance,
			s.Address,
		)
	}

	ia.StakingValue, ok = big.NewInt(0).SetString(s.StakingValue, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			genesis.ErrInvalidStakingBalanceString,
			s.StakingValue,
			s.Address,
		)
	}

	ia.Address = s.Address
	ia.Delegation = s.Delegation

	return nil
}

// AddressBytes will return the address as raw bytes
func (ia *InitialAccount) AddressBytes() []byte {
	return ia.addressBytes
}

// SetAddressBytes will set the address as raw bytes
func (ia *InitialAccount) SetAddressBytes(address []byte) {
	ia.addressBytes = address
}

// Clone will return a new instance of the initial account holding the same information
func (ia *InitialAccount) Clone() genesis.InitialAccountHandler {
	newInitialAccount := &InitialAccount{
		Address:      ia.Address,
		Supply:       big.NewInt(0).Set(ia.Supply),
		Balance:      big.NewInt(0).Set(ia.Balance),
		StakingValue: big.NewInt(0).Set(ia.StakingValue),
		Delegation:   ia.Delegation.Clone(),
		addressBytes: make([]byte, len(ia.addressBytes)),
	}

	copy(newInitialAccount.addressBytes, ia.addressBytes)

	return newInitialAccount
}

// GetAddress returns the address of the initial account
func (ia *InitialAccount) GetAddress() string {
	return ia.Address
}

// GetStakingValue returns the staking value
func (ia *InitialAccount) GetStakingValue() *big.Int {
	return ia.StakingValue
}

// GetBalanceValue returns the initial balance value
func (ia *InitialAccount) GetBalanceValue() *big.Int {
	return ia.Balance
}

// GetSupply returns the account's supply value
func (ia *InitialAccount) GetSupply() *big.Int {
	return ia.Supply
}

// GetDelegationHandler returns the delegation handler
func (ia *InitialAccount) GetDelegationHandler() genesis.DelegationDataHandler {
	return ia.Delegation
}

// IsInterfaceNil returns if underlying object is true
func (ia *InitialAccount) IsInterfaceNil() bool {
	return ia == nil
}
