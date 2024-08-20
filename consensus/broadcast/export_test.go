package broadcast

import (
	"github.com/multiversx/mx-chain-core-go/marshal"
)

func (scm *shardChainMessenger) SetMarshalizer(
	m marshal.Marshalizer,
) {
	scm.marshalizer = m
}

func (mcm *metaChainMessenger) SetMarshalizerMeta(
	m marshal.Marshalizer,
) {
	mcm.marshalizer = m
}
