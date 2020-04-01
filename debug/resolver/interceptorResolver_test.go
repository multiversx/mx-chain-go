package resolver

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var topic = "topic"
var hash = []byte("hash")
var numIntra = 10
var numCross = 9

//------- NewInterceptorResolver

func TestNewInterceptorResolver_InvalidSizeShouldErr(t *testing.T) {
	t.Parallel()

	ir, err := NewInterceptorResolver(-1)

	assert.True(t, check.IfNil(ir))
	assert.NotNil(t, err)
}

func TestNewInterceptorResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	ir, err := NewInterceptorResolver(1)

	assert.False(t, check.IfNil(ir))
	assert.Nil(t, err)
}

//------- RequestedData

func TestInterceptorResolver_RequestedDataWithFiveIdentifiersShouldWork(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(1000)
	numIdentifiers := 5
	foundMap := make(map[string]struct{})
	for i := 0; i < numIdentifiers; i++ {
		newTopic := fmt.Sprintf("topic%d", i)
		ir.RequestedData(newTopic, hash, numIntra, numCross)
		foundMap[newTopic] = struct{}{}
	}

	s := strings.Join(ir.Query("*"), "\r\n")
	fmt.Println(s)

	requests := ir.Requests()
	require.Equal(t, numIdentifiers, len(requests))
	for _, req := range requests {
		_, ok := foundMap[req.topic]
		assert.True(t, ok, "topic: "+req.topic)
		delete(foundMap, req.topic)
	}

	assert.Equal(t, 0, len(foundMap))
}

func TestInterceptorResolver_RequestedDataSameIdentifierShouldAddRequested(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(1000)
	ir.RequestedData(topic, hash, numIntra, numCross)
	requests := ir.Requests()
	require.Equal(t, 1, len(requests))
	expected := &request{
		hash:        hash,
		topic:       topic,
		numReqIntra: numIntra,
		numReqCross: numCross,
	}

	assert.Equal(t, expected, requests[0])

	ir.RequestedData(topic, hash, numIntra, numCross)
	requests = ir.Requests()
	require.Equal(t, 1, len(requests))
	expected = &request{
		hash:        hash,
		topic:       topic,
		numReqIntra: numIntra * 2,
		numReqCross: numCross * 2,
	}

	assert.Equal(t, expected, requests[0])
	fmt.Println(ir.Query("*"))
}

//------- ProcessedHash

func TestInterceptorResolver_ProcessedHashNotFoundShouldNotAdd(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(1000)

	ir.ProcessedHash(topic, hash, nil)

	require.Equal(t, 0, len(ir.Requests()))
}

func TestInterceptorResolver_ProcessedHashExistingNoErrorShouldRemove(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(1000)
	ir.RequestedData(topic, hash, numIntra, numCross)
	require.Equal(t, 1, len(ir.Requests()))

	ir.ProcessedHash(topic, hash, nil)

	require.Equal(t, 0, len(ir.Requests()))
}

func TestInterceptorResolver_ProcessedHashExistingWithErrorShouldIncrementProcessed(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(1000)
	ir.RequestedData(topic, hash, numIntra, numCross)
	require.Equal(t, 1, len(ir.Requests()))

	err := errors.New("expected err")
	ir.ProcessedHash(topic, hash, err)

	requests := ir.Requests()
	require.Equal(t, 1, len(requests))

	expected := &request{
		hash:         hash,
		topic:        topic,
		numReqIntra:  numIntra,
		numReqCross:  numCross,
		lastErr:      err,
		numProcessed: 1,
		numReceived:  0,
	}

	assert.Equal(t, expected, requests[0])
	fmt.Println(ir.Query("*"))
}

//------- ReceivedHash

func TestInterceptorResolver_ReceivedHashNotFoundShouldNotAdd(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(1000)

	ir.ReceivedHash(topic, hash)

	require.Equal(t, 0, len(ir.Requests()))
}

func TestInterceptorResolver_ReceiveddHashExistingShouldIncrementReceived(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(1000)
	ir.RequestedData(topic, hash, numIntra, numCross)
	require.Equal(t, 1, len(ir.Requests()))

	ir.ReceivedHash(topic, hash)

	requests := ir.Requests()
	require.Equal(t, 1, len(requests))

	expected := &request{
		hash:         hash,
		topic:        topic,
		numReqIntra:  numIntra,
		numReqCross:  numCross,
		lastErr:      nil,
		numProcessed: 0,
		numReceived:  1,
	}

	assert.Equal(t, expected, requests[0])
}

//------- functions

func TestInterceptorResolver_Query(t *testing.T) {
	t.Parallel()

	topic1 := "topic1"
	topic2 := "aaaa"
	ir, _ := NewInterceptorResolver(1000)
	ir.RequestedData(topic1, hash, numIntra, numCross)
	ir.RequestedData(topic2, hash, numIntra, numCross)

	assert.Equal(t, 0, len(ir.Query("not a topic")))
	assert.Equal(t, 1, len(ir.Query(topic1)))
	assert.Equal(t, 1, len(ir.Query(topic2)))
	assert.Equal(t, 2, len(ir.Query("*")))
}

func TestInterceptorResolver_Enabled(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(1)

	assert.True(t, ir.Enabled())
}
