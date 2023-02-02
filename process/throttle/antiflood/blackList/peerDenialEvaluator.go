package blackList

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type peerDenialEvaluator struct {
	blackListIDsCache          process.PeerBlackListCacher
	blackListedPublicKeysCache process.TimeCacher
	peerShardMapper            process.PeerShardMapper
}

// NewPeerDenialEvaluator will create a new instance of a peer deny cache evaluator
func NewPeerDenialEvaluator(
	blackListIDsCache process.PeerBlackListCacher,
	blackListedPublicKeysCache process.TimeCacher,
	psm process.PeerShardMapper,
) (*peerDenialEvaluator, error) {

	if check.IfNil(blackListIDsCache) {
		return nil, fmt.Errorf("%w for peer IDs cacher", process.ErrNilBlackListCacher)
	}
	if check.IfNil(blackListedPublicKeysCache) {
		return nil, fmt.Errorf("%w for public keys cacher", process.ErrNilBlackListCacher)
	}
	if check.IfNil(psm) {
		return nil, process.ErrNilPeerShardMapper
	}

	return &peerDenialEvaluator{
		blackListIDsCache:          blackListIDsCache,
		blackListedPublicKeysCache: blackListedPublicKeysCache,
		peerShardMapper:            psm,
	}, nil
}

// IsDenied returns true if the provided peer id is denied to access the network
// It also checks if the provided peer id has a backing public key, checking also that the public key is not denied
func (pde *peerDenialEvaluator) IsDenied(pid core.PeerID) bool {
	if pde.blackListIDsCache.Has(pid) {
		return true
	}

	peerInfo := pde.peerShardMapper.GetPeerInfo(pid)
	pkBytes := peerInfo.PkBytes
	if len(pkBytes) == 0 {
		return false //no need to further search in the next cache, this is an unknown peer
	}

	return pde.blackListedPublicKeysCache.Has(string(pkBytes))
}

// UpsertPeerID will update or insert the provided peer id in the corresponding time cache
func (pde *peerDenialEvaluator) UpsertPeerID(pid core.PeerID, duration time.Duration) error {
	return pde.blackListIDsCache.Upsert(pid, duration)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pde *peerDenialEvaluator) IsInterfaceNil() bool {
	return pde == nil
}
