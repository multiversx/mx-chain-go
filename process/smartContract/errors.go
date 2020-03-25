package smartContract

import "github.com/pkg/errors"

// ErrNotEnoughArgumentsToDeploy signals that there are not enough arguments to deploy the smart contract
var ErrNotEnoughArgumentsToDeploy = errors.New("not enough arguments to deploy the smart contract")

// ErrInvalidVMType signals an invalid VMType
var ErrInvalidVMType = errors.New("invalid vm type")

// ErrInvalidCodeMetadata signals an invalid code metadata
var ErrInvalidCodeMetadata = errors.New("invalid code metadata")
