package pendingMb

import "github.com/multiversx/mx-chain-core-go/data"

// HeadersPool defines what a headers pool structure can perform
type HeadersPool interface {
	GetHeaderByHash(hash []byte) (data.HeaderHandler, error)
	IsInterfaceNil() bool
}
