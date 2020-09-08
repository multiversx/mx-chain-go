package track_test

import (
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockNotifier_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, err := track.NewBlockNotifier()

	assert.Nil(t, err)
	assert.NotNil(t, bn)
}

func TestRegisterHandler_ShouldNotRegisterNilHandler(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotifier()

	bn.RegisterHandler(nil)

	assert.Equal(t, 0, len(bn.GetNotarizedHeadersHandlers()))
}

func TestRegisterHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotifier()

	f1 := func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {}
	f2 := func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {}
	bn.RegisterHandler(f1)
	bn.RegisterHandler(f2)

	assert.Equal(t, 2, len(bn.GetNotarizedHeadersHandlers()))
}

func TestCallHandler_ShouldNotCallHandlersWhenHeadersSliceIsEmpty(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotifier()

	var callf1, callf2 bool
	f1 := func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
		callf1 = true
	}
	f2 := func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
		callf2 = true
	}

	bn.RegisterHandler(f1)
	bn.RegisterHandler(f2)
	bn.CallHandlers(0, nil, nil)

	assert.Equal(t, false, callf1)
	assert.Equal(t, false, callf2)
}

func TestCallHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotifier()

	wg := sync.WaitGroup{}
	wg.Add(2)

	var callf1, callf2 bool
	f1 := func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
		callf1 = true
		wg.Done()
	}
	f2 := func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
		callf2 = true
		wg.Done()
	}

	bn.RegisterHandler(f1)
	bn.RegisterHandler(f2)
	bn.CallHandlers(0, []data.HeaderHandler{&block.Header{}}, nil)

	wg.Wait()

	assert.Equal(t, true, callf1)
	assert.Equal(t, true, callf2)
}

func TestGetNumRegisteredHandlers_ShouldWork(t *testing.T) {
	t.Parallel()

	bn, _ := track.NewBlockNotifier()

	f1 := func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {}
	f2 := func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {}
	bn.RegisterHandler(f1)
	bn.RegisterHandler(f2)

	assert.Equal(t, 2, bn.GetNumRegisteredHandlers())
}
