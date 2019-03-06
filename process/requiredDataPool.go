package process

import (
	"bytes"
	"sync"
)

// RequiredDataPool represents a structure that can hold a list of required data.
// Any time one of the expected hash value is received, the associated bit
// in the receivedBitmap property is set to 1. All of the data is
// considered received when the ones count of the received bitmap
// is equal to the expected data length
type RequiredDataPool struct {
	dataLock       sync.RWMutex
	expectedData   [][]byte
	receivedBitmap []byte
}

// ExpectedData returns the RequiredDataPool's expected data
func (rh *RequiredDataPool) ExpectedData() [][]byte {
	return rh.expectedData
}

// Reset unsets the expectedData and bitmap fields and set them to nil values
func (rh *RequiredDataPool) Reset() {
	rh.dataLock.Lock()
	rh.expectedData = nil
	rh.receivedBitmap = nil
	rh.dataLock.Unlock()
}
// SetHashes sets the expected data to the passed hashes parameter. The bitmap is also
// reset and adapted to the length of the new expected data
func (rh *RequiredDataPool) SetHashes(hashes [][]byte) {
	hashLength := len(hashes)
	if hashLength < 1 {
		rh.Reset()
		return
	}
	rh.dataLock.Lock()
	rh.expectedData = hashes
	rh.receivedBitmap = make([]byte, hashLength/8+1)
	rh.dataLock.Unlock()
}

// SetReceivedHash finds the hash in the expected values and sets the appropriate
// bit to 1. Nothing will happen if the hash is not actually expected
func (rh *RequiredDataPool) SetReceivedHash(hash []byte) {
	rh.dataLock.Lock()
	hashLength := len(rh.expectedData)
	for i := 0; i < hashLength; i++ {
		if bytes.Equal(rh.expectedData[i], hash) {
			rh.receivedBitmap[i/8] |= 1 << (uint16(i) % 8)
		}
	}
	rh.dataLock.Unlock()
}

// ReceivedAll will return true if the count of ones in the bitmap is greater
// or equal to the expected data length
func (rh *RequiredDataPool) ReceivedAll() bool {
	flags := 0

	rh.dataLock.Lock()

	bitmapLength := len(rh.receivedBitmap)
	dataLength := len(rh.expectedData)
	for i := 0; i < bitmapLength; i++ {
		for j := 0; j < 8; j++ {
			if rh.receivedBitmap[i]&(1<<uint8(j)) != 0 {
				flags++
			}
		}
	}
    rh.dataLock.Unlock()
	return flags >= dataLength
}
