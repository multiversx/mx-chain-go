package transaction

import "sync"

type txHashesHolder struct {
	sync.RWMutex
	hashes [][]byte
}

// NewTxHashesHolder returns a new instance of txHashesHolder
func NewTxHashesHolder() *txHashesHolder {
	return &txHashesHolder{
		hashes: make([][]byte, 0),
	}
}

// Append appends the provided hash into the internal hashes slice
func (holder *txHashesHolder) Append(hash []byte) {
	holder.Lock()
	defer holder.Unlock()

	holder.hashes = append(holder.hashes, hash)
}

// GetAllHashes returns all internal hashes
func (holder *txHashesHolder) GetAllHashes() [][]byte {
	holder.RLock()
	defer holder.RUnlock()

	return holder.hashes
}

// Reset resets the internal hashes slice
func (holder *txHashesHolder) Reset() {
	holder.Lock()
	defer holder.Unlock()

	holder.hashes = make([][]byte, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *txHashesHolder) IsInterfaceNil() bool {
	return holder == nil
}
