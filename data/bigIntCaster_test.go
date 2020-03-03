package data

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	name string
	val  int64
}

func getTestData() []testData {
	return []testData{
		{
			name: "zero",
			val:  int64(0),
		},

		{
			name: "one",
			val:  int64(1),
		},

		{
			name: "minus one",
			val:  int64(-1),
		},

		{
			name: "min",
			val:  math.MinInt64,
		},

		{
			name: "max",
			val:  math.MaxInt64,
		},
	}
}

func TestBigIntCaster_MarshalBytes(t *testing.T) {
	for _, td := range getTestData() {
		t.Run(td.name, func(t *testing.T) {
			v0 := big.NewInt(td.val)
			caster := &BigIntCaster{}

			sz := caster.Size(v0)
			assert.Greater(t, sz, 0)

			data := make([]byte, sz)
			dataSz, err := caster.MarshalTo(v0, data)
			assert.Nil(t, err)
			assert.Equal(t, dataSz, sz)

			v1, err := caster.Unmarshal(data)
			assert.Nil(t, err)

			assert.True(t, caster.Equal(v0, v1))
			assert.True(t, caster.Equal(v1, v0))
			assert.Equal(t, td.val, v1.Int64())

		})
	}
}

func TestBigIntCaster_MarshalToBadBuffer(t *testing.T) {
	v0 := big.NewInt(-2020)
	caster := &BigIntCaster{}
	data := make([]byte, caster.Size(v0)-1)
	data_len, err := caster.MarshalTo(v0, data)

	assert.NotNil(t, err)
	assert.Equal(t, 0, data_len)
	assert.Equal(t, int64(-2020), v0.Int64())
}

func TestBigIntCaster_BadUnmarshal(t *testing.T) {
	t.Run("Nil buffer", func(t *testing.T) {
		caster := &BigIntCaster{}
		v0, err := caster.Unmarshal(nil)
		assert.NotNil(t, err)
		assert.Nil(t, nil, v0)
	})

	t.Run("Empty buffer", func(t *testing.T) {
		caster := &BigIntCaster{}
		v0, err := caster.Unmarshal([]byte{})
		assert.NotNil(t, err)
		assert.Nil(t, v0)
	})

	t.Run("Bad sign", func(t *testing.T) {
		caster := &BigIntCaster{}
		v0, err := caster.Unmarshal([]byte{2, 3, 4})
		assert.NotNil(t, err)
		assert.Nil(t, v0)
	})
}
