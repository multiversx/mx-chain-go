package preprocess_test

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
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

	assert.True(t, bc.IsAddressSet(address1))
	assert.True(t, bc.IsAddressSet(address2))

	bc.Init()

	assert.False(t, bc.IsAddressSet(address1))
	assert.False(t, bc.IsAddressSet(address2))
}

func TestNewBalanceComputation_SetAndGetBalanceShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	address1 := []byte("address1")
	address2 := []byte("address2")
	value1 := big.NewInt(10)
	value2 := big.NewInt(20)

	bc.SetBalanceToAddress(address1, value1)
	balance := bc.GetBalanceOfAddress(address1)
	assert.Equal(t, value1, balance)

	bc.SetBalanceToAddress(address1, value2)
	balance = bc.GetBalanceOfAddress(address1)
	assert.Equal(t, value2, balance)

	balance = bc.GetBalanceOfAddress(address2)
	assert.Nil(t, balance)
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
	balance1 := bc.GetBalanceOfAddress(address1)
	balance2 := bc.GetBalanceOfAddress(address2)

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

	subtractedWithSuccess := bc.SubBalanceFromAddress(address2, value2)
	assert.False(t, subtractedWithSuccess)
	balance := bc.GetBalanceOfAddress(address1)
	assert.Equal(t, value1, balance)

	subtractedWithSuccess = bc.SubBalanceFromAddress(address1, value2)
	assert.False(t, subtractedWithSuccess)
	balance = bc.GetBalanceOfAddress(address1)
	assert.Equal(t, value1, balance)

	subtractedWithSuccess = bc.SubBalanceFromAddress(address1, value3)
	assert.True(t, subtractedWithSuccess)
	balance = bc.GetBalanceOfAddress(address1)
	assert.Equal(t, result, balance)
}

func TestNewBalanceComputation_IsAddressSetShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	address1 := []byte("address1")
	address2 := []byte("address2")
	bc.SetBalanceToAddress(address1, big.NewInt(0))

	assert.True(t, bc.IsAddressSet(address1))
	assert.False(t, bc.IsAddressSet(address2))
}

func TestNewBalanceComputation_AddressHasEnoughBalanceShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := preprocess.NewBalanceComputation()

	value := big.NewInt(5)
	address1 := []byte("address1")
	address2 := []byte("address2")
	bc.SetBalanceToAddress(address1, value)

	assert.False(t, bc.AddressHasEnoughBalance(address2, big.NewInt(0)))
	assert.False(t, bc.AddressHasEnoughBalance(address1, big.NewInt(6)))
	assert.True(t, bc.AddressHasEnoughBalance(address1, big.NewInt(5)))
}
