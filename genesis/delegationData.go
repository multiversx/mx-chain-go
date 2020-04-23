package genesis

import (
	"encoding/json"
	"fmt"
	"math/big"
)

const decodeBase = 10

// DelegationData specify the delegation address and the balance provided
type DelegationData struct {
	Address      string   `json:"address"`
	Value        *big.Int `json:"value"`
	AddressBytes []byte   `json:"-"`
}

// MarshalJSON is the function called when trying to serialize the object using the JSON marshaler
func (dd *DelegationData) MarshalJSON() ([]byte, error) {
	value := dd.Value
	if value == nil {
		value = big.NewInt(0)
	}

	s := struct {
		Address string `json:"address"`
		Value   string `json:"value"`
	}{
		Address: dd.Address,
		Value:   value.String(),
	}

	return json.Marshal(&s)
}

// UnmarshalJSON is the function called when trying to de-serialize the object using the JSON marshaler
func (dd *DelegationData) UnmarshalJSON(data []byte) error {
	s := struct {
		Address string `json:"address"`
		Value   string `json:"value"`
	}{}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	dd.Value, ok = big.NewInt(0).SetString(s.Value, decodeBase)
	if !ok {
		return fmt.Errorf("%w for '%s', delegation address %s",
			ErrInvalidDelegationValueString,
			s.Value,
			s.Address,
		)
	}

	dd.Address = s.Address

	return nil
}
