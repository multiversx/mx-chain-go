package metachain

import "github.com/multiversx/mx-chain-go/process"

type economicsFactory struct {
}

// NewEconomicsFactory creates an end of epoch economics factory
func NewEconomicsFactory() *economicsFactory {
	return &economicsFactory{}
}

// CreateEndOfEpochEconomics creates end of epoch economics data creator for meta chain
func (f *economicsFactory) CreateEndOfEpochEconomics(args ArgsNewEpochEconomics) (process.EndOfEpochEconomics, error) {
	return NewEndOfEpochEconomicsDataCreator(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *economicsFactory) IsInterfaceNil() bool {
	return f == nil
}
