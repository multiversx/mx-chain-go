package containers_test

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewPreProcessorsContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	assert.NotNil(t, c)
}

//------- Add

func TestPreProcessorsContainer_AddAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	_ = c.Add(block.TxBlock, &mock.PreProcessorMock{})
	err := c.Add(block.TxBlock, &mock.PreProcessorMock{})

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestPreProcessorsContainer_AddNilShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	err := c.Add(block.TxBlock, nil)

	assert.Equal(t, process.ErrNilContainerElement, err)
}

func TestPreProcessorsContainer_AddShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	err := c.Add(block.TxBlock, &mock.PreProcessorMock{})

	assert.Nil(t, err)
	assert.Equal(t, 1, c.Len())
}

//------- AddMultiple

func TestPreProcessorsContainer_AddMultipleAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	keys := []block.Type{block.TxBlock, block.TxBlock}
	preprocessors := []process.PreProcessor{&mock.PreProcessorMock{}, &mock.PreProcessorMock{}}

	err := c.AddMultiple(keys, preprocessors)

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestPreProcessorsContainer_AddMultipleLenMismatchShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	keys := []block.Type{block.TxBlock}
	preprocessors := []process.PreProcessor{&mock.PreProcessorMock{}, &mock.PreProcessorMock{}}

	err := c.AddMultiple(keys, preprocessors)

	assert.Equal(t, process.ErrLenMismatch, err)
}

func TestPreProcessorsContainer_AddMultipleShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	keys := []block.Type{block.TxBlock, block.SmartContractResultBlock}
	preprocessors := []process.PreProcessor{&mock.PreProcessorMock{}, &mock.PreProcessorMock{}}

	err := c.AddMultiple(keys, preprocessors)

	assert.Nil(t, err)
	assert.Equal(t, 2, c.Len())
}

//------- Get

func TestPreProcessorsContainer_GetNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	key := block.TxBlock
	keyNotFound := block.SmartContractResultBlock
	val := &mock.PreProcessorMock{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(keyNotFound)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

func TestPreProcessorsContainer_GetWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	key := block.TxBlock

	_ = c.Insert(key, "string value")
	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrWrongTypeInContainer, err)
}

func TestPreProcessorsContainer_GetShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	key := block.TxBlock
	val := &mock.PreProcessorMock{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(key)

	assert.True(t, val == valRecovered)
	assert.Nil(t, err)
}

//------- Replace

func TestPreProcessorsContainer_ReplaceNilValueShouldErrAndNotModify(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	key := block.TxBlock
	val := &mock.PreProcessorMock{}

	_ = c.Add(key, val)
	err := c.Replace(key, nil)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, process.ErrNilContainerElement, err)
	assert.Equal(t, val, valRecovered)
}

func TestPreProcessorsContainer_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	key := block.TxBlock
	val := &mock.PreProcessorMock{}
	val2 := &mock.PreProcessorMock{}

	_ = c.Add(key, val)
	err := c.Replace(key, val2)

	valRecovered, _ := c.Get(key)

	assert.True(t, val2 == valRecovered)
	assert.Nil(t, err)
}

//------- Remove

func TestPreProcessorsContainer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	key := block.TxBlock
	val := &mock.PreProcessorMock{}

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrInvalidContainerKey, err)
}

//------- Len

func TestPreProcessorsContainer_LenShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewPreProcessorsContainer()

	_ = c.Add(block.TxBlock, &mock.PreProcessorMock{})
	assert.Equal(t, 1, c.Len())

	_ = c.Add(block.PeerBlock, &mock.PreProcessorMock{})
	assert.Equal(t, 2, c.Len())

	c.Remove(block.PeerBlock)
	assert.Equal(t, 1, c.Len())
}
