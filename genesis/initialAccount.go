package genesis

import (
	"encoding/json"
	"fmt"
	"math/big"
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
			ErrInvalidSupplyString,
			s.Supply,
			s.Address,
		)
	}

	ia.Balance, ok = big.NewInt(0).SetString(s.Balance, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidBalanceString,
			s.Balance,
			s.Address,
		)
	}

	ia.StakingValue, ok = big.NewInt(0).SetString(s.StakingValue, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', address %s",
			ErrInvalidStakingBalanceString,
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
func (ia *InitialAccount) Clone() *InitialAccount {
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
