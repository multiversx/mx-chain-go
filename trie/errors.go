package trie

import (
	"errors"
)

// ErrInvalidNode is raised when we reach an invalid node
var ErrInvalidNode = errors.New("invalid node")

// ErrNilHasher is raised when the NewTrie() function is called, but a hasher isn't provided
var ErrNilHasher = errors.New("no hasher provided")

// ErrNilMarshalizer is raised when the NewTrie() function is called, but a marshalizer isn't provided
var ErrNilMarshalizer = errors.New("no marshalizer provided")

// ErrNilDatabase is raised when a database operation is called, but no database is provided
var ErrNilDatabase = errors.New("no database provided")

// ErrNilStorer is raised when a nil storer has been provided
var ErrNilStorer = errors.New("nil storer")

// ErrInvalidEncoding is raised when the encoded information cannot be decoded
var ErrInvalidEncoding = errors.New("cannot decode this invalid encoding")

// ErrValueTooShort is raised when we try to remove something from a value, and the value is too short
var ErrValueTooShort = errors.New("cannot remove bytes from value because value is too short")

// ErrChildPosOutOfRange is raised when the position of a child in a branch node is less than 0 or greater than 16
var ErrChildPosOutOfRange = errors.New("the position of the child is out of range")

// ErrNodeNotFound is raised when we try to get a node that is not present in the trie
var ErrNodeNotFound = errors.New("the node is not present in the trie")

// ErrEmptyBranchNode is raised when we reach an empty branch node (a node with no children)
var ErrEmptyBranchNode = errors.New("the branch node is empty")

// ErrEmptyExtensionNode is raised when we reach an empty extension node (a node with no child)
var ErrEmptyExtensionNode = errors.New("the extension node is empty")

// ErrEmptyLeafNode is raised when we reach an empty leaf node (a node with no value)
var ErrEmptyLeafNode = errors.New("the leaf node is empty")

// ErrNilBranchNode is raised when we reach a nil branch node
var ErrNilBranchNode = errors.New("the branch node is nil")

// ErrNilExtensionNode is raised when we reach a nil extension node
var ErrNilExtensionNode = errors.New("the extension node is nil")

// ErrNilLeafNode is raised when we reach a nil leaf node
var ErrNilLeafNode = errors.New("the leaf node is nil")

// ErrNilNode is raised when we reach a nil node
var ErrNilNode = errors.New("the node is nil")

// ErrWrongTypeAssertion signals that wrong type was provided
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilTrie is raised when the trie is nil
var ErrNilTrie = errors.New("the trie is nil")

// ErrNilRequestHandler is raised when the given request handler is nil
var ErrNilRequestHandler = errors.New("the request handler is nil")

// ErrTimeIsOut signals that time is out
var ErrTimeIsOut = errors.New("time is out")

// ErrNilTrieStorage is raised when a nil trie storage is provided
var ErrNilTrieStorage = errors.New("nil trie storage provided")

// ErrNilPathManager signals that a nil path manager has been provided
var ErrNilPathManager = errors.New("nil path manager")

// ErrInvalidTrieTopic signals that invalid trie topic has been provided
var ErrInvalidTrieTopic = errors.New("invalid trie topic")

// ErrNilContext signals that nil context has been provided
var ErrNilContext = errors.New("nil context")

// ErrInvalidLevelValue signals that the given value for maxTrieLevelInMemory is invalid
var ErrInvalidLevelValue = errors.New("invalid trie level in memory value")

// ErrNilTrieSyncStatistics signals that a nil trie sync statistics handler was provided
var ErrNilTrieSyncStatistics = errors.New("nil trie sync statistics handler")

// ErrNilTimeoutHandler signals that a nil timeout handler has been provided
var ErrNilTimeoutHandler = errors.New("nil timeout handler")

// ErrInvalidMaxHardCapForMissingNodes signals that the maximum hardcap value for missing nodes is invalid
var ErrInvalidMaxHardCapForMissingNodes = errors.New("invalid max hardcap for missing nodes")

// ErrInvalidTrieSyncerVersion signals that an invalid trie syncer version was provided
var ErrInvalidTrieSyncerVersion = errors.New("invalid trie syncer version")

// ErrTrieSyncTimeout signals that a timeout occurred while syncing the trie
var ErrTrieSyncTimeout = errors.New("trie sync timeout")

// ErrKeyNotFound is raised when a key is not found
var ErrKeyNotFound = errors.New("key not found")

// ErrNilIdleNodeProvider signals that a nil idle node provider was provided
var ErrNilIdleNodeProvider = errors.New("nil idle node provider")

// ErrNilRootHashHolder signals that a nil root hash holder was provided
var ErrNilRootHashHolder = errors.New("nil root hash holder provided")

// ErrNilTrieIteratorChannels signals that nil trie iterator channels has been provided
var ErrNilTrieIteratorChannels = errors.New("nil trie iterator channels")

// ErrNilTrieIteratorLeavesChannel signals that a nil trie iterator leaves channel has been provided
var ErrNilTrieIteratorLeavesChannel = errors.New("nil trie iterator leaves channel")

// ErrNilTrieIteratorErrChannel signals that a nil trie iterator error channel has been provided
var ErrNilTrieIteratorErrChannel = errors.New("nil trie iterator error channel")

// ErrInvalidIdentifier signals that an invalid identifier was provided
var ErrInvalidIdentifier = errors.New("invalid identifier")

// ErrNilKeyBuilder signals that a nil key builder has been provided
var ErrNilKeyBuilder = errors.New("nil key builder")

// ErrNilTrieLeafParser signals that a nil trie leaf parser has been provided
var ErrNilTrieLeafParser = errors.New("nil trie leaf parser")

// ErrInvalidNodeVersion signals that an invalid node version has been provided
var ErrInvalidNodeVersion = errors.New("invalid node version provided")
