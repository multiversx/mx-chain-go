package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type interceptedMetaBlockMock struct {
	HeaderHandlerToUse data.HeaderHandler
	HashToUse          []byte
}

// NewInterceptedMetaBlockMock -
func NewInterceptedMetaBlockMock(hdr data.HeaderHandler, hash []byte) *interceptedMetaBlockMock {
	return &interceptedMetaBlockMock{
		HeaderHandlerToUse: hdr,
		HashToUse:          hash,
	}
}

// HeaderHandler -
func (i *interceptedMetaBlockMock) HeaderHandler() data.HeaderHandler {
	return i.HeaderHandlerToUse
}

// CheckValidity -
func (i *interceptedMetaBlockMock) CheckValidity() error {
	return nil
}

// IsForCurrentShard -
func (i *interceptedMetaBlockMock) IsForCurrentShard() bool {
	return true
}

// IsInterfaceNil -
func (i *interceptedMetaBlockMock) IsInterfaceNil() bool {
	return i == nil
}

// Hash -
func (i *interceptedMetaBlockMock) Hash() []byte {
	return i.HashToUse
}

// Type -
func (i *interceptedMetaBlockMock) Type() string {
	return "type"
}
