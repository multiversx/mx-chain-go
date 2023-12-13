package external

import "github.com/pkg/errors"

// ErrNilSCQueryService signals that a nil SC query service has been provided
var ErrNilSCQueryService = errors.New("nil SC query service")

// ErrNilStatusMetrics signals that a nil status metrics was provided
var ErrNilStatusMetrics = errors.New("nil status metrics handler")

// ErrNilAPITransactionEvaluator signals that a nil api transaction evaluator was provided
var ErrNilAPITransactionEvaluator = errors.New("nil api transaction evaluator")

// ErrNilTotalStakedValueHandler signals that a nil total staked value handler has been provided
var ErrNilTotalStakedValueHandler = errors.New("nil total staked value handler")

// ErrNilDirectStakeListHandler signals that a nil stake list handler has been provided
var ErrNilDirectStakeListHandler = errors.New("nil direct stake list handler")

// ErrNilDelegatedListHandler signals that a nil delegated list handler has been provided
var ErrNilDelegatedListHandler = errors.New("nil delegated list handler")

// ErrNilAPITransactionHandler signals that a nil api transaction handler has been provided
var ErrNilAPITransactionHandler = errors.New("nil api transaction handler")

// ErrNilAPIBlockHandler signals that a nil api block handler has been provided
var ErrNilAPIBlockHandler = errors.New("nil api block handler")

// ErrNilAPIInternalBlockHandler signals that a nil api internal block handler has been provided
var ErrNilAPIInternalBlockHandler = errors.New("nil api internal block handler")

// ErrNilGenesisNodesSetupHandler signals that a nil genesis nodes setup handler has been provided
var ErrNilGenesisNodesSetupHandler = errors.New("nil genesis nodes setup handler")

// ErrNilValidatorPubKeyConverter signals that a nil validator pubkey converter has been provided
var ErrNilValidatorPubKeyConverter = errors.New("nil validator public key converter")

// ErrNilAccountsParser signals that a nil accounts parser has been provided
var ErrNilAccountsParser = errors.New("nil accounts parser")

// ErrNilGasScheduler signals that a nil gas scheduler has been provided
var ErrNilGasScheduler = errors.New("nil gas scheduler")

// ErrNilManagedPeersMonitor signals that a nil managed peers monitor has been provided
var ErrNilManagedPeersMonitor = errors.New("nil managed peers monitor")

// ErrNilNodesCoordinator signals a nil nodes coordinator has been provided
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")
