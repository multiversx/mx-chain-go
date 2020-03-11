package core_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/stretchr/testify/assert"
)

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

	obj := &batch.Batch{Data: [][]byte{[]byte("object")}}
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

func TestGetShardIdString(t *testing.T) {
	t.Parallel()

	shardIdMeta := uint32(math.MaxUint32)
	assert.Equal(t, "metachain", core.GetShardIdString(shardIdMeta))

	shardId37 := uint32(37)
	assert.Equal(t, "37", core.GetShardIdString(shardId37))
}

func TestEpochStartIdentifier(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	res := core.EpochStartIdentifier(epoch)
	assert.True(t, strings.Contains(res, fmt.Sprintf("%d", epoch)))
}

func TestIsUnknownEpochIdentifier_InvalidIdentifierShouldReturnTrue(t *testing.T) {
	t.Parallel()

	identifier := "epoch5"
	res, err := core.IsUnknownEpochIdentifier([]byte(identifier))
	assert.False(t, res)
	assert.Equal(t, core.ErrInvalidIdentifierForEpochStartBlockRequest, err)
}

func TestIsUnknownEpochIdentifier_IdentifierNotNumericShouldReturnFalse(t *testing.T) {
	t.Parallel()

	identifier := "epochStartBlock_xx"
	res, err := core.IsUnknownEpochIdentifier([]byte(identifier))
	assert.False(t, res)
	assert.Equal(t, core.ErrInvalidIdentifierForEpochStartBlockRequest, err)
}

func TestIsUnknownEpochIdentifier_OkIdentifierShouldReturnFalse(t *testing.T) {
	t.Parallel()

	identifier := "epochStartBlock_5"
	res, err := core.IsUnknownEpochIdentifier([]byte(identifier))
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestCalculateHash_Good(t *testing.T) {
	t.Parallel()

	obj := &batch.Batch{Data: [][]byte{[]byte("object")}}
	results := []byte{0x90, 0xe2, 0x17, 0x2c, 0xaa, 0xa5, 0x4c, 0xb2, 0xad, 0x55, 0xd4, 0xd1, 0x26, 0x91, 0x87, 0xa4, 0xe6, 0x6e, 0xcf, 0x12, 0xfd, 0xc2, 0x5b, 0xf8, 0x67, 0xb7, 0x7, 0x9, 0x6d, 0xe5, 0x43, 0xdd}
	hash, err := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, obj)
	assert.NotNil(t, hash)
	assert.Nil(t, err)
	assert.Equal(t, results, hash)
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
