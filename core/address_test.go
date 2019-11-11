package core

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddress_isSmartContractAddress(t *testing.T) {
	t.Parallel()

	address, _ := hex.DecodeString("000000000001000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	assert.False(t, IsSmartContractAddress(address))

	scaddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	assert.True(t, IsSmartContractAddress(scaddress))

	emAddress, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	assert.True(t, IsSmartContractAddress(emAddress))
}
