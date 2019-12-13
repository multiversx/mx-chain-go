package data

import (
	"math"
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

func TestProtoBigInt_MarshalBytes(t *testing.T) {
	for _, td := range getTestData() {
		t.Run(td.name, func(t *testing.T) {
			v0 := ProtoBigInt{}
			v0.SetInt64(td.val)
			data, err := v0.Marshal()

			assert.Nil(t, err)
			assert.Equal(t, len(data), v0.Size())

			v1 := ProtoBigInt{}
			err = v1.Unmarshal(data)

			assert.Nil(t, err)
			assert.True(t, v1.Equal(v0))
			assert.True(t, v0.Equal(v1))
			assert.Equal(t, td.val, v1.Int64())

		})
	}
}

func TestProtoBigInt_MarshalTo(t *testing.T) {
	t.Run("GoodBuffer", func(t *testing.T) {
		v0 := ProtoBigInt{}
		v0.SetInt64(-1)
		data := make([]byte, 5)
		data_len, err := v0.MarshalTo(data)

		assert.Nil(t, err)
		assert.Equal(t, data_len, v0.Size())

		// trim the extra bytes
		data = data[:data_len]

		v1 := ProtoBigInt{}
		err = v1.Unmarshal(data)

		assert.Nil(t, err)
		assert.True(t, v1.Equal(v0))
		assert.True(t, v0.Equal(v1))
		assert.Equal(t, int64(-1), v1.Int64())
	})

	t.Run("Small", func(t *testing.T) {
		v0 := ProtoBigInt{}
		v0.SetInt64(-2019)
		data := make([]byte, v0.Size()-1)
		data_len, err := v0.MarshalTo(data)

		assert.NotNil(t, err)
		assert.Equal(t, 0, data_len)
		assert.Equal(t, int64(-2019), v0.Int64())
	})
}

func TestProtoBigInt_MarshalJSON(t *testing.T) {
	for _, td := range getTestData() {
		t.Run(td.name, func(t *testing.T) {
			v0 := ProtoBigInt{}
			v0.SetInt64(td.val)
			data, err := v0.MarshalJSON()

			assert.Nil(t, err)

			v1 := ProtoBigInt{}
			t.Log(data)
			err = v1.UnmarshalJSON(data)

			assert.Nil(t, err)
			assert.True(t, v1.Equal(v0))
			assert.True(t, v0.Equal(v1))
			assert.Equal(t, td.val, v1.Int64())

		})
	}
}

func TestProtoBigInt_BadUnmarshal(t *testing.T) {
	t.Run("Nil buffer", func(t *testing.T) {
		v0 := ProtoBigInt{}
		err := v0.Unmarshal(nil)
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), v0.Int64())
	})

	t.Run("Nil buffer JSON", func(t *testing.T) {
		v0 := ProtoBigInt{}
		err := v0.UnmarshalJSON(nil)
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), v0.Int64())
	})

	t.Run("Empty buffer", func(t *testing.T) {
		v0 := ProtoBigInt{}
		err := v0.Unmarshal([]byte{})
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), v0.Int64())
	})

	t.Run("Empty buffer JSON", func(t *testing.T) {
		v0 := ProtoBigInt{}
		err := v0.UnmarshalJSON([]byte{})
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), v0.Int64())
	})

	t.Run("Bad sign", func(t *testing.T) {
		v0 := ProtoBigInt{}
		err := v0.Unmarshal([]byte{2, 3, 4})
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), v0.Int64())
	})

	t.Run("Bad string JSON", func(t *testing.T) {
		v0 := ProtoBigInt{}
		err := v0.UnmarshalJSON([]byte("-123"))
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), v0.Int64())

		err = v0.UnmarshalJSON([]byte("\"a\""))
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), v0.Int64())
	})
}

func TestProtobBigInt_Compare(t *testing.T) {

	v0 := ProtoBigInt{}
	assert.Equal(t, int64(0), v0.Int64())

	v1 := ProtoBigInt{}
	v1.SetInt64(1)
	assert.Equal(t, int64(1), v1.Int64())
	assert.Equal(t, 1, v1.Compare(v0))
	assert.Equal(t, -1, v0.Compare(v1))

	vm1 := ProtoBigInt{}
	vm1.SetInt64(-1)
	assert.Equal(t, int64(-1), vm1.Int64())
	assert.Equal(t, -1, vm1.Compare(v0))
	assert.Equal(t, 1, v0.Compare(vm1))

	assert.Equal(t, -1, vm1.Compare(v1))
	assert.Equal(t, 1, v1.Compare(vm1))
}
