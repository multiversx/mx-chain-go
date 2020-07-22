package keyValStorage_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/keyValStorage"
	"github.com/stretchr/testify/assert"
)

func TestNewKeyValStorage_GetKeyAndVal(t *testing.T) {
	t.Parallel()

	key := []byte("key")
	value := []byte("value")

	keyVal := keyValStorage.NewKeyValStorage(key, value)
	assert.NotNil(t, keyVal)
	assert.Equal(t, key, keyVal.Key())
	assert.Equal(t, value, keyVal.Value())
}
