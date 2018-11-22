package state_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

type hasherMock127 struct {
}

func (hasherMock127) Compute(s string) []byte {
	buff := make([]byte, 0)

	var i byte
	for i = 0; i < 127; i++ {
		buff = append(buff, i)
	}

	return buff
}

func (hasherMock127) EmptyHash() []byte {
	return nil
}

func (hasherMock127) Size() int {
	return 64
}

func TestAddress_NewAddress_InvalidData_ShouldErr(t *testing.T) {
	t.Parallel()

	//nil address
	_, err := state.NewAddress(nil)
	assert.NotNil(t, err)
	fmt.Printf("Error: %v\n", err.Error())

	//empty address
	_, err = state.NewAddress(make([]byte, 0))
	assert.NotNil(t, err)
	fmt.Printf("Error: %v\n", err.Error())

	//wrong size address
	for i := 1; i < state.AdrLen*2; i++ {
		if i == state.AdrLen {
			//for this case, it should work so skipping for now
			continue
		}

		_, err = state.NewAddress(make([]byte, i))
		assert.NotNil(t, err)
		fmt.Printf("Error: %v\n", err.Error())
	}
}

func TestAddress_NewAddress_ValidData_ShouldWork(t *testing.T) {
	t.Parallel()

	address, err := state.NewAddress(make([]byte, state.AdrLen))
	assert.Nil(t, err)
	assert.NotNil(t, address)
	assert.Equal(t, address.Bytes(), make([]byte, state.AdrLen))
}

func TestAddress_Hex_ShouldWork(t *testing.T) {
	t.Parallel()

	//generating a random byte slice
	buff := make([]byte, state.AdrLen)
	_, err := rand.Read(buff)
	assert.Nil(t, err)

	address, err := state.NewAddress(buff)

	fmt.Printf("Address is: %s\n", address.Hex())
	assert.Equal(t, state.HexPrefix+hex.EncodeToString(buff), address.Hex())
}

func TestAddress_FromHexAddress_InvalidData_ShouldErr(t *testing.T) {
	t.Parallel()

	//empty string
	_, err := state.FromHexAddress("")
	assert.NotNil(t, err)

	//invalid characters
	adr := ""
	for i := 0; i < state.AdrLen*2; i++ {
		adr = adr + "t"
	}
	_, err = state.FromHexAddress(adr)
	assert.NotNil(t, err)

	//odd characters length
	_, err = state.FromHexAddress("a0f")
	assert.NotNil(t, err)

	//more characters than needed
	adr = ""
	for i := 0; i < state.AdrLen*2+4; i++ {
		adr = adr + "a"
	}

	//less characters than needed
	_, err = state.FromHexAddress("aa")
	assert.NotNil(t, err)
}

func TestAddress_FromHexAddress_ValidData_ShouldWork(t *testing.T) {
	t.Parallel()

	//generating a random byte slice
	buff := make([]byte, state.AdrLen)
	_, err := rand.Read(buff)
	assert.Nil(t, err)

	//extract from string with prefix
	adrPrefix, err := state.FromHexAddress(state.HexPrefix + hex.EncodeToString(buff))
	assert.Nil(t, err)

	//extract from string without prefix
	adrNoPrefix, err := state.FromHexAddress(hex.EncodeToString(buff))
	assert.Nil(t, err)

	//both should be the same
	assert.Equal(t, adrPrefix.Bytes(), adrNoPrefix.Bytes())
}

func TestAddress_Hash_ValidData_ShouldWork(t *testing.T) {
	t.Parallel()

	//generating a random byte slice
	buff := make([]byte, state.AdrLen)
	_, err := rand.Read(buff)
	assert.Nil(t, err)

	adr, err := state.NewAddress(buff)
	assert.Nil(t, err)

	h := mock.HasherMock{}
	hashExpected := h.Compute(string(buff))

	assert.Equal(t, hashExpected, adr.Hash(h))
}

func TestAddress_FromPubKeyBytes_InvalidData_ShouldErr(t *testing.T) {
	t.Parallel()

	//nil bytes
	_, err := state.FromPubKeyBytes(nil, mock.HasherMock{})
	assert.NotNil(t, err)

	//nil hasher
	_, err = state.FromPubKeyBytes(make([]byte, state.AdrLen), nil)
	assert.NotNil(t, err)

	//pub key length is lower than AdrLen
	for i := 0; i < state.AdrLen; i++ {
		_, err = state.FromPubKeyBytes(make([]byte, i), mock.HasherMock{})
		assert.NotNil(t, err)
	}
}

func TestAddress_FromPubKeyBytes_ValidDataAdrLen_ShouldWork(t *testing.T) {
	t.Parallel()

	//generating a random byte slice with len > AdrLen
	buff := make([]byte, state.AdrLen*10)
	_, err := rand.Read(buff)
	assert.Nil(t, err)

	//compute expected hash
	h := mock.HasherMock{}
	hashExpected := h.Compute(string(buff))

	adr, err := state.FromPubKeyBytes(buff, h)

	if state.AdrLen == len(hashExpected) {
		assert.Equal(t, hashExpected, adr.Bytes())
	} else {
		assert.Equal(t, hashExpected[len(hashExpected)-state.AdrLen:], adr.Bytes())
	}

}

func TestAddress_FromPubKeyBytes_ValidDataTrim_ShouldWork(t *testing.T) {
	t.Parallel()

	//generating a random byte slice with len > AdrLen
	buff := make([]byte, state.AdrLen*10)
	_, err := rand.Read(buff)
	assert.Nil(t, err)

	//compute expected hash
	h := hasherMock127{}
	hashExpected := h.Compute(string(buff))

	adr, err := state.FromPubKeyBytes(buff, h)

	if state.AdrLen == len(hashExpected) {
		assert.Equal(t, hashExpected, adr.Bytes())
	} else {
		assert.Equal(t, hashExpected[len(hashExpected)-state.AdrLen:], adr.Bytes())
	}

}
