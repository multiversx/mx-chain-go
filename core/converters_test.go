package core_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/assert"
)

func TestToB64ShouldReturnNil(t *testing.T) {
	val := core.ToB64(nil)
	assert.Equal(t, "<NIL>", val)
}

func TestToB64ShouldWork(t *testing.T) {
	buff := []byte("test")
	val := core.ToB64(buff)
	assert.Equal(t, base64.StdEncoding.EncodeToString(buff), val)
}

func TestToHexShouldReturnNil(t *testing.T) {
	val := core.ToHex(nil)
	assert.Equal(t, "<NIL>", val)
}

func TestToHexShouldWork(t *testing.T) {
	buff := []byte("test")
	val := core.ToHex(buff)
	assert.Equal(t, hex.EncodeToString(buff), val)
}

func TestCalculateHash_NilMarshalizer(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	hash, err := core.CalculateHash(nil, &mock.HasherMock{}, obj)
	assert.Nil(t, hash)
	assert.Equal(t, core.ErrNilMarshalizer, err)
}

func TestCalculateHash_NilHasher(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	hash, err := core.CalculateHash(&mock.MarshalizerMock{}, nil, obj)
	assert.Nil(t, hash)
	assert.Equal(t, core.ErrNilHasher, err)
}

func TestCalculateHash_ErrMarshalizer(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	marshalizer := &mock.MarshalizerMock{
		Fail: true,
	}
	hash, err := core.CalculateHash(marshalizer, &mock.HasherMock{}, obj)
	assert.Nil(t, hash)
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestCalculateHash_NilObject(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hash, err := core.CalculateHash(marshalizer, &mock.HasherMock{}, nil)
	assert.Nil(t, hash)
	assert.Equal(t, mock.ErrNilObjectToMarshal, err)
}

func TestCalculateHash_Good(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	results := []byte{0x44, 0xa7, 0x94, 0xc0, 0x7e, 0x54, 0x30, 0x79, 0x7a, 0xb5, 0xc6, 0xf, 0xdc, 0x57, 0x9c, 0x44, 0xff, 0x8b, 0xdc, 0x3d, 0xa0, 0x64, 0xdd, 0xfb, 0x36, 0x19, 0xe4, 0x28, 0xfe, 0xaf, 0x35, 0x3b}
	hash, err := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, obj)
	assert.NotNil(t, hash)
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(results, hash))
}

func TestSecondsToHourMinSec_ShouldWork(t *testing.T) {
	t.Parallel()

	second := 1
	secondsInAMinute := 60
	secondsInAHour := 3600

	testInputOutput := map[int]string{
		0:                                                   "",
		second:                                              "1 second ",
		2 * second:                                          "2 seconds ",
		1 * secondsInAMinute:                                "1 minute ",
		1*secondsInAMinute + second:                         "1 minute 1 second ",
		1*secondsInAMinute + 2*second:                       "1 minute 2 seconds ",
		2*secondsInAMinute + second:                         "2 minutes 1 second ",
		2*secondsInAMinute + 10*second:                      "2 minutes 10 seconds ",
		59*secondsInAMinute + 59*second:                     "59 minutes 59 seconds ",
		secondsInAHour:                                      "1 hour ",
		secondsInAHour + second:                             "1 hour 1 second ",
		secondsInAHour + 2*second:                           "1 hour 2 seconds ",
		secondsInAHour + secondsInAMinute:                   "1 hour 1 minute ",
		secondsInAHour + 2*secondsInAMinute:                 "1 hour 2 minutes ",
		secondsInAHour + secondsInAMinute + second:          "1 hour 1 minute 1 second ",
		secondsInAHour + 2*secondsInAMinute + second:        "1 hour 2 minutes 1 second ",
		secondsInAHour + 2*secondsInAMinute + 10*second:     "1 hour 2 minutes 10 seconds ",
		2*secondsInAHour + 2*secondsInAMinute + 10*second:   "2 hours 2 minutes 10 seconds ",
		60*secondsInAHour + 15*secondsInAMinute + 20*second: "60 hours 15 minutes 20 seconds ",
	}

	for input, expectedOutput := range testInputOutput {
		result := core.SecondsToHourMinSec(input)
		assert.Equal(t, expectedOutput, result)
	}
}
