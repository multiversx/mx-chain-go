package broadcast

import (
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// Function to set a different Marshalizer for metaChainMessenger
func (mcm *metaChainMessenger) SetMarshalizerMeta(
	m marshal.Marshalizer,
) {
	mcm.marshalizer = m
}
