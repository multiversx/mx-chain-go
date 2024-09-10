package receiptslog

import "errors"

// ErrNilTrieInteractor signals that a nil trie interactor has been provided
var ErrNilTrieInteractor = errors.New("trie interactor is nil")
