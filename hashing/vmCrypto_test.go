package hashing_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestVMCrypto_Sha256ReturnsASha256Hash(t *testing.T) {
	t.Parallel()

	vmc := &hashing.VMCrypto{}

	inputString := "input string for hashing"
	expectedHash := sha256.Sha256{}.Compute(inputString)
	resultedHash, err := vmc.Sha256(inputString)

	assert.Nil(t, err)
	assert.Equal(t, string(expectedHash), resultedHash)
}

func TestVMCrypto_Keccak256ReturnsAKeccak256Hash(t *testing.T) {
	t.Parallel()

	vmc := &hashing.VMCrypto{}

	inputString := "input string for hashing"
	expectedHash := keccak.Keccak{}.Compute(inputString)
	resultedHash, err := vmc.Keccak256(inputString)

	assert.Nil(t, err)
	assert.Equal(t, string(expectedHash), resultedHash)
}

func TestVMCrypto_DeprecatedFuncsReturnErr(t *testing.T) {
	t.Parallel()

	vmc := &hashing.VMCrypto{}

	_, err := vmc.Ripemd160("")
	assert.Equal(t, hashing.ErrNotImplemented, err)

	_, err = vmc.EcdsaRecover("", nil, "", "")
	assert.Equal(t, hashing.ErrNotImplemented, err)

	_, err = vmc.Bn128valid(vmcommon.Bn128Point{})
	assert.Equal(t, hashing.ErrNotImplemented, err)

	_, err = vmc.Bn128g2valid(vmcommon.Bn128G2Point{})
	assert.Equal(t, hashing.ErrNotImplemented, err)

	_, err = vmc.Bn128add(vmcommon.Bn128Point{}, vmcommon.Bn128Point{})
	assert.Equal(t, hashing.ErrNotImplemented, err)

	_, err = vmc.Bn128mul(nil, vmcommon.Bn128Point{})
	assert.Equal(t, hashing.ErrNotImplemented, err)

	_, err = vmc.Bn128ate(nil, nil)
	assert.Equal(t, hashing.ErrNotImplemented, err)
}
