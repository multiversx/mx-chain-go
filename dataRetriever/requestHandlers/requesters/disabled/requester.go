package disabled

import "github.com/multiversx/mx-chain-go/dataRetriever"

type requester struct {
}

// NewDisabledRequester returns a new instance of disabled requester
func NewDisabledRequester() *requester {
	return &requester{}
}

// RequestDataFromHash returns nil as it is disabled
func (r *requester) RequestDataFromHash(_ []byte, _ uint32) error {
	return nil
}

// SetNumPeersToQuery does nothing as it is disabled
func (r *requester) SetNumPeersToQuery(_ int, _ int) {
}

// NumPeersToQuery returns 0 as it is disabled
func (r *requester) NumPeersToQuery() (int, int) {
	return 0, 0
}

// SetDebugHandler returns nil as it is disabled
func (r *requester) SetDebugHandler(_ dataRetriever.DebugHandler) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *requester) IsInterfaceNil() bool {
	return r == nil
}
