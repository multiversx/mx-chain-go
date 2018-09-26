package trie_try01

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnload(t *testing.T) {
	nf := nodeFlag{dirty: true, hash: make([]byte, 0), gen: uint16(0)}

	assert.False(t, nf.canUnload(uint16(0), uint16(0)))

	nf.dirty = false

	assert.True(t, nf.canUnload(uint16(0), uint16(0)))

}
