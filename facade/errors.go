package facade

import "github.com/pkg/errors"

// ErrHeartbeatsNotActive signals that the heartbeat system is not active
var ErrHeartbeatsNotActive = errors.New("heartbeat system not active")

// ErrNilNode signals that a nil node instance has been provided
var ErrNilNode = errors.New("nil node")

// ErrNilApiResolver signals that a nil api resolver instance has been provided
var ErrNilApiResolver = errors.New("nil api resolver")

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNoApiRoutesConfig signals that no configuration was found for API routes
var ErrNoApiRoutesConfig = errors.New("no configuration found for API routes")

// ErrNilPeerState signals that a nil peer state has been provided
var ErrNilPeerState = errors.New("nil peer state")

// ErrNilAccountState signals that a nil account state has been provided
var ErrNilAccountState = errors.New("nil account state")

// ErrNilTransactionSimulatorProcessor signals that a nil transaction simulator processor has been provided
var ErrNilTransactionSimulatorProcessor = errors.New("nil transaction simulator processor")

// ErrNilBlockchain signals that a nil blockchain has been provided
var ErrNilBlockchain = errors.New("nil blockchain")

// ErrEmptyRootHash signals that the current root hash is empty
var ErrEmptyRootHash = errors.New("empty current root hash")

// ErrNilGenesisNodes signals that the provided genesis nodes configuration is nil
var ErrNilGenesisNodes = errors.New("nil genesis nodes")

// ErrNilGenesisBalances signals that the provided genesis balances slice is nil
var ErrNilGenesisBalances = errors.New("nil genesis balances slice")

// ErrEmptyGasConfigs signals that the provided gas configs map is empty
var ErrEmptyGasConfigs = errors.New("empty gas configs")

// ErrTooManyAddressesInBulk signals that there are too many addresses present in a bulk request
var ErrTooManyAddressesInBulk = errors.New("too many addresses in the bulk request")

// ErrNilStatusMetrics signals that a nil status metrics was provided
var ErrNilStatusMetrics = errors.New("nil status metrics handler")
