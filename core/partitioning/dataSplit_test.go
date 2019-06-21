package partitioning_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/stretchr/testify/assert"
)

func checkExpectedElementsNoUnmarshal(buffer [][]byte, expectedElements [][]byte) error {
	if len(buffer) != len(expectedElements) {
		return errors.New(fmt.Sprintf("expected %d elements, got %d", len(expectedElements), len(buffer)))
	}

	for idx, expElem := range expectedElements {
		elem := buffer[idx]
		if !bytes.Equal(elem, expElem) {
			return errors.New(fmt.Sprintf("error at index %d expected %v, got %v", idx, expElem, elem))
		}
	}

	return nil
}

//------- PackDataInChunks

func TestNewNumDataPacker_PackDataInChunksInvalidLimitShouldErr(t *testing.T) {
	t.Parallel()

	ds := &partitioning.DataSplit{}
	buff, err := ds.SplitDataInChunks(make([][]byte, 0), 0)

	assert.Equal(t, core.ErrInvalidValue, err)
	assert.Nil(t, buff)
}

func TestNewNumDataPacker_PackDataInChunksNilInputDataShouldErr(t *testing.T) {
	t.Parallel()

	ds := &partitioning.DataSplit{}
	buff, err := ds.SplitDataInChunks(nil, 1)

	assert.Equal(t, core.ErrNilInputData, err)
	assert.Nil(t, buff)
}

func TestNumDataPacker_PackDataInChunksEmptyDataShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	ds := &partitioning.DataSplit{}
	buff, err := ds.SplitDataInChunks(make([][]byte, 0), 1)

	assert.Nil(t, err)
	assert.Empty(t, buff)
}

func TestNumDataPacker_PackDataInChunksSmallElementsCountShouldPackTogheter(t *testing.T) {
	t.Parallel()

	numPackets := 1000
	ds := &partitioning.DataSplit{}

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elem3 := []byte("element3")

	buffSent, err := ds.SplitDataInChunks([][]byte{elem1, elem2, elem3}, numPackets)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(buffSent))
	assert.Nil(t, checkExpectedElementsNoUnmarshal(buffSent[0], [][]byte{elem1, elem2, elem3}))
}

func TestNumDataPacker_PackDataInChunksEqualElementsShouldPackTogheter(t *testing.T) {
	t.Parallel()

	numPackets := 3
	ds := &partitioning.DataSplit{}

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elem3 := []byte("element3")

	buffSent, err := ds.SplitDataInChunks([][]byte{elem1, elem2, elem3}, numPackets)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(buffSent))
	assert.Nil(t, checkExpectedElementsNoUnmarshal(buffSent[0], [][]byte{elem1, elem2, elem3}))
}

func TestNumDataPacker_PackDataInChunksLargeElementCountShouldPackTogheter(t *testing.T) {
	t.Parallel()

	numPackets := 2
	ds := &partitioning.DataSplit{}

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elem3 := []byte("element3")

	buffSent, err := ds.SplitDataInChunks([][]byte{elem1, elem2, elem3}, numPackets)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(buffSent))
	assert.Nil(t, checkExpectedElementsNoUnmarshal(buffSent[0], [][]byte{elem1, elem2}))
	assert.Nil(t, checkExpectedElementsNoUnmarshal(buffSent[1], [][]byte{elem3}))
}

func TestNumDataPacker_PackDataInChunksSingleElementsShouldPackTogheter(t *testing.T) {
	t.Parallel()

	numPackets := 1
	ds := &partitioning.DataSplit{}

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elem3 := []byte("element3")

	buffSent, err := ds.SplitDataInChunks([][]byte{elem1, elem2, elem3}, numPackets)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(buffSent))
	assert.Nil(t, checkExpectedElementsNoUnmarshal(buffSent[0], [][]byte{elem1}))
	assert.Nil(t, checkExpectedElementsNoUnmarshal(buffSent[1], [][]byte{elem2}))
	assert.Nil(t, checkExpectedElementsNoUnmarshal(buffSent[2], [][]byte{elem3}))
}
