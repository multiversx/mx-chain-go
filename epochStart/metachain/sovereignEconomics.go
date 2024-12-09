package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignEconomics struct {
	*economics
}

// NewSovereignEconomics creates a new sovereign end of epoch economics data object creator
func NewSovereignEconomics(ec *economics) (*sovereignEconomics, error) {
	if check.IfNil(ec) {
		return nil, process.ErrNilEconomicsData
	}

	ec.baseEconomicsHandler = &sovereignBaseEconomics{
		&baseEconomics{
			marshalizer:           ec.marshalizer,
			store:                 ec.store,
			shardCoordinator:      ec.shardCoordinator,
			economicsDataNotified: ec.economicsDataNotified,
			genesisEpoch:          ec.genesisEpoch,
			genesisNonce:          ec.genesisNonce,
		},
	}

	return &sovereignEconomics{
		ec,
	}, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (se *sovereignEconomics) IsInterfaceNil() bool {
	return se == nil
}
