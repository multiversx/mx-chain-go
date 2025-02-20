package rating

import "github.com/multiversx/mx-chain-go/process"

type sovereignRatingsDataFactory struct {
}

// NewSovereignRatingsDataFactory creates a ratings data factory for sovereign chain
func NewSovereignRatingsDataFactory() *sovereignRatingsDataFactory {
	return &sovereignRatingsDataFactory{}
}

// CreateRatingsData creates ratings info data for sovereign chain
func (rdf *sovereignRatingsDataFactory) CreateRatingsData(args RatingsDataArg) (process.RatingsInfoHandler, error) {
	return NewSovereignRatingsData(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (rdf *sovereignRatingsDataFactory) IsInterfaceNil() bool {
	return rdf == nil
}
