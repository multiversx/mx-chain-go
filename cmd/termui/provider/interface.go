package provider

import (
	"github.com/ElrondNetwork/elrond-go/cmd/termui/view"
)

// PresenterHandler defines what a component which will handle the presentation of data in the termui should do
type PresenterHandler interface {
	Increment(key string)
	AddUint64(key string, val uint64)
	Decrement(key string)
	SetInt64Value(key string, value int64)
	SetUInt64Value(key string, value uint64)
	SetStringValue(key string, value string)
	Close()
	Write(p []byte) (n int, err error)
	view.Presenter
}
