package rating

import "github.com/multiversx/mx-chain-go/process"

type ratingsDataFactory struct {
}

// NewRatingsDataFactory creates a ratings data factory for regular chain
func NewRatingsDataFactory() *ratingsDataFactory {
	return &ratingsDataFactory{}
}

// CreateRatingsData creates ratings info data for regular chain
func (rdf *ratingsDataFactory) CreateRatingsData(args RatingsDataArg) (process.RatingsInfoHandler, error) {
	return NewRatingsData(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (rdf *ratingsDataFactory) IsInterfaceNil() bool {
	return rdf == nil
}
