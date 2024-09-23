package receiptslog

import "errors"

// ErrNilTrieInteractor signals that a nil trie interactor has been provided
var ErrNilTrieInteractor = errors.New("trie interactor is nil")

// ErrNilReceiptsDataSyncer signals that a nil receipt data syncer has been provided
var ErrNilReceiptsDataSyncer = errors.New("receipts data syncer is nil")
