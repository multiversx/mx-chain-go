package trie

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type trieSyncer struct {
	trie             *patriciaMerkleTrie
	resolver         dataRetriever.Resolver
	interceptedNodes storage.Cacher
	chRcvTrieNodes   chan bool
	waitTime         time.Duration

	requestedHashes      [][]byte
	requestedHashesMutex sync.Mutex
}

// NewTrieSyncer creates a new instance of trieSyncer
func NewTrieSyncer(
	resolver dataRetriever.Resolver,
	interceptedNodes storage.Cacher,
	trie data.Trie,
	waitTime time.Duration,
) (*trieSyncer, error) {
	if check.IfNil(resolver) {
		return nil, ErrNilResolver
	}
	if check.IfNil(interceptedNodes) {
		return nil, data.ErrNilCacher
	}
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}

	pmt, ok := trie.(*patriciaMerkleTrie)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return &trieSyncer{
		resolver:         resolver,
		interceptedNodes: interceptedNodes,
		trie:             pmt,
		chRcvTrieNodes:   make(chan bool),
		requestedHashes:  make([][]byte, 0),
		waitTime:         waitTime,
	}, nil
}

// StartSyncing completes the trie, asking for missing trie nodes on the network
func (ts *trieSyncer) StartSyncing(rootHash []byte) error {
	if len(rootHash) == 0 {
		return ErrInvalidHash
	}
	ts.interceptedNodes.RegisterHandler(ts.trieNodeIntercepted)

	nextNodes := make([]node, 0)

	currentNode, err := ts.getNode(rootHash)
	if err != nil {
		return err
	}

	ts.trie.root = currentNode
	err = ts.trie.root.loadChildren(ts)
	if err != nil {
		return err
	}

	nextNodes, err = ts.trie.root.getChildren()
	if err != nil {
		return err
	}

	for len(nextNodes) != 0 {
		currentNode, err = ts.getNode(nextNodes[0].getHash())
		if err != nil {
			return err
		}

		nextNodes = nextNodes[1:]

		err = currentNode.loadChildren(ts)
		if err != nil {
			return err
		}

		var children []node
		children, err = currentNode.getChildren()
		if err != nil {
			return err
		}
		nextNodes = append(nextNodes, children...)
	}

	err = ts.trie.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (ts *trieSyncer) getNode(hash []byte) (node, error) {
	n, ok := ts.interceptedNodes.Get(hash)
	if ok {
		return trieNode(n)
	}

	err := ts.requestNode(hash)
	if err != nil {
		return nil, err
	}

	n, _ = ts.interceptedNodes.Get(hash)
	return trieNode(n)
}

func trieNode(data interface{}) (node, error) {
	n, ok := data.(*InterceptedTrieNode)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return n.node, nil
}

func (ts *trieSyncer) requestNode(hash []byte) error {
	err := ts.resolver.RequestDataFromHash(hash)
	if err != nil {
		return err
	}

	receivedRequestedHashTrigger := append(hash, hash...)
	ts.requestedHashesMutex.Lock()
	ts.requestedHashes = append(ts.requestedHashes, receivedRequestedHashTrigger)
	ts.requestedHashesMutex.Unlock()

	return ts.waitForTrieNode()
}

func (ts *trieSyncer) waitForTrieNode() error {
	select {
	case <-ts.chRcvTrieNodes:
		return nil
	case <-time.After(ts.waitTime):
		return ErrTimeIsOut
	}
}

func (ts *trieSyncer) trieNodeIntercepted(hash []byte) {
	ts.requestedHashesMutex.Lock()

	if hashInSlice(hash, ts.requestedHashes) {
		ts.chRcvTrieNodes <- true
		ts.removeRequestedHash(hash)
	}
	ts.requestedHashesMutex.Unlock()
}

func (ts *trieSyncer) removeRequestedHash(hash []byte) {
	for i := range ts.requestedHashes {
		if bytes.Equal(ts.requestedHashes[i], hash) {
			ts.requestedHashes = append(ts.requestedHashes[:i], ts.requestedHashes[i+1:]...)
		}
	}
}

func hashInSlice(hash []byte, hashes [][]byte) bool {
	for _, h := range hashes {
		if bytes.Equal(h, hash) {
			return true
		}
	}
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *trieSyncer) IsInterfaceNil() bool {
	return ts == nil
}
