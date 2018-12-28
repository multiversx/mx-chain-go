package uint64ByteSlice_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/stretchr/testify/assert"
)

func TestBigEndianConverter_ToUint64NilBuffShouldErr(t *testing.T) {
	t.Parallel()

	bec := uint64ByteSlice.NewBigEndianConverter()
	val, err := bec.ToUint64(nil)

	assert.Equal(t, uint64(0), val)
	assert.Equal(t, typeConverters.ErrNilByteSlice, err)
}

func TestBigEndianConverter_ToUint64WrongBuffSizeShouldErr(t *testing.T) {
	t.Parallel()

	bec := uint64ByteSlice.NewBigEndianConverter()

	for i := 0; i < 50; i++ {
		if i == 8 {
			//this should actually work, continuing
			continue
		}

		buff := make([]byte, i)
		val, err := bec.ToUint64(buff)
		assert.Equal(t, uint64(0), val)
		assert.Equal(t, typeConverters.ErrByteSliceLenShouldHaveBeen8, err)
	}
}

func TestBigEndianConverter_ToUint64ValsOkShouldWork(t *testing.T) {
	t.Parallel()

	buff := make([]byte, 8)

	bec := uint64ByteSlice.NewBigEndianConverter()
	val, err := bec.ToUint64(buff)

	assert.Equal(t, uint64(0), val)
	assert.Nil(t, err)
}

func TestBigEndianConverter_ToUint64ToByteSlice(t *testing.T) {
	t.Parallel()

	bec := uint64ByteSlice.NewBigEndianConverter()

	buff := bec.ToByteSlice(0)
	assert.Equal(t, make([]byte, 8), buff)
}

func TestBigEndianConverter_ToByteSliceOkValsShouldWork(t *testing.T) {
	t.Parallel()

	vals := make(map[uint64][]byte)

	vals[uint64(0)] = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	vals[uint64(1)] = []byte{0, 0, 0, 0, 0, 0, 0, 1}
	vals[uint64(255)] = []byte{0, 0, 0, 0, 0, 0, 0, 255}
	vals[uint64(256)] = []byte{0, 0, 0, 0, 0, 0, 1, 0}
	vals[uint64(65536)] = []byte{0, 0, 0, 0, 0, 1, 0, 0}
	vals[uint64(1<<64-1)] = []byte{255, 255, 255, 255, 255, 255, 255, 255}

	bec := uint64ByteSlice.NewBigEndianConverter()

	for k, v := range vals {
		buff := bec.ToByteSlice(k)

		assert.Equal(t, v, buff)
	}
}

func TestBigEndianConverter_Converting100kRandomValuesShouldWork(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	for i := 0; i < 100000; i++ {
		val := r.Uint64()

		bec := uint64ByteSlice.NewBigEndianConverter()

		buff := bec.ToByteSlice(val)
		valConverted, err := bec.ToUint64(buff)

		assert.Equal(t, val, valConverted)
		assert.Nil(t, err)
	}
}

func BenchmarkNonceToHashCacher_Uint64ToByteArrayConversion(b *testing.B) {
	bec := uint64ByteSlice.NewBigEndianConverter()

	for i := 0; i < b.N; i++ {
		_ = bec.ToByteSlice(uint64(i))
	}
}

func BenchmarkNonceToHashCacher_Uint64ToByteArrayConversionAndBackToUint64(b *testing.B) {
	bec := uint64ByteSlice.NewBigEndianConverter()

	for i := 0; i < b.N; i++ {
		buff := bec.ToByteSlice(uint64(i))
		val, _ := bec.ToUint64(buff)

		if uint64(i) != val {
			assert.Fail(b, fmt.Sprintf("Not equal %v, got %v\n", i, val))
		}
	}
}
