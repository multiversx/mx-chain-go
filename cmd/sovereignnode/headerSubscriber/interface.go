package headerSubscriber

import "github.com/multiversx/mx-chain-core-go/data"

// HeadersPool should be able to add new headers in pool
type HeadersPool interface {
	AddHeader(headerHash []byte, header data.HeaderHandler)
	IsInterfaceNil() bool
}
