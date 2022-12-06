package converter

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDirectStringPubkeyConverter(t *testing.T) {
	t.Parallel()

	converter := NewDirectStringPubkeyConverter()
	assert.False(t, check.IfNil(converter))
}

func TestDirectStringPubkeyConverter_Encode(t *testing.T) {
	t.Parallel()

	converter := NewDirectStringPubkeyConverter()
	testData := []byte("test data")

	encoded := converter.Encode(testData)
	assert.Equal(t, string(testData), encoded)
}

func TestDirectStringPubkeyConverter_Decode(t *testing.T) {
	t.Parallel()

	converter := NewDirectStringPubkeyConverter()
	testData := "test data"

	decoded, err := converter.Decode(testData)
	assert.Equal(t, []byte(testData), decoded)
	assert.Nil(t, err)
}
