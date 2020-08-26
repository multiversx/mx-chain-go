package node

import (
	"errors"
)

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("trying to set nil marshalizer")

// ErrNilAccountsAdapter signals that a nil accounts adapter has been provided
var ErrNilAccountsAdapter = errors.New("trying to set nil accounts adapter")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to use a nil pubkey converter")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("trying to set nil private key")

// ErrAccountNotFound signals that an account was not found in trie
var ErrAccountNotFound = errors.New("account not found")

// ErrZeroRoundDurationNotSupported signals that 0 seconds round duration is not supported
var ErrZeroRoundDurationNotSupported = errors.New("0 round duration time is not supported")

// ErrNegativeOrZeroConsensusGroupSize signals that 0 elements consensus group is not supported
var ErrNegativeOrZeroConsensusGroupSize = errors.New("group size should be a strict positive number")

// ErrNilDataPool signals that a nil data pool has been provided
var ErrNilDataPool = errors.New("trying to set nil data pool")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

// ErrNilSingleSig signals that a nil singlesig object has been provided
var ErrNilSingleSig = errors.New("trying to set nil singlesig")

// ErrNilMultiSig signals that a nil multiSigner object has been provided
var ErrNilMultiSig = errors.New("trying to set nil multiSigner")

// ErrNilResolversFinder signals that a nil resolvers finder has been provided
var ErrNilResolversFinder = errors.New("nil resolvers finder")

// ErrNilPeerDenialEvaluator signals that a nil peer denial evaluator was provided
var ErrNilPeerDenialEvaluator = errors.New("nil peer denial evaluator")

// ErrNilRequestedItemsHandler signals that a nil requested items handler was provided
var ErrNilRequestedItemsHandler = errors.New("nil requested items handler")

// ErrSystemBusyGeneratingTransactions signals that to many transactions are trying to get generated
var ErrSystemBusyGeneratingTransactions = errors.New("system busy while generating bulk transactions")

// ErrNoTxToProcess signals that no transaction were sent for processing
var ErrNoTxToProcess = errors.New("no transaction to process")

// ErrInvalidValue signals that an invalid value has been provided such as NaN to an integer field
var ErrInvalidValue = errors.New("invalid value")

// ErrNilNetworkShardingCollector defines the error for setting a nil network sharding collector
var ErrNilNetworkShardingCollector = errors.New("nil network sharding collector")

// ErrInvalidChainID signals that an invalid chain ID has been provided
var ErrInvalidChainID = errors.New("invalid chain ID in Node")

// ErrNilTxAccumulator signals that a nil Accumulator instance has been provided
var ErrNilTxAccumulator = errors.New("nil tx accumulator")

// ErrNilHardforkTrigger signals that a nil hardfork trigger has been provided
var ErrNilHardforkTrigger = errors.New("nil hardfork trigger")

// ErrNilWhiteListHandler signals that white list handler is nil
var ErrNilWhiteListHandler = errors.New("nil whitelist handler")

// ErrNilNodeStopChannel signals that a nil channel for node process stop has been provided
var ErrNilNodeStopChannel = errors.New("nil node stop channel")

// ErrNilQueryHandler signals that a nil query handler has been provided
var ErrNilQueryHandler = errors.New("nil query handler")

// ErrQueryHandlerAlreadyExists signals that the query handler is already registered
var ErrQueryHandlerAlreadyExists = errors.New("query handler already exists")

// ErrEmptyQueryHandlerName signals that an empty string can not be used to be used in the query handler container
var ErrEmptyQueryHandlerName = errors.New("empty query handler name")

// ErrUnknownPeerID signals that the provided peer is unknown by the current node
var ErrUnknownPeerID = errors.New("unknown peer ID")

// ErrNilWatchdog signals that a nil watchdog has been provided
var ErrNilWatchdog = errors.New("nil watchdog")

// ErrInvalidTransactionVersion signals  that an invalid transaction version has been provided
var ErrInvalidTransactionVersion = errors.New("invalid transaction version")

// ErrNilHistoryRepository signals that history repository is nil
var ErrNilHistoryRepository = errors.New("history repository is nil")

// ErrNilBootstrapComponents signals that a nil bootstrap components instance has been provided
var ErrNilBootstrapComponents = errors.New("nil bootstrap componennts")

// ErrNilCoreComponents signals that a nil core components instance has been provided
var ErrNilCoreComponents = errors.New("nil core components")

// ErrNilCryptoComponents signals that a nil crypto components instance has been provided
var ErrNilCryptoComponents = errors.New("nil crypto components")

// ErrNilDataComponents signals that a nil data components instance has been provided
var ErrNilDataComponents = errors.New("nil data components")

// ErrNilNetworkComponents signals that a nil network components instance has been provided
var ErrNilNetworkComponents = errors.New("nil network components")

// ErrNilProcessComponents signals that a nil process components instance has been provided
var ErrNilProcessComponents = errors.New("nil process components")

// ErrNilStateComponents signals that a nil state components instance has been provided
var ErrNilStateComponents = errors.New("nil state components")

// ErrNilStatusComponents signals that a nil status components instance has been provided
var ErrNilStatusComponents = errors.New("nil status components")
