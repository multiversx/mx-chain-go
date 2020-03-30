package pubkeyConverter_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/pubkeyConverter"
	"github.com/stretchr/testify/assert"
)

func TestNewHexPubkeyConverter_InvalidSizeShouldErr(t *testing.T) {
	t.Parallel()

	hpc, err := pubkeyConverter.NewHexPubkeyConverter(-1)
	assert.True(t, errors.Is(err, state.ErrInvalidAddressLength))
	assert.True(t, check.IfNil(hpc))

	hpc, err = pubkeyConverter.NewHexPubkeyConverter(0)
	assert.True(t, errors.Is(err, state.ErrInvalidAddressLength))
	assert.True(t, check.IfNil(hpc))

	hpc, err = pubkeyConverter.NewHexPubkeyConverter(3)
	assert.True(t, errors.Is(err, state.ErrInvalidAddressLength))
	assert.True(t, check.IfNil(hpc))
}

func TestNewHexPubkeyConverter_ShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 28
	hpc, err := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(hpc))
	assert.Equal(t, addressLen, hpc.AddressLen())
}

func TestHexPubkeyConverter_BytesShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 4
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	buff, err := hpc.Bytes("aaff")
	assert.True(t, errors.Is(err, state.ErrWrongSize))
	assert.Equal(t, 0, len(buff))

	buff, err = hpc.Bytes("not a hex")
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(buff))
}

func TestHexPubkeyConverter_BytesShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 2
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	buff, err := hpc.Bytes("aaff")

	assert.Nil(t, err)
	assert.Equal(t, []byte{170, 255}, buff)
}

func TestHexPubkeyConverter_StringShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 4
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	str, err := hpc.String([]byte{170, 255})
	assert.Nil(t, err)
	assert.Equal(t, "aaff", str)
}

func TestHexPubkeyConverter_StringBytesShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 16
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	value := "123456789012345678901234567890af"
	buff, err := hpc.Bytes(value)
	assert.Nil(t, err)

	revoveredValue, _ := hpc.String(buff)

	assert.Equal(t, value, revoveredValue)
}

func TestHexPubkeyConverter_CreateAddressFromBytesInvalidSizeShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 32
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	adr, err := hpc.CreateAddressFromBytes([]byte{0, 1})

	assert.True(t, check.IfNil(adr))
	assert.True(t, errors.Is(err, state.ErrWrongSize))
}

func TestHexPubkeyConverter_CreateAddressFromBytesShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 4
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	buff := []byte{0, 1, 2, 3}
	adr, err := hpc.CreateAddressFromBytes(buff)

	assert.Nil(t, err)
	assert.Equal(t, buff, adr.Bytes())
}

func TestHexPubkeyConverter_CreateAddressFromStringInvalidHexShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 32
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	adr, err := hpc.CreateAddressFromString("not a hex")

	assert.True(t, check.IfNil(adr))
	assert.NotNil(t, err)
}

func TestHexPubkeyConverter_CreateAddressFromStringShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 2
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	adr, err := hpc.CreateAddressFromString("aaff")

	assert.Nil(t, err)
	assert.Equal(t, []byte{170, 255}, adr.Bytes())
}
