package pubkeyConverter_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/pubkeyConverter"
	"github.com/stretchr/testify/assert"
)

func TestNewBech32PubkeyConverter_InvalidSizeShouldErr(t *testing.T) {
	t.Parallel()

	bpc, err := pubkeyConverter.NewBech32PubkeyConverter(-1)
	assert.True(t, errors.Is(err, state.ErrInvalidAddressLength))
	assert.True(t, check.IfNil(bpc))

	bpc, err = pubkeyConverter.NewBech32PubkeyConverter(0)
	assert.True(t, errors.Is(err, state.ErrInvalidAddressLength))
	assert.True(t, check.IfNil(bpc))

	bpc, err = pubkeyConverter.NewBech32PubkeyConverter(3)
	assert.True(t, errors.Is(err, state.ErrInvalidAddressLength))
	assert.True(t, check.IfNil(bpc))
}

func TestNewBech32PubkeyConverter_ShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 28
	bpc, err := pubkeyConverter.NewBech32PubkeyConverter(addressLen)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(bpc))
	assert.Equal(t, addressLen, bpc.Len())
}

func TestBech32PubkeyConverter_BytesInvalidStringShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 32
	bpc, _ := pubkeyConverter.NewBech32PubkeyConverter(addressLen)

	str, err := bpc.Bytes("not a bech32 string")

	assert.Equal(t, 0, len(str))
	assert.NotNil(t, err)
}

func TestBech32PubkeyConverter_BytesPrefixMismatchShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 32
	bpc, _ := pubkeyConverter.NewBech32PubkeyConverter(addressLen)

	str, err := bpc.Bytes("err1xyerxdp4xcmnswfsxyerxdp4xcmnswfsxyerxdp4xcmnswfsxyeqnyphvl")

	assert.Equal(t, 0, len(str))
	assert.True(t, errors.Is(err, state.ErrInvalidErdAddress))
}

func TestBech32PubkeyConverter_BytesWrongSizeShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 32
	bpc, _ := pubkeyConverter.NewBech32PubkeyConverter(addressLen)

	str, err := bpc.Bytes("erd1xyerxdp4xcmnswfsxyeqqzq40r")

	assert.Equal(t, 0, len(str))
	assert.True(t, errors.Is(err, state.ErrWrongSize))
}

func TestBech32PubkeyConverter_StringBytesShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 32
	bpc, _ := pubkeyConverter.NewBech32PubkeyConverter(addressLen)

	buff := []byte("12345678901234567890123456789012")
	str, err := bpc.String(buff)

	assert.Nil(t, err)
	assert.Equal(t, 0, strings.Index(str, pubkeyConverter.Prefix))

	fmt.Printf("generated address: %s\n", str)

	recoveredBuff, err := bpc.Bytes(str)

	assert.Nil(t, err)
	assert.Equal(t, buff, recoveredBuff)
}

func TestBech32PubkeyConverter_CreateAddressFromBytesInvalidSizeShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 32
	bpc, _ := pubkeyConverter.NewBech32PubkeyConverter(addressLen)

	adr, err := bpc.CreateAddressFromBytes([]byte{0, 1})

	assert.True(t, check.IfNil(adr))
	assert.True(t, errors.Is(err, state.ErrWrongSize))
}

func TestBech32PubkeyConverter_CreateAddressFromBytesShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 4
	bpc, _ := pubkeyConverter.NewBech32PubkeyConverter(addressLen)

	buff := []byte{0, 1, 2, 3}
	adr, err := bpc.CreateAddressFromBytes(buff)

	assert.Nil(t, err)
	assert.Equal(t, buff, adr.Bytes())
}

func TestBech32PubkeyConverter_CreateAddressFromStringInvalidStringShouldErr(t *testing.T) {
	t.Parallel()

	addressLen := 32
	bpc, _ := pubkeyConverter.NewHexPubkeyConverter(addressLen)

	adr, err := bpc.CreateAddressFromString("not a hex")

	assert.True(t, check.IfNil(adr))
	assert.NotNil(t, err)
}

func TestBech32PubkeyConverter_CreateAddressFromStringShouldWork(t *testing.T) {
	t.Parallel()

	addressLen := 32
	bpc, _ := pubkeyConverter.NewBech32PubkeyConverter(addressLen)

	adr, err := bpc.CreateAddressFromString("erd1xyerxdp4xcmnswfsxyerxdp4xcmnswfsxyerxdp4xcmnswfsxyeqlrqt99")
	expectedBuff := []byte("12345678901234567890123456789012")

	assert.Nil(t, err)
	assert.Equal(t, expectedBuff, adr.Bytes())
}
