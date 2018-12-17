package state_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewAddressConverter

func TestNewAddressConverterNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	_, err := state.NewHashAddressConverter(nil, 1, "")
	assert.Equal(t, state.ErrNilHasher, err)
}

func TestNewAddressConverterNegativeSizeShouldErr(t *testing.T) {
	t.Parallel()

	_, err := state.NewHashAddressConverter(&mock.HasherMock{}, -1, "")
	assert.Equal(t, state.ErrNegativeValue, err)
}

func TestNewAddressConverterOkValsShouldWork(t *testing.T) {
	t.Parallel()

	_, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "")
	assert.Nil(t, err)
}

//------- CreateAddressFromPublicKeyBytes

func TestAddressConverterCreateAddressFromPublicKeyBytesNilBytesShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "")
	assert.Nil(t, err)
	_, err = ac.CreateAddressFromPublicKeyBytes(nil)
	assert.Equal(t, state.ErrNilPubKeysBytes, err)
}

func TestAddressConverterCreateAddressFromPublicKeyBytesLenLowerAddrlenShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "")
	assert.Nil(t, err)

	for i := 0; i < 32; i++ {
		_, err = ac.CreateAddressFromPublicKeyBytes(make([]byte, i))
		assert.NotNil(t, err)
		fmt.Printf("%v\n", err)
	}
}

func TestAddressConverterCreateAddressFromPublicKeyBytesOkValsShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "")
	assert.Nil(t, err)

	//generating a random byte slice with len > AdrLen
	buff := make([]byte, 320)
	_, err = rand.Read(buff)
	assert.Nil(t, err)

	//compute expected hash
	h := mock.HasherMock{}
	hashExpected := h.Compute(string(buff))

	adr, err := ac.CreateAddressFromPublicKeyBytes(buff)

	assert.Equal(t, hashExpected, adr.Bytes())
}

func TestAddressConverterCreateAddressFromPublicKeyBytesOkValsTrimShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock127{}, 32, "")
	assert.Nil(t, err)

	//generating a random byte slice with len > AdrLen
	buff := make([]byte, 320)
	_, err = rand.Read(buff)
	assert.Nil(t, err)

	//compute expected hash
	h := mock.HasherMock127{}
	hashExpected := h.Compute(string(buff))

	adr, err := ac.CreateAddressFromPublicKeyBytes(buff)

	assert.Equal(t, hashExpected[len(hashExpected)-32:], adr.Bytes())

	fmt.Printf("Address %v has the hash %v\n", hex.EncodeToString(adr.Bytes()), hex.EncodeToString(mock.HasherMock{}.Compute(string(adr.Bytes()))))
}

//------- ConvertToHex

func TestAddressConverterConvertToHexNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock127{}, 32, "")
	assert.Nil(t, err)

	_, err = ac.ConvertToHex(nil)
	assert.Equal(t, state.ErrNilAddressContainer, err)
}

func TestAddressConverterConvertToHexOkValsShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "0x")
	assert.Nil(t, err)

	adr := mock.NewAddressDummy([]byte{255}, []byte{255})

	str, err := ac.ConvertToHex(adr)
	assert.Nil(t, err)
	assert.Equal(t, "0xff", strings.ToLower(str))
}

//------- CreateAddressFromHex

func TestAddressConverterCreateAddressFromHexEmptyHexAddressShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "")
	assert.Nil(t, err)

	_, err = ac.CreateAddressFromHex("")
	assert.Equal(t, state.ErrEmptyAddress, err)
}

func TestAddressConverterCreateAddressFromHexEmptyBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "")
	assert.Nil(t, err)

	_, err = ac.CreateAddressFromHex("A")
	assert.NotNil(t, err)
}

func TestAddressConverterCreateAddressFromHexEmptyBadStringShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "")
	assert.Nil(t, err)

	adr := ""
	for i := 0; i < 64; i++ {
		adr = adr + "t"
	}

	_, err = ac.CreateAddressFromHex(adr)
	assert.NotNil(t, err)
}

func TestAddressConverterFromHexAddressValidDataWithPrefixShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "0x")
	assert.Nil(t, err)

	//generating a random byte slice
	buff := make([]byte, 32)
	_, err = rand.Read(buff)
	assert.Nil(t, err)

	str := "0x" + hex.EncodeToString(buff)

	adr, err := ac.CreateAddressFromHex(str)
	assert.Nil(t, err)
	assert.Equal(t, buff, adr.Bytes())
}

//------- PrepareAddressBytes

func TestAddressConverterPrepareAddressBytesNilAddrShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "0x")
	assert.Nil(t, err)

	_, err = ac.PrepareAddressBytes(nil)
	assert.Equal(t, state.ErrNilAddressContainer, err)
}

func TestAddressConverterPrepareAddressBytesEmptyAddrShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "0x")
	assert.Nil(t, err)

	_, err = ac.PrepareAddressBytes(make([]byte, 0))
	assert.Equal(t, state.ErrEmptyAddress, err)
}

func TestAddressConverterPrepareAddressBytesWrongSizeShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "0x")
	assert.Nil(t, err)

	_, err = ac.PrepareAddressBytes(make([]byte, 31))
	assert.NotNil(t, err)
}

func TestAddressConverterPrepareAddressBytesOkValsShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := state.NewHashAddressConverter(&mock.HasherMock{}, 32, "0x")
	assert.Nil(t, err)

	//generating a random byte slice
	buff := make([]byte, 32)
	_, err = rand.Read(buff)
	assert.Nil(t, err)

	checked, err := ac.PrepareAddressBytes(buff)
	assert.Nil(t, err)
	assert.Equal(t, buff, checked)
}
