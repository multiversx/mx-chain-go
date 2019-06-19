package hooks_test

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestVMCrypto_Sha256ReturnsASha256Hash(t *testing.T) {
	t.Parallel()

	vmch := &hooks.VMCryptoHook{}

	inputString := "input string for hashing"
	expectedHash := hex.EncodeToString(sha256.Sha256{}.Compute(inputString))
	resultedHash, err := vmch.Sha256(inputString)

	assert.Nil(t, err)
	assert.Equal(t, expectedHash, resultedHash)
}

func TestVMCrypto_Keccak256ReturnsAKeccak256Hash(t *testing.T) {
	t.Parallel()

	vmch := &hooks.VMCryptoHook{}

	inputString := "input string for hashing"
	expectedHash := hex.EncodeToString(keccak.Keccak{}.Compute(inputString))
	resultedHash, err := vmch.Keccak256(inputString)

	assert.Nil(t, err)
	assert.Equal(t, expectedHash, resultedHash)
}

func TestVMCrypto_DeprecatedFuncsReturnErr(t *testing.T) {
	t.Parallel()

	vmch := &hooks.VMCryptoHook{}

	_, err := vmch.Ripemd160("")
	assert.Equal(t, hooks.ErrNotImplemented, err)

	_, err = vmch.EcdsaRecover("", nil, "", "")
	assert.Equal(t, hooks.ErrNotImplemented, err)

	_, err = vmch.Bn128valid(vmcommon.Bn128Point{})
	assert.Equal(t, hooks.ErrNotImplemented, err)

	_, err = vmch.Bn128g2valid(vmcommon.Bn128G2Point{})
	assert.Equal(t, hooks.ErrNotImplemented, err)

	_, err = vmch.Bn128add(vmcommon.Bn128Point{}, vmcommon.Bn128Point{})
	assert.Equal(t, hooks.ErrNotImplemented, err)

	_, err = vmch.Bn128mul(nil, vmcommon.Bn128Point{})
	assert.Equal(t, hooks.ErrNotImplemented, err)

	_, err = vmch.Bn128ate(nil, nil)
	assert.Equal(t, hooks.ErrNotImplemented, err)
}
