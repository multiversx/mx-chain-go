package storage

import "github.com/multiversx/mx-chain-go/storage"

// ImportDBPeerStorer exposes the memory generation operations used by import-db tests.
type ImportDBPeerStorer interface {
	storage.Storer
	GetFromCurrentEpoch([]byte) ([]byte, error)
	Rotate()
	SnapshotFinished()
}
