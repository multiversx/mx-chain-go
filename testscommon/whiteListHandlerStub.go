package testscommon

import "github.com/multiversx/mx-chain-go/process"

// WhiteListHandlerStub -
type WhiteListHandlerStub struct {
	RemoveCalled                  func(keys [][]byte)
	AddCalled                     func(keys [][]byte)
	IsWhiteListedCalled           func(interceptedData process.InterceptedData) bool
	IsWhiteListedAtLeastOneCalled func(identifiers [][]byte) bool
	IsForCurrentShardCalled       func(interceptedData process.InterceptedData) bool
}

// IsWhiteListed -
func (w *WhiteListHandlerStub) IsWhiteListed(interceptedData process.InterceptedData) bool {
	if w.IsWhiteListedCalled != nil {
		return w.IsWhiteListedCalled(interceptedData)
	}
	return false
}

// IsWhiteListedAtLeastOne -
func (w *WhiteListHandlerStub) IsWhiteListedAtLeastOne(identifiers [][]byte) bool {
	if w.IsWhiteListedAtLeastOneCalled != nil {
		return w.IsWhiteListedAtLeastOneCalled(identifiers)
	}

	return false
}

// IsForCurrentShard -
func (w *WhiteListHandlerStub) IsForCurrentShard(interceptedData process.InterceptedData) bool {
	if w.IsForCurrentShardCalled != nil {
		return w.IsForCurrentShardCalled(interceptedData)
	}
	return true
}

// Remove -
func (w *WhiteListHandlerStub) Remove(keys [][]byte) {
	if w.RemoveCalled != nil {
		w.RemoveCalled(keys)
	}
}

// Add -
func (w *WhiteListHandlerStub) Add(keys [][]byte) {
	if w.AddCalled != nil {
		w.AddCalled(keys)
	}
}

// IsInterfaceNil -
func (w *WhiteListHandlerStub) IsInterfaceNil() bool {
	return w == nil
}
