package factory

import "errors"

// ErrNilConfiguration signals that a nil configuration has been provided
var ErrNilConfiguration = errors.New("nil configuration provided")

// ErrNilGenesisConfiguration signals that a nil genesis configuration has been provided
var ErrNilGenesisConfiguration = errors.New("nil genesis configuration provided")

// ErrNilCoreComponents signals that nil core components have been provided
var ErrNilCoreComponents = errors.New("nil core components provided")

// ErrNilShardCoordinator signals that nil core components have been provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator provided")

// ErrNilPathManager signals that a nil path manager has been provided
var ErrNilPathManager = errors.New("nil path manager provided")

// ErrHasherCreation signals that the hasher cannot be created based on provided data
var ErrHasherCreation = errors.New("error creating hasher")

// ErrMarshalizerCreation signals that the marshalizer cannot be created based on provided data
var ErrMarshalizerCreation = errors.New("error creating marshalizer")

// ErrPubKeyConverterCreation signals that the public key converter cannot be created based on provided data
var ErrPubKeyConverterCreation = errors.New("error creating public key converter")

// ErrAccountsAdapterCreation signals that the accounts adapter cannot be created based on provided data
var ErrAccountsAdapterCreation = errors.New("error creating accounts adapter")

// ErrInitialBalancesCreation signals that the initial balances cannot be created based on provided data
var ErrInitialBalancesCreation = errors.New("error creating initial balances")
