package process_test

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

func TestRequiredDataPool_ExpectedData(t *testing.T) {
	hashes := [][]byte{
		[]byte("first_data"),
		[]byte("second_data"),
	}
	rd := process.RequiredDataPool{}
	rd.SetHashes(hashes)

	assert.Equal(t, hashes, rd.ExpectedData())
}

func TestRequiredDataPool_SetHashesNilListResets(t *testing.T) {
	hashes := [][]byte{
		[]byte("first_data"),
		[]byte("second_data"),
	}
	rd := process.RequiredDataPool{}
	rd.SetHashes(hashes)
	rd.SetHashes(nil)

	assert.Nil(t, rd.ExpectedData())
}

func TestRequiredDataPool_SetHashesEmptyListResets(t *testing.T) {
	hashes := [][]byte{
		[]byte("first_data"),
		[]byte("second_data"),
	}
	rd := process.RequiredDataPool{}
	rd.SetHashes(hashes)
	rd.SetHashes(make([][]byte, 0))

	assert.Nil(t, rd.ExpectedData())
}

func TestRequiredDataPool_SetReceivedHashSettingAllWorks(t *testing.T) {
	hashes := [][]byte{
		[]byte("first_data"),
		[]byte("second_data"),
	}
	rd := process.RequiredDataPool{}
	rd.SetHashes(hashes)
	rd.SetReceivedHash([]byte("first_data"))
	rd.SetReceivedHash([]byte("second_data"))

	assert.True(t, rd.ReceivedAll())
}

func TestRequiredDataPool_SetReceivedHashSettingIncompleteWorks(t *testing.T) {
	hashes := [][]byte{
		[]byte("first_data"),
		[]byte("second_data"),
	}
	rd := process.RequiredDataPool{}
	rd.SetHashes(hashes)
	rd.SetReceivedHash([]byte("first_data"))

	assert.False(t, rd.ReceivedAll())
}

func TestRequiredDataPool_SetReceivedHashSettingWrongHashWorks(t *testing.T) {
	hashes := [][]byte{
		[]byte("first_data"),
		[]byte("second_data"),
	}
	rd := process.RequiredDataPool{}
	rd.SetHashes(hashes)
	rd.SetReceivedHash([]byte("first_data"))
	rd.SetReceivedHash([]byte("third_data"))

	assert.False(t, rd.ReceivedAll())
}

func TestRequiredDataPool_SetReceivedHashSettingWrongHashThenCorrectWorks(t *testing.T) {
	hashes := [][]byte{
		[]byte("first_data"),
		[]byte("second_data"),
	}
	rd := process.RequiredDataPool{}
	rd.SetHashes(hashes)
	rd.SetReceivedHash([]byte("first_data"))
	rd.SetReceivedHash([]byte("third_data"))
	rd.SetReceivedHash([]byte("second_data"))

	assert.True(t, rd.ReceivedAll())
}

func TestRequiredDataPool_SetReceivedHashMultipleSettingAllWorks(t *testing.T) {
	maxHashes := 8

	hashes := make([][]byte, maxHashes)

	for idx := 0; idx < maxHashes; idx++ {
		hashes[idx] = []byte("hash " + strconv.Itoa(idx))
	}

	rd := process.RequiredDataPool{}
	rd.SetHashes(hashes)

	for idx := 0; idx < maxHashes; idx++ {
		rd.SetReceivedHash([]byte("hash " + strconv.Itoa(idx)))
	}

	assert.True(t, rd.ReceivedAll())
}