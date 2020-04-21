package genesis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelegationData_MarshalUnmarshalEmptyStruct(t *testing.T) {
	t.Parallel()

	input := &DelegationData{
		Address: "",
		Value:   nil,
	}
	expected := &DelegationData{
		Address: "",
		Value:   big.NewInt(0),
	}

	buff, err := json.Marshal(input)
	assert.Nil(t, err)

	recovered := &DelegationData{}
	err = json.Unmarshal(buff, recovered)

	assert.Nil(t, err)
	assert.Equal(t, expected, recovered)
}

func TestDelegationData_MarshalUnmarshalValidStruct(t *testing.T) {
	t.Parallel()

	address := "address"
	value := int64(2242)
	input := &DelegationData{
		Address: address,
		Value:   big.NewInt(value),
	}

	buff, err := json.Marshal(input)
	assert.Nil(t, err)

	recovered := &DelegationData{}
	err = json.Unmarshal(buff, recovered)

	assert.Nil(t, err)
	assert.Equal(t, input, recovered)
}

func TestDelegationData_UnmarshalNotAValidValueShouldErr(t *testing.T) {
	t.Parallel()

	address := "address"
	value := int64(2242)
	input := &DelegationData{
		Address: address,
		Value:   big.NewInt(value),
	}

	buff, _ := json.Marshal(input)
	buff = bytes.Replace(buff, []byte(fmt.Sprintf("%d", value)), []byte("not a number"), 1)

	recovered := &DelegationData{}
	err := json.Unmarshal(buff, recovered)

	assert.True(t, errors.Is(err, ErrInvalidDelegationValueString))
}
