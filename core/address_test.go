package core

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_isSmartContractAddress(t *testing.T) {
	t.Parallel()

	address, _ := hex.DecodeString("12345")
	assert.False(t, IsSmartContractAddress(address))

	address, _ = hex.DecodeString("000000000001000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	assert.False(t, IsSmartContractAddress(address))

	scaddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	assert.True(t, IsSmartContractAddress(scaddress))

	emAddress, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	assert.True(t, IsSmartContractAddress(emAddress))
}

func TestAddress_IsMetachainIdentifier(t *testing.T) {
	t.Parallel()

	identifier := make([]byte, 0)
	assert.False(t, IsMetachainIdentifier(identifier))

	identifier, _ = hex.DecodeString("001234567801000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	assert.False(t, IsMetachainIdentifier(identifier))

	identifier, _ = hex.DecodeString("ff")
	assert.True(t, IsMetachainIdentifier(identifier))
}

func TestIsSmartContractOnMetachain_ScAddressTooShortShouldReturnFalse(t *testing.T) {
	t.Parallel()

	identifier, _ := hex.DecodeString("ff")
	scAddress, _ := hex.DecodeString("000000000000eb3a1")
	assert.False(t, IsSmartContractOnMetachain(identifier, scAddress))
}

func TestIsSmartContractOnMetachain_IdentifierNotMetaShouldReturnFalse(t *testing.T) {
	t.Parallel()

	identifier, _ := hex.DecodeString("aa")
	scAddress, _ := hex.DecodeString("00000000000000000000000000000022cd8429ce92f8973bba2a9fb51e0eb3a1")
	assert.False(t, IsSmartContractOnMetachain(identifier, scAddress))
}

func TestIsSmartContractOnMetachain_AddressIsNotSCShouldReturnFalse(t *testing.T) {
	t.Parallel()

	identifier, _ := hex.DecodeString("ff")
	scAddress, _ := hex.DecodeString("000000000001000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	assert.False(t, IsSmartContractOnMetachain(identifier, scAddress))
}

func TestIsSmartContractOnMetachain_ShouldReturnFalse(t *testing.T) {
	t.Parallel()

	identifier, _ := hex.DecodeString("ff")
	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	assert.False(t, IsSmartContractOnMetachain(identifier, scAddress))
}

func TestIsSmartContractOnMetachain_ShouldReturnTrue(t *testing.T) {
	t.Parallel()

	identifier, _ := hex.DecodeString("ff")
	scAddress, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000b51e0eb3a1")
	assert.True(t, IsSmartContractOnMetachain(identifier, scAddress))
}
