package containers_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewResolversContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	assert.NotNil(t, c)
}

//------- Add

func TestResolversContainer_AddAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	_ = c.Add("key", &mock.ResolverStub{})
	err := c.Add("key", &mock.ResolverStub{})

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestResolversContainer_AddNilShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	err := c.Add("key", nil)

	assert.Equal(t, process.ErrNilContainerElement, err)
}

func TestResolversContainer_AddEmptyKeyShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	err := c.Add("", &mock.ResolverStub{})

	assert.Nil(t, err)
}

func TestResolversContainer_AddShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	err := c.Add("key", &mock.ResolverStub{})

	assert.Nil(t, err)
}

//------- Get

func TestResolversContainer_GetNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	keyNotFound := "key not found"
	val := &mock.ResolverStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(keyNotFound)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

func TestResolversContainer_GetWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"

	_ = c.Insert(key, "string value")
	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrWrongTypeInContainer, err)
}

func TestResolversContainer_GetShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	val := &mock.ResolverStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(key)

	assert.True(t, val == valRecovered)
	assert.Nil(t, err)
}

//------- Replace

func TestResolversContainer_ReplaceNilValueShouldErrAndNotModify(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	val := &mock.ResolverStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, nil)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, process.ErrNilContainerElement, err)
	assert.Equal(t, val, valRecovered)
}

func TestResolversContainer_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	val := &mock.ResolverStub{}
	val2 := &mock.ResolverStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, val2)

	valRecovered, _ := c.Get(key)

	assert.True(t, val2 == valRecovered)
	assert.Nil(t, err)
}

//------- Remove

func TestResolversContainer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	val := &mock.ResolverStub{}

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

//------- Len

func TestResolversContainer_LenShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	_ = c.Add("key1", &mock.ResolverStub{})
	assert.Equal(t, 1, c.Len())

	_ = c.Add("key2", &mock.ResolverStub{})
	assert.Equal(t, 2, c.Len())

	c.Remove("key1")
	assert.Equal(t, 1, c.Len())
}
