package checkpoint

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckpoint_JSON(t *testing.T) {
	t.Parallel()

	epoch := uint32(42)
	cp := &Checkpoint{
		BlockNonce:    1234567,
		BlockHash:     "aabb",
		BlockRootHash: "ccdd",
		Epoch:         &epoch,
		IteratorState: [][]byte{[]byte("first"), []byte("second")},
	}

	buff, err := json.Marshal(cp)
	require.Nil(t, err)
	assert.Equal(t, `{"blockNonce":1234567,"blockHash":"aabb","blockRootHash":"ccdd","epoch":42,"iteratorState":["Zmlyc3Q=","c2Vjb25k"]}`, string(buff))

	decoded := &Checkpoint{}
	require.Nil(t, json.Unmarshal(buff, decoded))
	assert.Equal(t, cp, decoded)

	buff, err = json.Marshal(&Checkpoint{})
	require.Nil(t, err)
	assert.Equal(t, `{"blockNonce":0,"blockHash":"","blockRootHash":"","iteratorState":null}`, string(buff))
}

func TestEncodeDecode(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		cp := &Checkpoint{BlockNonce: 7, BlockHash: "aa", IteratorState: [][]byte{[]byte("state")}}

		encoded, err := Encode(cp)
		require.Nil(t, err)
		decoded, err := Decode(encoded)
		require.Nil(t, err)
		assert.Equal(t, cp, decoded)
	})
	t.Run("invalid input", func(t *testing.T) {
		t.Parallel()

		for _, encoded := range []string{
			"",
			"%%%",
			"bm90IGpzb24",                      // not json
			"eyJibG9ja05vbmNlIjoiYSJ9",         // {"blockNonce":"a"}
			"eyJpdGVyYXRvclN0YXRlIjpbIiUlIl19", // {"iteratorState":["%%"]}
		} {
			cp, err := Decode(encoded)
			assert.Nil(t, cp)
			assert.ErrorIs(t, err, ErrInvalidCheckpoint)
		}
	})
	t.Run("epoch zero is kept", func(t *testing.T) {
		t.Parallel()

		epoch := uint32(0)
		encoded, err := Encode(&Checkpoint{Epoch: &epoch})
		require.Nil(t, err)
		decoded, err := Decode(encoded)
		require.Nil(t, err)
		require.NotNil(t, decoded.Epoch)
		assert.Equal(t, uint32(0), *decoded.Epoch)
	})
	t.Run("empty object is an empty checkpoint", func(t *testing.T) {
		t.Parallel()

		cp, err := Decode("e30") // {}
		require.Nil(t, err)
		assert.Equal(t, &Checkpoint{}, cp)
	})
}
