package errors

import (
	"errors"
)

// ErrInvalidAppContext signals an invalid context passed to the routing system
var ErrInvalidAppContext = errors.New("invalid app context")

// ErrCouldNotGetAccount signals that a requested account could not be retrieved
var ErrCouldNotGetAccount = errors.New("could not get requested account")

// ErrGetBalance signals an error in getting the balance for an account
var ErrGetBalance = errors.New("get balance error")

// ErrEmptyAddress signals an empty address was provided
var ErrEmptyAddress = errors.New("address was empty")

// ErrNodeAlreadyRunning signals the node is already running
var ErrNodeAlreadyRunning = errors.New("node already running")

// ErrNodeAlreadyStopped signals the node is already stopped
var ErrNodeAlreadyStopped = errors.New("node already stopped")

// ErrCouldNotStopNode signals the node is already stopped
var ErrCouldNotStopNode = errors.New("could not stop node")

// ErrBadInitOfNode signals the node is could not be started correctly
var ErrBadInitOfNode = errors.New("bad init of node")

// ErrCouldNotParsePubKey signals that a given public key could not be parsed
var ErrCouldNotParsePubKey = errors.New("cound not parse node's public key")

// ErrValidation signals an error in validation
var ErrValidation = errors.New("validation error")

// ErrTxGenerationFailed signals an error generating a transaction
var ErrTxGenerationFailed = errors.New("transaction generation failed")

// ErrMultipleTxGenerationFailed signals an error generating multiple transactions
var ErrMultipleTxGenerationFailed = errors.New("multiple transaction generation failed")

// ErrInvalidSignatureHex signals a wrong hex value was provided for the signature
var ErrInvalidSignatureHex = errors.New("invalid signature, could not decode hex value")

// ErrValidationEmptyTxHash signals an empty tx hash was provided
var ErrValidationEmptyTxHash = errors.New("TxHash is empty")

// ErrGetTransaction signals an error happend trying to fetch a transaction
var ErrGetTransaction = errors.New("transaction getting failed")

// ErrTxNotFound signals an error happend trying to fetch a transaction
var ErrTxNotFound = errors.New("transaction was not found")

// ErrShardIdOutOfRange signals an error when shard id is out of range
var ErrShardIdOutOfRange = errors.New("shard id out of range")
