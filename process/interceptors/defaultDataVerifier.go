package interceptors

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
)

type defaultDataVerifier struct {
}

// NewDefaultDataVerifier returns a default data verifier
func NewDefaultDataVerifier() *defaultDataVerifier {
	return &defaultDataVerifier{}
}

// IsForCurrentShard return true if intercepted data is for shard
func (d *defaultDataVerifier) IsForCurrentShard(interceptedData process.InterceptedData) bool {
	if check.IfNil(interceptedData) {
		return false
	}
	return interceptedData.IsForCurrentShard()
}

// IsInterfaceNil returns true if underlying object is nil
func (d *defaultDataVerifier) IsInterfaceNil() bool {
	return d == nil
}
