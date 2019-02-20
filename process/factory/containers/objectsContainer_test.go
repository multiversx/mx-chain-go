package containers_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/stretchr/testify/assert"
)

func TestNewObjectsContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	assert.NotNil(t, c)
}

//------- Add

func TestObjectsContainer_AddAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	_ = c.Add("key", "val")
	err := c.Add("key", "value")

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestObjectsContainer_AddNilShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	err := c.Add("key", nil)

	assert.Equal(t, process.ErrNilContainerElement, err)
}

func TestObjectsContainer_AddShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	err := c.Add("key", "value")

	assert.Nil(t, err)
}

//------- Get

func TestObjectsContainer_GetNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	key := "key"
	keyNotFound := "key not found"
	val := "value"

	_ = c.Add(key, val)
	valRecovered, err := c.Get(keyNotFound)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

func TestObjectsContainer_GetShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	key := "key"
	val := "value"

	_ = c.Add(key, val)
	valRecovered, err := c.Get(key)

	assert.Equal(t, val, valRecovered)
	assert.Nil(t, err)
}

//------- Replace

func TestObjectsContainer_ReplaceNilValueShouldErrAndNotModify(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	key := "key"
	val := "value"

	_ = c.Add(key, val)
	err := c.Replace(key, nil)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, process.ErrNilContainerElement, err)
	assert.Equal(t, val, valRecovered)
}

func TestObjectsContainer_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	key := "key"
	val := "value"
	val2 := "value2"

	_ = c.Add(key, val)
	err := c.Replace(key, val2)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, val2, valRecovered)
	assert.Nil(t, err)
}

//------- Remove

func TestObjectsContainer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	key := "key"
	val := "value"

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

//------- Len

func TestObjectsContainer_LenShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewObjectsContainer()

	_ = c.Add("key1", "val")
	assert.Equal(t, 1, c.Len())

	_ = c.Add("key2", "val")
	assert.Equal(t, 2, c.Len())

	c.Remove("key1")
	assert.Equal(t, 1, c.Len())
}
