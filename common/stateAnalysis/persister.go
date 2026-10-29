package stateAnalysis

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-storage-go/types"
)

const slowGetThreshold = time.Millisecond

// UnitFromIdentifier maps a storer identifier (UserAccountsUnit, PeerAccountsUnit) or a db path
// (.../AccountsTrie, .../PeerAccountsTrie) to a storage unit index
func UnitFromIdentifier(identifier string) int {
	base := filepath.Base(identifier)
	switch {
	case strings.Contains(base, "PeerAccounts"):
		return UnitPeer
	case strings.Contains(base, "Accounts"):
		return UnitAccounts
	default:
		return UnitOther
	}
}

type timedPersister struct {
	types.Persister
	unit int
}

// WrapPersister decorates a persister so that every Get and Put is timed and counted for the given unit
func WrapPersister(p types.Persister, identifier string) types.Persister {
	if !Enabled() || p == nil {
		return p
	}
	return &timedPersister{Persister: p, unit: UnitFromIdentifier(identifier)}
}

// Get times the underlying Get
func (tp *timedPersister) Get(key []byte) ([]byte, error) {
	start := time.Now()
	val, err := tp.Persister.Get(key)
	elapsed := time.Since(start)

	Inc(CStDbGet.unit(tp.unit))
	Add(CStDbGetNs.unit(tp.unit), int64(elapsed))
	Max(MStDbGetNs, int64(elapsed))
	if elapsed > slowGetThreshold {
		Inc(CStDbGetSlow.unit(tp.unit))
	}
	if err == nil {
		Inc(CStDbGetHit.unit(tp.unit))
		Add(CStDbGetBytes.unit(tp.unit), int64(len(val)))
	}
	return val, err
}

// Put times the underlying Put
func (tp *timedPersister) Put(key, val []byte) error {
	start := time.Now()
	err := tp.Persister.Put(key, val)
	Inc(CStDbPut.unit(tp.unit))
	Add(CStDbPutNs.unit(tp.unit), int64(time.Since(start)))
	Add(CStDbPutBytes.unit(tp.unit), int64(len(val)))
	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (tp *timedPersister) IsInterfaceNil() bool {
	return tp == nil
}
