package headerSubscriber

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type incomingHeaderHandler struct {
	headersPool HeadersPool
}

// NewIncomingHeaderHandler creates an incoming header handler which should be able to receive incoming headers and events
// from a chain to local sovereign chain. This handler will validate the events(using proofs in the future) and create
// incoming miniblocks and transaction(which will be added in pool) to be executed in sovereign shard.
func NewIncomingHeaderHandler(headersPool HeadersPool) (*incomingHeaderHandler, error) {
	if check.IfNil(headersPool) {
		return nil, errNilHeadersPool
	}

	return &incomingHeaderHandler{
		headersPool: headersPool,
	}, nil
}

// AddHeader will receive the incoming header, validate it, create incoming mbs and transactions and add them to pool
func (ihs *incomingHeaderHandler) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) {

}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihs *incomingHeaderHandler) IsInterfaceNil() bool {
	return ihs == nil
}
