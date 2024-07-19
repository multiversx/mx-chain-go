package metachain

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

type sovereignTrigger struct {
	*trigger
}

// NewSovereignTrigger creates a new sovereign epoch start trigger
func NewSovereignTrigger(args *ArgsNewMetaEpochStartTrigger) (*sovereignTrigger, error) {
	metaTrigger, err := newTrigger(args, &block.SovereignChainHeader{}, &sovereignTriggerRegistryCreator{}, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return &sovereignTrigger{
		trigger: metaTrigger,
	}, nil
}

// SetProcessed sets start of epoch to false and cleans underlying structure
func (st *sovereignTrigger) SetProcessed(header data.HeaderHandler, body data.BodyHandler) {
	st.mutTrigger.Lock()
	defer st.mutTrigger.Unlock()

	sovChainHeader, ok := header.(*block.SovereignChainHeader)
	if !ok {
		log.Error("sovereignTrigger.trigger", "error", data.ErrInvalidTypeAssertion)
		return
	}

	st.baseSetProcessed(sovChainHeader, body)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (st *sovereignTrigger) IsInterfaceNil() bool {
	return st == nil
}
