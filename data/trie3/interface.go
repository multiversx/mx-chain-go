package trie3

//Trie is an interface necessary for being able to use multiple implementations of Merkle Tries
type Trie interface {
	Get(key []byte) ([]byte, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	Root() ([]byte, error)
	Prove(key []byte) ([][]byte, error)
	VerifyProof(proofs [][]byte, key []byte) (bool, error)
	NodeIterator() NodeIterator
}

// NodeIterator is an iterator to traverse the trie pre-order.
type NodeIterator interface {
	// Next moves the iterator to the next node.
	Next() (bool, error)

	// Hash returns the hash of the current node.
	Hash() []byte

	// Parent returns the hash of the parent of the current node. The hash may be the one
	// grandparent.
	Parent() []byte

	// Path returns the hex-encoded path to the current node.
	// Callers must not retain references to the return value after calling Next.
	// For leaf nodes, the last element of the path is the 'terminator symbol' 0x10.
	Path() []byte

	// Leaf returns true iff the current node is a leaf node.
	Leaf() bool

	// LeafKey returns the key of the leaf. The method returns an error if the iterator is not
	// positioned at a leaf. Callers must not retain references to the value after
	// calling Next.
	LeafKey() ([]byte, error)

	// LeafBlob returns the content of the leaf. The method returns an error if the iterator
	// is not positioned at a leaf. Callers must not retain references to the value
	// after calling Next.
	LeafBlob() ([]byte, error)

	// LeafProof returns the Merkle proof of the leaf. The method returns an error if the
	// iterator is not positioned at a leaf. Callers must not retain references
	// to the value after calling Next.
	LeafProof() ([][]byte, error)
}

// DBWriteCacher is used to cache changes made to the trie, and only write to the database when it's needed
type DBWriteCacher interface {
	Put(key, val []byte) error
	Get(key []byte) ([]byte, error)
}
