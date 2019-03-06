package containers_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptorsContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	assert.NotNil(t, c)
}

//------- Add

func TestInterceptorsContainer_AddAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	_ = c.Add("key", &mock.InterceptorStub{})
	err := c.Add("key", &mock.InterceptorStub{})

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestInterceptorsContainer_AddNilShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	err := c.Add("key", nil)

	assert.Equal(t, process.ErrNilContainerElement, err)
}

func TestInterceptorsContainer_AddShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	err := c.Add("key", &mock.InterceptorStub{})

	assert.Nil(t, err)
}

//------- Get

func TestInterceptorsContainer_GetNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	key := "key"
	keyNotFound := "key not found"
	val := &mock.InterceptorStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(keyNotFound)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

func TestInterceptorsContainer_GetShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	key := "key"
	val := &mock.InterceptorStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(key)

	assert.True(t, val == valRecovered)
	assert.Nil(t, err)
}

//------- Replace

func TestInterceptorsContainer_ReplaceNilValueShouldErrAndNotModify(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	key := "key"
	val := &mock.InterceptorStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, nil)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, process.ErrNilContainerElement, err)
	assert.Equal(t, val, valRecovered)
}

func TestInterceptorsContainer_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	key := "key"
	val := &mock.InterceptorStub{}
	val2 := &mock.InterceptorStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, val2)

	valRecovered, _ := c.Get(key)

	assert.True(t, val2 == valRecovered)
	assert.Nil(t, err)
}

//------- Remove

func TestInterceptorsContainer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	key := "key"
	val := &mock.InterceptorStub{}

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

//------- Len

func TestInterceptorsContainer_LenShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewInterceptorsContainer()

	_ = c.Add("key1", &mock.InterceptorStub{})
	assert.Equal(t, 1, c.Len())

	_ = c.Add("key2", &mock.InterceptorStub{})
	assert.Equal(t, 2, c.Len())

	c.Remove("key1")
	assert.Equal(t, 1, c.Len())
}
