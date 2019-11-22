package display

import (
	"encoding/hex"
	"sync"

	"github.com/ElrondNetwork/elrond-go/logger"
)

const ellipsisCharacter = "\u2026"

var mutDisplayByteSlice sync.RWMutex
var displayByteSlice func(slice []byte) string

func init() {
	mutDisplayByteSlice.Lock()
	displayByteSlice = func(slice []byte) string {
		return hex.EncodeToString(slice)
	}

	_ = logger.SetDisplayByteSlice(displayByteSlice)
	mutDisplayByteSlice.Unlock()
}

// SetDisplayByteSlice sets the converter function from byte slice to string
// default, this will call hex.EncodeToString. It will also change the logger's convert function
// so that the messages will be consistent
func SetDisplayByteSlice(f func(slice []byte) string) error {
	if f == nil {
		return ErrNilDisplayByteSliceHandler
	}

	mutDisplayByteSlice.Lock()
	displayByteSlice = f
	mutDisplayByteSlice.Unlock()

	return logger.SetDisplayByteSlice(f)
}

// DisplayByteSlice converts the provided byte slice to its string representation using
// displayByteSlice function pointer
func DisplayByteSlice(slice []byte) string {

	mutDisplayByteSlice.RLock()
	f := displayByteSlice
	mutDisplayByteSlice.RUnlock()

	return f(slice)
}

// ToHexShort generates a short-hand of provided bytes slice showing only the first 3 and the last 3 bytes as hex
// in total, the resulting string is maximum 13 characters long
func ToHexShort(slice []byte) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) < 6 {
		return hex.EncodeToString(slice)
	}

	prefix := hex.EncodeToString(slice[:3])
	suffix := hex.EncodeToString(slice[len(slice)-3:])
	return prefix + ellipsisCharacter + suffix
}
