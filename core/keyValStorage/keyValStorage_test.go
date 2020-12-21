package keyValStorage_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
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

func TestKeyValStorage_ValueWithoutSuffix(t *testing.T) {
	t.Parallel()

	keyVal := keyValStorage.NewKeyValStorage([]byte("key"), []byte("val"))
	trimmedData, err := keyVal.ValueWithoutSuffix([]byte("val2"))
	assert.Equal(t, core.ErrSuffixNotPresentOrInIncorrectPosition, err)
	assert.Nil(t, trimmedData)

	trimmedData, err = keyVal.ValueWithoutSuffix([]byte("va"))
	assert.Equal(t, core.ErrSuffixNotPresentOrInIncorrectPosition, err)
	assert.Nil(t, trimmedData)

	trimmedData, err = keyVal.ValueWithoutSuffix([]byte("l"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("va"), trimmedData)

	trimmedData, err = keyVal.ValueWithoutSuffix([]byte("al"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("v"), trimmedData)

	trimmedData, err = keyVal.ValueWithoutSuffix([]byte("val"))
	assert.Nil(t, err)
	assert.Equal(t, []byte(""), trimmedData)

	trimmedData, err = keyVal.ValueWithoutSuffix([]byte(""))
	assert.Nil(t, err)
	assert.Equal(t, []byte("val"), trimmedData)

	keyVal = keyValStorage.NewKeyValStorage([]byte(""), []byte(""))
	trimmedData, err = keyVal.ValueWithoutSuffix([]byte(""))
	assert.Nil(t, err)
	assert.Equal(t, []byte(""), trimmedData)
}
