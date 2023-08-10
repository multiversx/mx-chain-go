package rating

import "github.com/multiversx/mx-chain-go/process"

// RatingsDataFactory defines a RatingsInfoHandler factory behavior
type RatingsDataFactory interface {
	CreateRatingsData(args RatingsDataArg) (process.RatingsInfoHandler, error)
	IsInterfaceNil() bool
}
