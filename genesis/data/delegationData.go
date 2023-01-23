package data

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-go/genesis"
)

const decodeBase = 10

// DelegationData specify the delegation address and the balance provided
type DelegationData struct {
	Address      string   `json:"address"`
	Value        *big.Int `json:"value"`
	addressBytes []byte
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
			genesis.ErrInvalidDelegationValueString,
			s.Value,
			s.Address,
		)
	}

	dd.Address = s.Address

	return nil
}

// AddressBytes will return the delegation address as raw bytes
func (dd *DelegationData) AddressBytes() []byte {
	return dd.addressBytes
}

// SetAddressBytes will set the delegation address as raw bytes
func (dd *DelegationData) SetAddressBytes(address []byte) {
	dd.addressBytes = address
}

// Clone will return a new instance of the delegation data holding the same information
func (dd *DelegationData) Clone() *DelegationData {
	newDelegationData := &DelegationData{
		Address:      dd.Address,
		Value:        big.NewInt(0).Set(dd.Value),
		addressBytes: make([]byte, len(dd.addressBytes)),
	}
	copy(newDelegationData.addressBytes, dd.addressBytes)

	return newDelegationData
}

// GetAddress returns the address as string
func (dd *DelegationData) GetAddress() string {
	return dd.Address
}

// GetValue returns the delegated value
func (dd *DelegationData) GetValue() *big.Int {
	return dd.Value
}

// IsInterfaceNil returns if underlying object is true
func (dd *DelegationData) IsInterfaceNil() bool {
	return dd == nil
}
