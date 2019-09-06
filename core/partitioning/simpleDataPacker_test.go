package partitioning_test

import (
	"crypto/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/stretchr/testify/assert"
)

func TestNewSimpleDataPacker_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	sdp, err := partitioning.NewSimpleDataPacker(nil)

	assert.Nil(t, sdp)
	assert.Equal(t, core.ErrNilMarshalizer, err)
}

func TestNewSimpleDataPacker_ValuesOkShouldWork(t *testing.T) {
	t.Parallel()

	sdp, err := partitioning.NewSimpleDataPacker(&mock.MarshalizerMock{})

	assert.NotNil(t, sdp)
	assert.Nil(t, err)
}

//------- PackDataInChunks

func TestSimpleDataPacker_PackDataInChunksInvalidLimitShouldErr(t *testing.T) {
	t.Parallel()

	sdp, _ := partitioning.NewSimpleDataPacker(&mock.MarshalizerMock{})

	buff, err := sdp.PackDataInChunks(make([][]byte, 0), 0)

	assert.Equal(t, core.ErrInvalidValue, err)
	assert.Nil(t, buff)
}

func TestSimpleDataPacker_PackDataInChunksNilInputDataShouldErr(t *testing.T) {
	t.Parallel()

	sdp, _ := partitioning.NewSimpleDataPacker(&mock.MarshalizerMock{})
	buff, err := sdp.PackDataInChunks(nil, 1)

	assert.Equal(t, core.ErrNilInputData, err)
	assert.Nil(t, buff)
}

func TestSimpleDataPacker_PackDataInChunksEmptyDataShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	sdp, _ := partitioning.NewSimpleDataPacker(&mock.MarshalizerMock{})
	buff, err := sdp.PackDataInChunks(make([][]byte, 0), 1)

	assert.Nil(t, err)
	assert.Empty(t, buff)
}

func TestSimpleDataPacker_PackDataInChunksSmallElementsShouldPackTogether(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	sdp, _ := partitioning.NewSimpleDataPacker(marshalizer)

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elem3 := []byte("element3")

	buffSent, err := sdp.PackDataInChunks([][]byte{elem1, elem2, elem3}, maxPacketSize)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(buffSent))
	assert.Nil(t, checkExpectedElements(buffSent[0], marshalizer, [][]byte{elem1, elem2, elem3}))
}

func TestSimpleSplitter_SendDataInChunksWithALargeElementShouldSplit(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	sdp, _ := partitioning.NewSimpleDataPacker(marshalizer)

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

func TestSimpleSplitter_SendDataInChunksWithOnlyOneLargeElementShouldWork(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	sdp, _ := partitioning.NewSimpleDataPacker(marshalizer)

	elemLarge := make([]byte, maxPacketSize)
	_, _ = rand.Read(elemLarge)

	buffSent, err := sdp.PackDataInChunks([][]byte{elemLarge}, maxPacketSize)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(buffSent))
	assert.Nil(t, checkExpectedElements(buffSent[0], marshalizer, [][]byte{elemLarge}))
}
