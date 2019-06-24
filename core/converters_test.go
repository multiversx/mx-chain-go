package core_test

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go/core/mock"
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

func TestCalculateHash_NilMarshalizer(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	hash, err := core.CalculateHash(nil, &mock.HasherMock{}, obj)
	assert.Nil(t, hash)
	assert.Equal(t, core.ErrNilMarshalizer, err)
}

func TestCalculateHash_NilHasher(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	hash, err := core.CalculateHash(&mock.MarshalizerMock{}, nil, obj)
	assert.Nil(t, hash)
	assert.Equal(t, core.ErrNilHasher, err)
}

func TestCalculateHash_ErrMarshalizer(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	marshalizer := &mock.MarshalizerMock{
		Fail: true,
	}
	hash, err := core.CalculateHash(marshalizer, &mock.HasherMock{}, obj)
	assert.Nil(t, hash)
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestCalculateHash_NilObject(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hash, err := core.CalculateHash(marshalizer, &mock.HasherMock{}, nil)
	assert.Nil(t, hash)
	assert.Equal(t, mock.ErrNilObjectToMarshal, err)
}

func TestCalculateHash_Good(t *testing.T) {
	t.Parallel()

	obj := []byte("object")
	hash, err := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, obj)
	assert.NotNil(t, hash)
	assert.Nil(t, err)
}
