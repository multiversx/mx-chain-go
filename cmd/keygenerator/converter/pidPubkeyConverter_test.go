package converter

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewPidPubkeyConverter(t *testing.T) {
	t.Parallel()

	converter := NewPidPubkeyConverter()
	assert.False(t, check.IfNil(converter))
}

func TestPidPubkeyConverter_Encode(t *testing.T) {
	t.Parallel()

	converter := NewPidPubkeyConverter()
	t.Run("not a valid public key should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
			}
		}()

		encoded, err := converter.Encode([]byte("not a valid PK"))
		assert.Contains(t, err.Error(), "parameter is invalid")
		assert.Empty(t, encoded)
	})
	t.Run("valid public key should work", func(t *testing.T) {
		t.Parallel()

		pkHex := "03ca8ec5bd3b84d05e59d5c9cecd548059106649d9e3465f628498628732ab23c9"
		pkBytes, err := hex.DecodeString(pkHex)
		assert.Nil(t, err)

		encoded, err := converter.Encode(pkBytes)
		assert.Nil(t, err)
		assert.Equal(t, "16Uiu2HAmSHgyTYyawhsZv9opxTHX77vKjoPeGkyCYS5fYVMssHjN", encoded)
	})
}

func TestPidPubkeyConverter_Decode(t *testing.T) {
	t.Parallel()

	converter := NewPidPubkeyConverter()

	decoded, err := converter.Decode("")
	assert.Nil(t, decoded)
	assert.Equal(t, errNotImplemented, err)
}
