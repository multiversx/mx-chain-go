package builtInFunctions

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewBuiltInFunctionContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	assert.False(t, check.IfNil(c))
}

//------- Add

func TestBuiltInFunctionContainer_AddAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	_ = c.Add("key", &mock.BuiltInFunctionStub{})
	err := c.Add("key", &mock.BuiltInFunctionStub{})

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestBuiltInFunctionContainer_AddNilShouldErr(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	err := c.Add("key", nil)

	assert.Equal(t, process.ErrNilContainerElement, err)
}

func TestBuiltInFunctionContainer_AddShouldWork(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	err := c.Add("key", &mock.BuiltInFunctionStub{})

	assert.Nil(t, err)
	assert.Equal(t, 1, c.Len())
}

//------- Get

func TestBuiltInFunctionContainer_GetNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	key := "key"
	keyNotFound := "key not found"
	val := &mock.BuiltInFunctionStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(keyNotFound)

	assert.Nil(t, valRecovered)
	assert.True(t, errors.Is(err, process.ErrInvalidContainerKey))
}

func TestBuiltInFunctionContainer_GetShouldWork(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	key := "key"
	val := &mock.BuiltInFunctionStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(key)

	assert.True(t, val == valRecovered)
	assert.Nil(t, err)
}

//------- Replace

func TestBuiltInFunctionContainer_ReplaceNilValueShouldErrAndNotModify(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	key := "key"
	val := &mock.BuiltInFunctionStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, nil)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, process.ErrNilContainerElement, err)
	assert.Equal(t, val, valRecovered)
}

func TestBuiltInFunctionContainer_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	key := "key"
	val := &mock.BuiltInFunctionStub{}
	val2 := &mock.BuiltInFunctionStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, val2)

	valRecovered, _ := c.Get(key)

	assert.True(t, val2 == valRecovered)
	assert.Nil(t, err)
}

//------- Remove

func TestBuiltInFunctionContainer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	key := "key"
	val := &mock.BuiltInFunctionStub{}

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.True(t, errors.Is(err, process.ErrInvalidContainerKey))
}

//------- Len

func TestBuiltInFunctionContainer_LenShouldWork(t *testing.T) {
	t.Parallel()

	c := NewBuiltInFunctionContainer()

	_ = c.Add("key1", &mock.BuiltInFunctionStub{})
	assert.Equal(t, 1, c.Len())

	_ = c.Add("key2", &mock.BuiltInFunctionStub{})
	assert.Equal(t, 2, c.Len())

	c.Remove("key1")
	assert.Equal(t, 1, c.Len())
}
