package state

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
)

// ErrAccountNotFoundAtBlock is an error-compatible struct holding the block info at which an account was not found
type ErrAccountNotFoundAtBlock struct {
	BlockInfo common.BlockInfo
}

// NewErrAccountNotFoundAtBlock returns a new error (custom struct)
func NewErrAccountNotFoundAtBlock(blockInfo common.BlockInfo) *ErrAccountNotFoundAtBlock {
	return &ErrAccountNotFoundAtBlock{BlockInfo: blockInfo}
}

// Error returns the error as string
func (e *ErrAccountNotFoundAtBlock) Error() string {
	return fmt.Sprintf("account was not found at block: nonce = %d, hash = %s, rootHash = %s",
		e.BlockInfo.GetNonce(),
		hex.EncodeToString(e.BlockInfo.GetHash()),
		hex.EncodeToString(e.BlockInfo.GetRootHash()),
	)
}

// ErrNilAccountsAdapter defines the error when trying to revert on nil accounts
var ErrNilAccountsAdapter = errors.New("nil AccountsAdapter")

// ErrNilAddress defines the error when trying to work with a nil address
var ErrNilAddress = errors.New("nil address")

// ErrEmptyAddress defines the error when trying to work with an empty address
var ErrEmptyAddress = errors.New("empty Address")

// ErrNilTrie signals that a trie is nil and no operation can be made
var ErrNilTrie = errors.New("trie is nil")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil hasher")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil marshalizer implementation
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNegativeValue signals that an operation has been attempted with a negative value
var ErrNegativeValue = errors.New("negative values are not permited")

// ErrNilAccountFactory signals that a nil account factory was provided
var ErrNilAccountFactory = errors.New("account factory is nil")

// ErrNilUpdater signals that a nil updater has been provided
var ErrNilUpdater = errors.New("updater is nil")

// ErrNilAccountHandler signals that a nil account wrapper was provided
var ErrNilAccountHandler = errors.New("account wrapper is nil")

// ErrNilShardCoordinator signals that nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("shard coordinator is nil")

// ErrWrongTypeAssertion signals that a wrong type assertion occurred
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilTrackableDataTrie signals that a nil trackable data trie has been provided
var ErrNilTrackableDataTrie = errors.New("nil trackable data trie")

// ErrAccNotFound signals that account was not found in state trie
var ErrAccNotFound = errors.New("account was not found")

// ErrNilBLSPublicKey signals that the provided BLS public key is nil
var ErrNilBLSPublicKey = errors.New("bls public key is nil")

// ErrNilOrEmptyDataTrieUpdates signals that there are no data trie updates
var ErrNilOrEmptyDataTrieUpdates = errors.New("no data trie updates")

// ErrOperationNotPermitted signals that operation is not permitted
var ErrOperationNotPermitted = errors.New("operation in account not permitted")

// ErrInvalidAddressLength signals that address length is invalid
var ErrInvalidAddressLength = errors.New("invalid address length")

// ErrInsufficientFunds signals the funds are insufficient for the move balance operation but the
// transaction fee is covered by the current balance
var ErrInsufficientFunds = errors.New("insufficient funds")

// ErrNilStorageManager signals that nil storage manager has been provided
var ErrNilStorageManager = errors.New("nil storage manager")

// ErrNilRequestHandler signals that nil request handler has been provided
var ErrNilRequestHandler = errors.New("nil request handler")

// ErrNilCacher signals that nil cacher has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrSnapshotValueOutOfBounds signals that the snapshot value is out of bounds
var ErrSnapshotValueOutOfBounds = errors.New("snapshot value out of bounds")

// ErrInvalidPubkeyConverterType signals that the provided pubkey converter type is invalid
var ErrInvalidPubkeyConverterType = errors.New("invalid pubkey converter type")

// ErrNilMapOfHashes signals that the provided map of hashes is nil
var ErrNilMapOfHashes = errors.New("nil map of hashes")

// ErrInvalidRootHash signals that the provided root hash is invalid
var ErrInvalidRootHash = errors.New("invalid root hash")

// ErrInvalidMaxHardCapForMissingNodes signals that the maximum hardcap value for missing nodes is invalid
var ErrInvalidMaxHardCapForMissingNodes = errors.New("invalid max hardcap for missing nodes")

// ErrNilEvictionWaitingList is raised when a nil eviction waiting list is provided
var ErrNilEvictionWaitingList = errors.New("nil eviction waiting list provided")

// ErrInvalidIdentifier signals that the root hash has an  invalid identifier
var ErrInvalidIdentifier = errors.New("invalid identifier")

// ErrNilStoragePruningManager signals that a nil storagePruningManager was provided
var ErrNilStoragePruningManager = errors.New("nil storagePruningManager")

// ErrInvalidKey is raised when the given key is invalid
var ErrInvalidKey = errors.New("invalid key")

// ErrNilRootHash signals that a nil root hash was provided
var ErrNilRootHash = errors.New("nil root hash")

// ErrNilProcessStatusHandler signals that a nil process status handler was provided
var ErrNilProcessStatusHandler = errors.New("nil process status handler")

// ErrNilAppStatusHandler signals that a nil app status handler was provided
var ErrNilAppStatusHandler = errors.New("nil app status handler")

// ErrNilBlockInfo signals that a nil block info was provided
var ErrNilBlockInfo = errors.New("nil block info")

// ErrNilBlockInfoProvider signals that a nil block info provider was provided
var ErrNilBlockInfoProvider = errors.New("nil block info provider")

// ErrFunctionalityNotImplemented signals that the functionality has not been implemented yet
var ErrFunctionalityNotImplemented = errors.New("functionality not implemented yet")

// ErrNilTrieSyncer signals that the trie syncer is nil
var ErrNilTrieSyncer = errors.New("trie syncer is nil")

// ErrNilSyncStatisticsHandler signals that a nil sync statistics handler was provided
var ErrNilSyncStatisticsHandler = errors.New("nil sync statistics handler")

// ErrNilAddressConverter signals that a nil address converter was provided
var ErrNilAddressConverter = errors.New("nil address converter")

// ErrNilEnableEpochsHandler signals that a nil enable epochs handler has been provided
var ErrNilEnableEpochsHandler = errors.New("nil enable epochs handler")
