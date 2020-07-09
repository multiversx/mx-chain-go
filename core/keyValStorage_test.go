package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyValStorage(t *testing.T) {
	t.Parallel()

	key := []byte("key")
	val := []byte("val")
	kvs := &KeyValStorage{
		KeyField: key,
		ValField: val,
	}

	assert.Equal(t, key, kvs.KeyField)
	assert.Equal(t, val, kvs.ValField)
}
