package preprocess_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/stretchr/testify/assert"
)

func TestNewBalanceComputation_ShouldWork(t *testing.T) {
	t.Parallel()

	bc, err := preprocess.NewBalanceComputation()

	assert.False(t, check.IfNil(bc))
	assert.Nil(t, err)
}

func TestNewBalanceComputation_InitShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	address1 := []byte("address1")
	address2 := []byte("address2")
	value1 := big.NewInt(10)
	value2 := big.NewInt(20)

	bc.SetBalanceToAddress(address1, value1)
	bc.SetBalanceToAddress(address2, value2)

	assert.True(t, bc.HasAddressBalanceSet(address1))
	assert.True(t, bc.HasAddressBalanceSet(address2))

	bc.Init()

	assert.False(t, bc.HasAddressBalanceSet(address1))
	assert.False(t, bc.HasAddressBalanceSet(address2))
}

func TestNewBalanceComputation_SetAndGetBalanceShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	address1 := []byte("address1")
	address2 := []byte("address2")
	value1 := big.NewInt(10)
	value2 := big.NewInt(20)

	bc.SetBalanceToAddress(address1, value1)
	balance, _ := bc.GetBalanceOfAddress(address1)
	assert.Equal(t, value1, balance)

	bc.SetBalanceToAddress(address1, value2)
	balance, _ = bc.GetBalanceOfAddress(address1)
	assert.Equal(t, value2, balance)

	balance, err := bc.GetBalanceOfAddress(address2)
	assert.Nil(t, balance)
	assert.Equal(t, process.ErrAddressHasNoBalanceSet, err)
}

func TestNewBalanceComputation_AddBalanceToAddressShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	address1 := []byte("address1")
	address2 := []byte("address2")
	value1 := big.NewInt(10)
	value2 := big.NewInt(20)
	result := big.NewInt(0).Add(value1, value2)

	bc.SetBalanceToAddress(address1, value1)

	addedWithSuccess1 := bc.AddBalanceToAddress(address1, value2)
	addedWithSuccess2 := bc.AddBalanceToAddress(address2, value2)
	balance1, _ := bc.GetBalanceOfAddress(address1)
	balance2, _ := bc.GetBalanceOfAddress(address2)

	assert.True(t, addedWithSuccess1)
	assert.False(t, addedWithSuccess2)
	assert.Equal(t, result, balance1)
	assert.Nil(t, balance2)
}

func TestNewBalanceComputation_SubBalanceFromAddressShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	address1 := []byte("address1")
	address2 := []byte("address2")
	value1 := big.NewInt(10)
	value2 := big.NewInt(20)
	value3 := big.NewInt(5)
	result := big.NewInt(0).Sub(value1, value3)

	bc.SetBalanceToAddress(address1, value1)

	addressExists, subtractedWithSuccess := bc.SubBalanceFromAddress(address2, value2)
	assert.False(t, addressExists)
	assert.False(t, subtractedWithSuccess)
	balance, _ := bc.GetBalanceOfAddress(address1)
	assert.Equal(t, value1, balance)

	addressExists, subtractedWithSuccess = bc.SubBalanceFromAddress(address1, value2)
	assert.True(t, addressExists)
	assert.False(t, subtractedWithSuccess)
	balance, _ = bc.GetBalanceOfAddress(address1)
	assert.Equal(t, value1, balance)

	addressExists, subtractedWithSuccess = bc.SubBalanceFromAddress(address1, value3)
	assert.True(t, addressExists)
	assert.True(t, subtractedWithSuccess)
	balance, _ = bc.GetBalanceOfAddress(address1)
	assert.Equal(t, result, balance)
}

func TestNewBalanceComputation_HasAddressBalanceSetShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	address1 := []byte("address1")
	address2 := []byte("address2")
	bc.SetBalanceToAddress(address1, big.NewInt(0))

	assert.True(t, bc.HasAddressBalanceSet(address1))
	assert.False(t, bc.HasAddressBalanceSet(address2))
}

func TestNewBalanceComputation_IsBalanceInAddressShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	value := big.NewInt(5)
	address1 := []byte("address1")
	address2 := []byte("address2")
	bc.SetBalanceToAddress(address1, value)

	assert.False(t, bc.IsBalanceInAddress(address2, big.NewInt(0)))
	assert.False(t, bc.IsBalanceInAddress(address1, big.NewInt(6)))
	assert.True(t, bc.IsBalanceInAddress(address1, big.NewInt(5)))
}
