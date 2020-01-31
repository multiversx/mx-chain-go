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

// ErrValidation signals an error in validation
var ErrValidation = errors.New("validation error")

// ErrTxGenerationFailed signals an error generating a transaction
var ErrTxGenerationFailed = errors.New("transaction generation failed")

// ErrInvalidSignatureHex signals a wrong hex value was provided for the signature
var ErrInvalidSignatureHex = errors.New("invalid signature, could not decode hex value")

// ErrValidationEmptyTxHash signals an empty tx hash was provided
var ErrValidationEmptyTxHash = errors.New("TxHash is empty")

// ErrGetTransaction signals an error happened trying to fetch a transaction
var ErrGetTransaction = errors.New("transaction getting failed")

// ErrTxNotFound signals an error happened trying to fetch a transaction
var ErrTxNotFound = errors.New("transaction was not found")

// ErrDupplicateThreshold signals that two thresholds are the same
var ErrDupplicateThreshold = errors.New("two thresholds are the same")

// ErrNoChancesForMaxThreshold signals that the max threshold has no chance defined
var ErrNoChancesForMaxThreshold = errors.New("max threshold has no chances")

// ErrNoChancesProvided signals that there were no chances provided
var ErrNoChancesProvided = errors.New("no chances are provided")
