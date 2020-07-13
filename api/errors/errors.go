package errors

import (
	"errors"
)

// ErrInvalidAppContext signals an invalid context passed to the routing system
var ErrInvalidAppContext = errors.New("invalid app context")

// ErrInvalidJSONRequest signals an error in json request formatting
var ErrInvalidJSONRequest = errors.New("invalid json request")

// ErrCouldNotGetAccount signals that a requested account could not be retrieved
var ErrCouldNotGetAccount = errors.New("could not get requested account")

// ErrGetBalance signals an error in getting the balance for an account
var ErrGetBalance = errors.New("get balance error")

// ErrGetValueForKey signals an error in getting the value of a key for an account
var ErrGetValueForKey = errors.New("get value for key error")

// ErrEmptyAddress signals an empty address was provided
var ErrEmptyAddress = errors.New("address is empty")

// ErrEmptyKey signals an empty key was provided
var ErrEmptyKey = errors.New("key is empty")

// ErrValidation signals an error in validation
var ErrValidation = errors.New("validation error")

// ErrTxGenerationFailed signals an error generating a transaction
var ErrTxGenerationFailed = errors.New("transaction generation failed")

// ErrValidationEmptyTxHash signals an empty tx hash was provided
var ErrValidationEmptyTxHash = errors.New("TxHash is empty")

// ErrGetTransaction signals an error happened trying to fetch a transaction
var ErrGetTransaction = errors.New("transaction getting failed")

// ErrQueryError signals a general query error
var ErrQueryError = errors.New("query error")

// ErrGetPidInfo signals that an error occurred while getting peer ID info
var ErrGetPidInfo = errors.New("error getting peer id info")

// ErrTooManyRequests signals that too many requests were simultaneously received
var ErrTooManyRequests = errors.New("too many requests")
