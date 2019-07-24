package containers_test

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewIntermediateTransactionHandlersContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	assert.NotNil(t, c)
}

//------- Add

func TestIntermediateTransactionHandlersContainer_AddAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	_ = c.Add(block.TxBlock, &mock.IntermediateTransactionHandlerMock{})
	err := c.Add(block.TxBlock, &mock.IntermediateTransactionHandlerMock{})

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestIntermediateTransactionHandlersContainer_AddNilShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	err := c.Add(block.TxBlock, nil)

	assert.Equal(t, process.ErrNilContainerElement, err)
}

func TestIntermediateTransactionHandlersContainer_AddShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	err := c.Add(block.TxBlock, &mock.IntermediateTransactionHandlerMock{})

	assert.Nil(t, err)
	assert.Equal(t, 1, c.Len())
}

//------- AddMultiple

func TestIntermediateTransactionHandlersContainer_AddMultipleAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	keys := []block.Type{block.TxBlock, block.TxBlock}
	preprocessors := []process.IntermediateTransactionHandler{&mock.IntermediateTransactionHandlerMock{}, &mock.IntermediateTransactionHandlerMock{}}

	err := c.AddMultiple(keys, preprocessors)

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestIntermediateTransactionHandlersContainer_AddMultipleLenMismatchShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	keys := []block.Type{block.TxBlock}
	preprocessors := []process.IntermediateTransactionHandler{&mock.IntermediateTransactionHandlerMock{}, &mock.IntermediateTransactionHandlerMock{}}

	err := c.AddMultiple(keys, preprocessors)

	assert.Equal(t, process.ErrLenMismatch, err)
}

func TestIntermediateTransactionHandlersContainer_AddMultipleShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	keys := []block.Type{block.TxBlock, block.SmartContractResultBlock}
	preprocessors := []process.IntermediateTransactionHandler{&mock.IntermediateTransactionHandlerMock{}, &mock.IntermediateTransactionHandlerMock{}}

	err := c.AddMultiple(keys, preprocessors)

	assert.Nil(t, err)
	assert.Equal(t, 2, c.Len())
}

//------- Get

func TestIntermediateTransactionHandlersContainer_GetNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	key := block.TxBlock
	keyNotFound := block.SmartContractResultBlock
	val := &mock.IntermediateTransactionHandlerMock{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(keyNotFound)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

func TestIntermediateTransactionHandlersContainer_GetWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	key := block.TxBlock

	_ = c.Insert(key, "string value")
	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrWrongTypeInContainer, err)
}

func TestIntermediateTransactionHandlersContainer_GetShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	key := block.TxBlock
	val := &mock.IntermediateTransactionHandlerMock{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(key)

	assert.True(t, val == valRecovered)
	assert.Nil(t, err)
}

//------- Replace

func TestIntermediateTransactionHandlersContainer_ReplaceNilValueShouldErrAndNotModify(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	key := block.TxBlock
	val := &mock.IntermediateTransactionHandlerMock{}

	_ = c.Add(key, val)
	err := c.Replace(key, nil)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, process.ErrNilContainerElement, err)
	assert.Equal(t, val, valRecovered)
}

func TestIntermediateTransactionHandlersContainer_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	key := block.TxBlock
	val := &mock.IntermediateTransactionHandlerMock{}
	val2 := &mock.IntermediateTransactionHandlerMock{}

	_ = c.Add(key, val)
	err := c.Replace(key, val2)

	valRecovered, _ := c.Get(key)

	assert.True(t, val2 == valRecovered)
	assert.Nil(t, err)
}

//------- Remove

func TestIntermediateTransactionHandlersContainer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	key := block.TxBlock
	val := &mock.IntermediateTransactionHandlerMock{}

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

//------- Len

func TestIntermediateTransactionHandlersContainer_LenShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewIntermediateTransactionHandlersContainer()

	_ = c.Add(block.TxBlock, &mock.IntermediateTransactionHandlerMock{})
	assert.Equal(t, 1, c.Len())

	keys := c.Keys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, block.TxBlock, keys[0])

	_ = c.Add(block.PeerBlock, &mock.IntermediateTransactionHandlerMock{})
	assert.Equal(t, 2, c.Len())

	keys = c.Keys()
	assert.Equal(t, 2, len(keys))
	assert.Contains(t, keys, block.TxBlock)
	assert.Contains(t, keys, block.PeerBlock)

	c.Remove(block.PeerBlock)
	assert.Equal(t, 1, c.Len())

	keys = c.Keys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, block.TxBlock, keys[0])
}
