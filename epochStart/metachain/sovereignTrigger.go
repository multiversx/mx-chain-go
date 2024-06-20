package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
)

type sovereignTrigger struct {
	*trigger
}

func NewSovereignTrigger(metaTrigger *trigger) (*sovereignTrigger, error) {
	if check.IfNil(metaTrigger) {
		return nil, errorsMx.ErrNilEpochStartTrigger
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
