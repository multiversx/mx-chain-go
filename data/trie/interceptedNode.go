package trie

import (
	"fmt"
	"math/big"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.TxValidatorHandler = (*InterceptedTrieNode)(nil)
var _ process.InterceptedData = (*InterceptedTrieNode)(nil)

// InterceptedTrieNode implements intercepted data interface and is used when trie nodes are intercepted
type InterceptedTrieNode struct {
	node           node
	serializedNode []byte
	hash           []byte

	mutex sync.RWMutex
}

// NewInterceptedTrieNode creates a new instance of InterceptedTrieNode
func NewInterceptedTrieNode(
	buff []byte,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (*InterceptedTrieNode, error) {
	if len(buff) == 0 {
		return nil, ErrValueTooShort
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	n, err := decodeNode(buff, marshalizer, hasher)
	if err != nil {
		return nil, err
	}
	n.setDirty(true)

	err = n.setHash()
	if err != nil {
		return nil, err
	}

	return &InterceptedTrieNode{
		node:           n,
		serializedNode: buff,
		hash:           n.getHash(),
	}, nil
}

// CheckValidity checks if the intercepted data is valid
func (inTn *InterceptedTrieNode) CheckValidity() error {
	if inTn.node.isValid() {
		return nil
	}
	return ErrInvalidNode
}

// IsForCurrentShard checks if the intercepted data is for the current shard
func (inTn *InterceptedTrieNode) IsForCurrentShard() bool {
	return true
}

// Hash returns the hash of the intercepted node
func (inTn *InterceptedTrieNode) Hash() []byte {
	return inTn.hash
}

// IsInterfaceNil returns true if there is no value under the interface
func (inTn *InterceptedTrieNode) IsInterfaceNil() bool {
	return inTn == nil
}

// GetSerialized returns the intercepted encoded node
func (inTn *InterceptedTrieNode) GetSerialized() []byte {
	inTn.mutex.RLock()
	defer inTn.mutex.RUnlock()

	return inTn.serializedNode
}

// SetSerialized sets the given bytes as the SerializedNode
func (inTn *InterceptedTrieNode) SetSerialized(serializedNode []byte) {
	inTn.mutex.Lock()
	defer inTn.mutex.Unlock()

	inTn.serializedNode = serializedNode
}

// Type returns the type of this intercepted data
func (inTn *InterceptedTrieNode) Type() string {
	return "intercepted trie node"
}

// String returns the trie node's most important fields as string
func (inTn *InterceptedTrieNode) String() string {
	return fmt.Sprintf("hash=%s",
		logger.DisplayByteSlice(inTn.hash),
	)
}

// SenderShardId returns 0
func (inTn *InterceptedTrieNode) SenderShardId() uint32 {
	return 0
}

// ReceiverShardId returns 0
func (inTn *InterceptedTrieNode) ReceiverShardId() uint32 {
	return 0
}

// Nonce return 0
func (inTn *InterceptedTrieNode) Nonce() uint64 {
	return 0
}

// SenderAddress returns nil
func (inTn *InterceptedTrieNode) SenderAddress() []byte {
	return nil
}

// Fee returns big.NewInt(0)
func (inTn *InterceptedTrieNode) Fee() *big.Int {
	return big.NewInt(0)
}

// SizeInBytes returns the size in bytes held by this instance plus the inner node's instance size
func (inTn *InterceptedTrieNode) SizeInBytes() int {
	inTn.mutex.RLock()
	defer inTn.mutex.RUnlock()

	return len(inTn.hash) + len(inTn.serializedNode) + inTn.node.sizeInBytes()
}

// Identifiers returns the identifiers used in requests
func (inTn *InterceptedTrieNode) Identifiers() [][]byte {
	return [][]byte{inTn.hash}
}
