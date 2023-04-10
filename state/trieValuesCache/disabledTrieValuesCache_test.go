package trieValuesCache

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledTrieValuesCache(t *testing.T) {
	t.Parallel()

	d := NewDisabledTrieValuesCache()
	assert.False(t, d.IsInterfaceNil())
}

func TestDisabledTrieValuesCache_PutAndGet(t *testing.T) {
	t.Parallel()

	d := NewDisabledTrieValuesCache()

	key := []byte("key")
	val := core.TrieData{
		Key:     key,
		Value:   []byte("value"),
		Version: 0,
	}

	d.Put([]byte("key"), val)
	_, ok := d.Get([]byte("key"))
	assert.False(t, ok)

	d.Clean()
}
