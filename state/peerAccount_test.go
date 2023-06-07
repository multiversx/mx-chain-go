package state_test

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/stretchr/testify/assert"
)

func TestNewEmptyPeerAccount(t *testing.T) {
	t.Parallel()

	acc := state.NewEmptyPeerAccount()

	assert.NotNil(t, acc)
	assert.Equal(t, big.NewInt(0), acc.AccumulatedFees)
}

func TestNewPeerAccount_NilAddressContainerShouldErr(t *testing.T) {
	t.Parallel()

	acc, err := state.NewPeerAccount(nil)
	assert.True(t, check.IfNil(acc))
	assert.Equal(t, state.ErrNilAddress, err)
}

func TestNewPeerAccount_OkParamsShouldWork(t *testing.T) {
	t.Parallel()

	acc, err := state.NewPeerAccount(make([]byte, 32))
	assert.Nil(t, err)
	assert.False(t, check.IfNil(acc))
}

func TestPeerAccount_SetInvalidBLSPublicKey(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	pubKey := []byte("")

	err := acc.SetBLSPublicKey(pubKey)
	assert.Equal(t, state.ErrNilBLSPublicKey, err)
}

func TestPeerAccount_SetAndGetBLSPublicKey(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	pubKey := []byte("BLSpubKey")

	err := acc.SetBLSPublicKey(pubKey)
	assert.Nil(t, err)
	assert.Equal(t, pubKey, acc.GetBLSPublicKey())
}

func TestPeerAccount_SetRewardAddressInvalidAddress(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))

	err := acc.SetRewardAddress([]byte{})
	assert.Equal(t, state.ErrEmptyAddress, err)
}

func TestPeerAccount_SetAndGetRewardAddress(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	addr := []byte("reward address")

	_ = acc.SetRewardAddress(addr)
	assert.Equal(t, addr, acc.GetRewardAddress())
}

func TestPeerAccount_SetAndGetAccumulatedFees(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	fees := big.NewInt(10)

	acc.AddToAccumulatedFees(fees)
	assert.Equal(t, fees, acc.GetAccumulatedFees())
}

func TestPeerAccount_SetAndGetLeaderSuccessRate(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	increaseVal := uint32(5)
	decreaseVal := uint32(3)

	acc.IncreaseLeaderSuccessRate(increaseVal)
	assert.Equal(t, increaseVal, acc.GetLeaderSuccessRate().NumSuccess)

	acc.DecreaseLeaderSuccessRate(decreaseVal)
	assert.Equal(t, decreaseVal, acc.GetLeaderSuccessRate().NumFailure)
}

func TestPeerAccount_SetAndGetValidatorSuccessRate(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	increaseVal := uint32(5)
	decreaseVal := uint32(3)

	acc.IncreaseValidatorSuccessRate(increaseVal)
	assert.Equal(t, increaseVal, acc.GetValidatorSuccessRate().NumSuccess)

	acc.DecreaseValidatorSuccessRate(decreaseVal)
	assert.Equal(t, decreaseVal, acc.GetValidatorSuccessRate().NumFailure)
}

func TestPeerAccount_IncreaseAndGetSetNumSelectedInSuccessBlocks(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))

	acc.IncreaseNumSelectedInSuccessBlocks()
	assert.Equal(t, uint32(1), acc.GetNumSelectedInSuccessBlocks())
}

func TestPeerAccount_SetAndGetRating(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	rating := uint32(10)

	acc.SetRating(rating)
	assert.Equal(t, rating, acc.GetRating())
}

func TestPeerAccount_SetAndGetTempRating(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	rating := uint32(10)

	acc.SetTempRating(rating)
	assert.Equal(t, rating, acc.GetTempRating())
}

func TestPeerAccount_ResetAtNewEpoch(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	acc.AddToAccumulatedFees(big.NewInt(15))
	tempRating := uint32(5)
	acc.SetTempRating(tempRating)
	acc.IncreaseLeaderSuccessRate(2)
	acc.DecreaseLeaderSuccessRate(2)
	acc.IncreaseValidatorSuccessRate(2)
	acc.DecreaseValidatorSuccessRate(2)
	acc.IncreaseNumSelectedInSuccessBlocks()
	acc.ConsecutiveProposerMisses = 7

	acc.ResetAtNewEpoch()
	assert.Equal(t, big.NewInt(0), acc.GetAccumulatedFees())
	assert.Equal(t, tempRating, acc.GetRating())
	assert.Equal(t, uint32(0), acc.GetLeaderSuccessRate().NumSuccess)
	assert.Equal(t, uint32(0), acc.GetLeaderSuccessRate().NumFailure)
	assert.Equal(t, uint32(0), acc.GetValidatorSuccessRate().NumSuccess)
	assert.Equal(t, uint32(0), acc.GetValidatorSuccessRate().NumFailure)
	assert.Equal(t, uint32(0), acc.GetNumSelectedInSuccessBlocks())
	assert.Equal(t, uint32(7), acc.GetConsecutiveProposerMisses())
}

func TestPeerAccount_IncreaseAndGetNonce(t *testing.T) {
	t.Parallel()

	acc, _ := state.NewPeerAccount(make([]byte, 32))
	nonce := uint64(5)

	acc.IncreaseNonce(nonce)
	assert.Equal(t, nonce, acc.GetNonce())
}

func TestPeerAccount_AddressBytes(t *testing.T) {
	t.Parallel()

	address := []byte("address bytes")
	acc, _ := state.NewPeerAccount(address)

	assert.Equal(t, address, acc.AddressBytes())

	newAddress := []byte("new address bytes")
	err := acc.SetBLSPublicKey(newAddress)
	assert.Nil(t, err)
	assert.Equal(t, newAddress, acc.AddressBytes())
}
