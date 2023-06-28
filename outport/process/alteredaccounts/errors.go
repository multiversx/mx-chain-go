package alteredaccounts

import "errors"

// errNilShardCoordinator signals that a nil shard coordinator has been provided
var errNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilPubKeyConverter signals that a nil public key converter has been provided
var ErrNilPubKeyConverter = errors.New("nil public key converter")

// ErrNilAccountsDB signals that a nil accounts DB has been provided
var ErrNilAccountsDB = errors.New("nil accounts DB")

// ErrNilESDTDataStorageHandler signals that a nil esdt data storage handler has been provided
var ErrNilESDTDataStorageHandler = errors.New("nil esdt data storage handler")

// errCannotCastToUserAccountHandler signals an issue while casting to user account handler
var errCannotCastToUserAccountHandler = errors.New("cannot cast account handler to vm common user account handler")

// errCannotCastToVmCommonUserAccountHandler signals an issue while casting to vm common user account handler
var errCannotCastToVmCommonUserAccountHandler = errors.New("cannot cast user account handler to vm common user account handler")

// ErrNilMetadata signals that a nil metadata has been provided
var ErrNilMetadata = errors.New("nil metadata provided")
