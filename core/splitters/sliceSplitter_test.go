package splitters_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/splitters"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func checkExpectedElements(buffer []byte, marshalizer marshal.Marshalizer, expectedElements [][]byte) error {
	elements := make([][]byte, 0)
	err := marshalizer.Unmarshal(&elements, buffer)
	if err != nil {
		return err
	}

	if len(elements) != len(expectedElements) {
		return errors.New(fmt.Sprintf("expected %d elements, got %d", len(expectedElements), len(elements)))
	}

	for idx, expElem := range expectedElements {
		elem := elements[idx]

		if !bytes.Equal(elem, expElem) {
			return errors.New(fmt.Sprintf("error at index %d expected %v, got %v", idx, expElem, elem))
		}
	}

	return nil
}

func TestNewSliceSplitter_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	ss, err := splitters.NewSliceSplitter(
		nil,
	)

	assert.Nil(t, ss)
	assert.Equal(t, core.ErrNilMarshalizer, err)
}

func TestNewSliceSplitter_ValuesOkShouldWork(t *testing.T) {
	t.Parallel()

	ss, err := splitters.NewSliceSplitter(
		&mock.MarshalizerMock{},
	)

	assert.NotNil(t, ss)
	assert.Nil(t, err)
}

//------- SendDataInChunks

func TestSliceSplitter_SendDataInChunksNilSendHandlerShouldErr(t *testing.T) {
	t.Parallel()

	ss, _ := splitters.NewSliceSplitter(
		&mock.MarshalizerMock{},
	)

	err := ss.SendDataInChunks(nil, nil, 10)

	assert.Equal(t, core.ErrNilSendHandler, err)
}

func TestSliceSplitter_SendDataInChunksInvalidMaxPacketSizeShouldErr(t *testing.T) {
	t.Parallel()

	ss, _ := splitters.NewSliceSplitter(
		&mock.MarshalizerMock{},
	)

	err := ss.SendDataInChunks(nil,
		func(buff []byte) error {
			return nil
		},
		0,
	)

	assert.Equal(t, core.ErrInvalidValue, err)
}

func TestSliceSplitter_SendDataInChunksNilDataShouldNotCallSend(t *testing.T) {
	t.Parallel()

	sendWasCalled := false
	ss, _ := splitters.NewSliceSplitter(
		&mock.MarshalizerMock{},
	)

	err := ss.SendDataInChunks(
		nil,
		func(buff []byte) error {
			sendWasCalled = true
			return nil
		},
		1,
	)

	assert.Nil(t, err)
	assert.False(t, sendWasCalled)
}

func TestSliceSplitter_SendDataInChunksSmallElementsShouldPackTogheter(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	var buffSent []byte
	ss, _ := splitters.NewSliceSplitter(
		marshalizer,
	)

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elem3 := []byte("element3")

	err := ss.SendDataInChunks(
		[][]byte{elem1, elem2, elem3},
		func(buff []byte) error {
			buffSent = buff
			return nil
		},
		maxPacketSize,
	)
	assert.Nil(t, err)

	assert.Nil(t, checkExpectedElements(buffSent, marshalizer, [][]byte{elem1, elem2, elem3}))
}

func TestSliceSplitter_SendDataInChunksSmallElementsSendErrorsShoudErr(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	ss, _ := splitters.NewSliceSplitter(
		marshalizer,
	)

	expectedErr := errors.New("expected err")
	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elem3 := []byte("element3")

	err := ss.SendDataInChunks(
		[][]byte{elem1, elem2, elem3},
		func(buff []byte) error {
			return expectedErr
		},
		maxPacketSize,
	)
	assert.Equal(t, expectedErr, err)
}

func TestSliceSplitter_SendDataInChunksWithALargeElementShouldSplit(t *testing.T) {
	t.Parallel()

	maxPacketSize := 1000
	marshalizer := &mock.MarshalizerMock{}
	var buffSent [][]byte
	ss, _ := splitters.NewSliceSplitter(
		marshalizer,
	)

	elem1 := []byte("element1")
	elem2 := []byte("element2")
	elemLarge := make([]byte, maxPacketSize)
	rand.Read(elemLarge)
	elem3 := []byte("element3")

	err := ss.SendDataInChunks(
		[][]byte{elem1, elem2, elemLarge, elem3},
		func(buff []byte) error {
			buffSent = append(buffSent, buff)
			return nil
		},
		maxPacketSize,
	)
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
	var buffSent [][]byte
	ss, _ := splitters.NewSliceSplitter(
		marshalizer,
	)

	elemLarge := make([]byte, maxPacketSize)
	rand.Read(elemLarge)

	err := ss.SendDataInChunks(
		[][]byte{elemLarge},
		func(buff []byte) error {
			buffSent = append(buffSent, buff)
			return nil
		},
		maxPacketSize,
	)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(buffSent))
	assert.Nil(t, checkExpectedElements(buffSent[0], marshalizer, [][]byte{elemLarge}))
}
