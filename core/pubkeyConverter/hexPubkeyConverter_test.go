package pubkeyConverter_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/data/state"
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
	assert.Equal(t, addressLen, hpc.Len())
}

func TestHexPubkeyConverter_DecodeShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 4
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	buff, err := hpc.Decode("aaff")
	assert.True(t, errors.Is(err, state.ErrWrongSize))
	assert.Equal(t, 0, len(buff))

	buff, err = hpc.Decode("not a hex")
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(buff))
}

func TestHexPubkeyConverter_DecodeShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 2
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	buff, err := hpc.Decode("aaff")

	assert.Nil(t, err)
	assert.Equal(t, []byte{170, 255}, buff)
}

func TestHexPubkeyConverter_EncodeShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 4
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	str := hpc.Encode([]byte{170, 255})
	assert.Equal(t, "aaff", str)
}

func TestHexPubkeyConverter_EncodeDecodeShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 16
	hpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	value := "123456789012345678901234567890af"
	buff, err := hpc.Decode(value)
	assert.Nil(t, err)

	revoveredValue := hpc.Encode(buff)

	assert.Equal(t, value, revoveredValue)
}
