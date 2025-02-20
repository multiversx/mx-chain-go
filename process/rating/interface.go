package rating

import "github.com/multiversx/mx-chain-go/process"

// RatingsDataFactory defines a ratings info handler factory behavior
type RatingsDataFactory interface {
	CreateRatingsData(args RatingsDataArg) (process.RatingsInfoHandler, error)
	IsInterfaceNil() bool
}
