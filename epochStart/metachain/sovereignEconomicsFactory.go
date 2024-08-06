package metachain

import "github.com/multiversx/mx-chain-go/process"

type sovereignEconomicsFactory struct {
}

// NewSovereignEconomicsFactory creates an end of epoch economics factory for sovereign chain
func NewSovereignEconomicsFactory() *sovereignEconomicsFactory {
	return &sovereignEconomicsFactory{}
}

// CreateEndOfEpochEconomics creates end of epoch economics data creator for sovereign chain
func (f *sovereignEconomicsFactory) CreateEndOfEpochEconomics(args ArgsNewEpochEconomics) (process.EndOfEpochEconomics, error) {
	ec, err := NewEndOfEpochEconomicsDataCreator(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignEconomics(ec)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignEconomicsFactory) IsInterfaceNil() bool {
	return f == nil
}
