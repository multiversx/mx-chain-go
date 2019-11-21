package state

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

// ErrorWrongSize is an error-compatible struct holding 2 values: Expected and Got
type ErrorWrongSize struct {
	Exp int
	Got int
}

// ErrorTrieNotNormalized is an error-compatible struct holding the hash length that is not normalized
type ErrorTrieNotNormalized struct {
	actualHashLength   int
	expectedHashLength int
}

// ErrMissingTrie is an error-compatible struct holding the root hash of the trie that is missing
type ErrMissingTrie struct {
	rootHash []byte
}

//------- ErrorWrongSize

// NewErrorWrongSize returns a new instantiated struct
func NewErrorWrongSize(exp int, got int) *ErrorWrongSize {
	return &ErrorWrongSize{Exp: exp, Got: got}
}

// Error returns the error as string
func (e *ErrorWrongSize) Error() string {
	return fmt.Sprintf("wrong size, expected: %d, got %d", e.Exp, e.Got)
}

//------- ErrorTrieNotNormalized

// NewErrorTrieNotNormalized returns a new instantiated struct
func NewErrorTrieNotNormalized(exp int, actual int) *ErrorTrieNotNormalized {
	return &ErrorTrieNotNormalized{expectedHashLength: exp, actualHashLength: actual}
}

// Error returns the error as string
func (e *ErrorTrieNotNormalized) Error() string {
	return "attempt to search a hash not normalized to " +
		strconv.Itoa(e.expectedHashLength) + " bytes (has: " +
		strconv.Itoa(e.actualHashLength) + ")"
}

//------- ErrMissingTrie

// NewErrMissingTrie  returns a new instantiated struct
func NewErrMissingTrie(rootHash []byte) *ErrMissingTrie {
	return &ErrMissingTrie{rootHash: rootHash}
}

// Error returns the error as string
func (e *ErrMissingTrie) Error() string {
	return "trie was not found for hash " + base64.StdEncoding.EncodeToString(e.rootHash)
}

// ErrNilAccountsAdapter defines the error when trying to revert on nil accounts
var ErrNilAccountsAdapter = errors.New("nil AccountsAdapter")

// ErrNilAddressContainer defines the error when trying to work with a nil address
var ErrNilAddressContainer = errors.New("nil AddressContainer")

// ErrEmptyAddress defines the error when trying to work with an empty address
var ErrEmptyAddress = errors.New("empty Address")

// ErrNilTrie signals that a trie is nil and no operation can be made
var ErrNilTrie = errors.New("trie is nil")

// ErrNilValue signals that an operation has been attempted to or with a nil value
var ErrNilValue = errors.New("nil value")

// ErrNilPubKeysBytes signals that an operation has been attempted to or with a nil public key slice
var ErrNilPubKeysBytes = errors.New("nil public key bytes")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil hasher")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil marshalizer implementation
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNegativeValue signals that an operation has been attempted with a negative value
var ErrNegativeValue = errors.New("negative values are not permited")

// ErrNilAccountFactory signals that a nil account factory was provided
var ErrNilAccountFactory = errors.New("account factory is nil")

// ErrNilAccountTracker signals that a nil account tracker has been provided
var ErrNilAccountTracker = errors.New("nil account tracker provided")

// ErrNilUpdater signals that a nil updater has been provided
var ErrNilUpdater = errors.New("updater is nil")

// ErrNilAccountHandler signals that a nil account wrapper was provided
var ErrNilAccountHandler = errors.New("account wrapper is nil")

// ErrNilOrEmptyKey signals that key empty key was provided
var ErrNilOrEmptyKey = errors.New("key is empty or nil")

// ErrNilShardCoordinator signals that nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("shard coordinator is nil")

// ErrWrongTypeAssertion signals that a wrong type assertion occurred
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilTrackableDataTrie signals that a nil trackable data trie has been provided
var ErrNilTrackableDataTrie = errors.New("nil trackable data trie")

// ErrNilCode signals that a nil code was provided
var ErrNilCode = errors.New("nil smart contract code")

// ErrAccNotFound signals that account was not found in state trie
var ErrAccNotFound = errors.New("account was not found")

// ErrUnknownShardId signals that shard id is not valid
var ErrUnknownShardId = errors.New("shard id is not valid")

// ErrNilAddressConverter signals that a nil address converter has been provided
var ErrNilAddressConverter = errors.New("nil address converter")

//ErrBech32ConvertError signals that conversion the 5bit alphabet to 8bit failed
var ErrBech32ConvertError = errors.New("can't convert bech32 string")

// ErrBech32WrongAddr signals that the string provided might not be in bech32 format
var ErrBech32WrongAddr = errors.New("wrong bech32 string")

// ErrNilStake signals that the provided stake is nil
var ErrNilStake = errors.New("stake is nil")

// ErrNilSchnorrPublicKey signals that the provided schnorr public is nil
var ErrNilSchnorrPublicKey = errors.New("schnorr public key is nil")

// ErrNilBLSPublicKey signals that the provided BLS public key is nil
var ErrNilBLSPublicKey = errors.New("bls public key is nil")

// ErrUnknownAccountType signals that the provided account type is unknown
var ErrUnknownAccountType = errors.New("account type is unknown")
