package addressConverters_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewPlainAddressConverter

func TestNewPlainAddressConverter_NegativeSizeShouldErr(t *testing.T) {
	t.Parallel()

	_, err := addressConverters.NewPlainAddressConverter(-1, "")
	assert.Equal(t, state.ErrNegativeValue, err)
}

func TestNewPlainAddressConverter_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	_, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)
}

//------- CreateAddressFromPublicKeyBytes

func TestPlainAddressConverter_CreateAddressFromPublicKeyBytesNilBytesShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)
	_, err = ac.CreateAddressFromPublicKeyBytes(nil)
	assert.Equal(t, state.ErrNilPubKeysBytes, err)
}

func TestPlainAddressConverter_CreateAddressFromPublicKeyBytesLenLowerAddrlenShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)

	for i := 0; i < 32; i++ {
		_, err = ac.CreateAddressFromPublicKeyBytes(make([]byte, i))
		assert.NotNil(t, err)
		fmt.Printf("%v\n", err)
	}
}

func TestPlainAddressConverter_CreateAddressFromPublicKeyBytesOkValsShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)

	//generating a random byte slice with len > AdrLen
	buff := make([]byte, 32)
	_, err = rand.Read(buff)
	assert.Nil(t, err)

	adr, err := ac.CreateAddressFromPublicKeyBytes(buff)

	assert.Nil(t, err)
	assert.Equal(t, buff, adr.Bytes())
}

func TestPlainAddressConverter_CreateAddressFromPublicKeyBytesOkValsTrimShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)

	//generating a random byte slice with len > AdrLen
	buff := make([]byte, 320)
	_, err = rand.Read(buff)
	assert.Nil(t, err)

	adr, err := ac.CreateAddressFromPublicKeyBytes(buff)

	assert.Nil(t, err)
	assert.Equal(t, buff[320-32:], adr.Bytes())
}

//------- ConvertToHex

func TestPlainAddressConverter_ConvertToHexNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)

	_, err = ac.ConvertToHex(nil)
	assert.Equal(t, state.ErrNilAddressContainer, err)
}

func TestPlainAddressConverter_ConvertToHexOkValsShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "0x")
	assert.Nil(t, err)

	adr := mock.NewAddressDummy([]byte{255}, []byte{255})

	str, err := ac.ConvertToHex(adr)
	assert.Nil(t, err)
	assert.Equal(t, "0xff", strings.ToLower(str))
}

//------- CreateAddressFromHex

func TestPlainAddressConverter_CreateAddressFromHexEmptyHexAddressShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)

	_, err = ac.CreateAddressFromHex("")
	assert.Equal(t, state.ErrEmptyAddress, err)
}

func TestPlainAddressConverter_CreateAddressFromHexEmptyBadLengthShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)

	_, err = ac.CreateAddressFromHex("A")
	assert.NotNil(t, err)
}

func TestPlainAddressConverter_CreateAddressFromHexEmptyBadStringShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "")
	assert.Nil(t, err)

	adr := ""
	for i := 0; i < 64; i++ {
		adr = adr + "t"
	}

	_, err = ac.CreateAddressFromHex(adr)
	assert.NotNil(t, err)
}

func TestPlainAddressConverter_FromHexAddressValidDataWithPrefixShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "0x")
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

func TestPlainAddressConverter_PrepareAddressBytesNilAddrShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "0x")
	assert.Nil(t, err)

	_, err = ac.PrepareAddressBytes(nil)
	assert.Equal(t, state.ErrNilAddressContainer, err)
}

func TestPlainAddressConverter_PrepareAddressBytesEmptyAddrShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "0x")
	assert.Nil(t, err)

	_, err = ac.PrepareAddressBytes(make([]byte, 0))
	assert.Equal(t, state.ErrEmptyAddress, err)
}

func TestPlainAddressConverter_PrepareAddressBytesWrongSizeShouldErr(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "0x")
	assert.Nil(t, err)

	_, err = ac.PrepareAddressBytes(make([]byte, 31))
	assert.NotNil(t, err)
}

func TestPlainAddressConverter_PrepareAddressBytesOkValsShouldWork(t *testing.T) {
	t.Parallel()

	ac, err := addressConverters.NewPlainAddressConverter(32, "0x")
	assert.Nil(t, err)

	//generating a random byte slice
	buff := make([]byte, 32)
	_, err = rand.Read(buff)
	assert.Nil(t, err)

	checked, err := ac.PrepareAddressBytes(buff)
	assert.Nil(t, err)
	assert.Equal(t, buff, checked)
}
