package core_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestToB64ShouldReturnNil(t *testing.T) {
	val := core.ToB64(nil)
	assert.Equal(t, "<NIL>", val)
}

func TestToB64ShouldWork(t *testing.T) {
	buff := []byte("test")
	val := core.ToB64(buff)
	assert.Equal(t, base64.StdEncoding.EncodeToString(buff), val)
}

func TestToHexShouldReturnNil(t *testing.T) {
	val := core.ToHex(nil)
	assert.Equal(t, "<NIL>", val)
}

func TestToHexShouldWork(t *testing.T) {
	buff := []byte("test")
	val := core.ToHex(buff)
	assert.Equal(t, "0x"+hex.EncodeToString(buff), val)
}
