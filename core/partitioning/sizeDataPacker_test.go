package partitioning_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func checkExpectedElements(buffer []byte, marshalizer marshal.Marshalizer, expectedElements [][]byte) error {
	b := &batch.Batch{}
	err := marshalizer.Unmarshal(b, buffer)
	if err != nil {
		return err
	}

	if len(b.Data) != len(expectedElements) {
		return errors.New(fmt.Sprintf("expected %d elements, got %d", len(expectedElements), len(b.Data)))
	}

	for idx, expElem := range expectedElements {
		elem := b.Data[idx]
		if !bytes.Equal(elem, expElem) {
			return errors.New(fmt.Sprintf("error at index %d expected %v, got %v", idx, expElem, elem))
		}
	}

	return nil
}

func TestNewSizeDataPacker_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	sdp, err := partitioning.NewSizeDataPacker(nil)

	assert.Nil(t, sdp)
	assert.Equal(t, core.ErrNilMarshalizer, err)
}

func TestNewSizeDataPacker_ValuesOkShouldWork(t *testing.T) {
	t.Parallel()

	sdp, err := partitioning.NewSizeDataPacker(&mock.MarshalizerMock{})

	assert.NotNil(t, sdp)
	assert.Nil(t, err)
}

//------- PackDataInChunks

func TestSizeDataPacker_PackDataInChunksInvalidLimitShouldErr(t *testing.T) {
	t.Parallel()

	sdp, _ := partitioning.NewSizeDataPacker(&mock.MarshalizerMock{})

	buff, err := sdp.PackDataInChunks(make([][]byte, 0), 0)

	assert.Equal(t, core.ErrInvalidValue, err)
	assert.Nil(t, buff)
}

func TestSizeDataPacker_PackDataInChunksNilInputDataShouldErr(t *testing.T) {
	t.Parallel()

	sdp, _ := partitioning.NewSizeDataPacker(&mock.MarshalizerMock{})
	buff, err := sdp.PackDataInChunks(nil, 1)

	assert.Equal(t, core.ErrNilInputData, err)
	assert.Nil(t, buff)
}

func TestSizeDataPacker_PackDataInChunksEmptyDataShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	sdp, _ := partitioning.NewSizeDataPacker(&mock.MarshalizerMock{})
	buff, err := sdp.PackDataInChunks(make([][]byte, 0), 1)

	assert.Nil(t, err)
	assert.Empty(t, buff)
}

func TestSizeDataPacker_PackDataInChunksSmallElementsShouldPackTogether(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	sdp, _ := partitioning.NewSizeDataPacker(marshalizer)

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elem3 := []byte("element3")

	buffSent, err := sdp.PackDataInChunks([][]byte{elem1, elem2, elem3}, maxPacketSize)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(buffSent))
	assert.Nil(t, checkExpectedElements(buffSent[0], marshalizer, [][]byte{elem1, elem2, elem3}))
}

func TestSliceSplitter_SendDataInChunksWithALargeElementShouldSplit(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	sdp, _ := partitioning.NewSizeDataPacker(marshalizer)

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elemLarge := make([]byte, maxPacketSize)
	_, _ = rand.Read(elemLarge)
	elem3 := []byte("element3")

	buffSent, err := sdp.PackDataInChunks([][]byte{elem1, elem2, elemLarge, elem3}, maxPacketSize)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(buffSent))
	assert.Nil(t, checkExpectedElements(buffSent[0], marshalizer, [][]byte{elem1, elem2}))
	assert.Nil(t, checkExpectedElements(buffSent[1], marshalizer, [][]byte{elemLarge}))
	assert.Nil(t, checkExpectedElements(buffSent[2], marshalizer, [][]byte{elem3}))
}

func TestSliceSplitter_SendDataInChunksWithOnlyOneLargeElementShouldWork(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	sdp, _ := partitioning.NewSizeDataPacker(marshalizer)

	elemLarge := make([]byte, maxPacketSize)
	_, _ = rand.Read(elemLarge)

	buffSent, err := sdp.PackDataInChunks([][]byte{elemLarge}, maxPacketSize)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(buffSent))
	assert.Nil(t, checkExpectedElements(buffSent[0], marshalizer, [][]byte{elemLarge}))
}
