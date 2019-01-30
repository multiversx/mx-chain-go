package trie2

type Trie interface {
	Get(key []byte) ([]byte, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	Root() []byte
	Copy() Trie
}

type Node interface {
	Hash() hashNode
}
