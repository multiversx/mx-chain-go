package external

import "github.com/pkg/errors"

// ErrNilSCQueryService signals that a nil SC query service has been provided
var ErrNilSCQueryService = errors.New("nil SC query service")

// ErrNilStatusMetrics signals that a nil status metrics was provided
var ErrNilStatusMetrics = errors.New("nil status metrics handler")

// ErrNilTransactionCostHandler signals that a nil transaction cost handler was provided
var ErrNilTransactionCostHandler = errors.New("nil transaction cost handler")

// ErrNilTotalStakedValueHandler signals that a nil total staked value handler has been provided
var ErrNilTotalStakedValueHandler = errors.New("nil total staked value handler")

// ErrNilDirectStakeListHandler signals that a nil stake list handler has been provided
var ErrNilDirectStakeListHandler = errors.New("nil direct stake list handler")

// ErrNilDelegatedListHandler signals that a nil delegated list handler has been provided
var ErrNilDelegatedListHandler = errors.New("nil delegated list handler")

// ErrNilVmContainer signals that a nil vm container has been provided
var ErrNilVmContainer = errors.New("nil vm container")

// ErrNilVmFactory signals that a nil vm factory has been provided
var ErrNilVmFactory = errors.New("nil vm factory")
