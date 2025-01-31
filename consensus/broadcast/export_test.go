package broadcast

import (
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// SetMarshalizerMeta sets the unexported marshaller
func (mcm *metaChainMessenger) SetMarshalizerMeta(
	m marshal.Marshalizer,
) {
	mcm.marshalizer = m
}
