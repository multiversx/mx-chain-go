package dataValidators

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("addressBlackList")

// addressBlacklist holds the blacklist data
type addressBlacklist struct {
	addresses       map[string]struct{}
	addrMut         sync.RWMutex
	pubKeyConverter core.PubkeyConverter
}

// NewAddressBlacklist creates a new Address blacklist instance
func NewAddressBlacklist(addresses []string, pubKeyConverter core.PubkeyConverter) (*addressBlacklist, error) {
	if check.IfNil(pubKeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}

	addrMap := make(map[string]struct{})
	for i := range addresses {
		addrBytes, err := pubKeyConverter.Decode(addresses[i])
		if err != nil {
			log.Warn("NewAddressBlacklist - error converting address", "address", addresses[i], "error", err)
			continue
		}
		addrMap[string(addrBytes)] = struct{}{}
	}

	return &addressBlacklist{
		pubKeyConverter: pubKeyConverter,
		addresses:       addrMap,
	}, nil
}

// IsBlacklisted returns true if the address is blacklisted
func (ab *addressBlacklist) IsBlacklisted(addrBytes []byte) bool {
	ab.addrMut.RLock()
	defer ab.addrMut.RUnlock()

	_, ok := ab.addresses[string(addrBytes)]
	return ok
}

// IsInterfaceNil returns true if the receiver is nil
func (ab *addressBlacklist) IsInterfaceNil() bool {
	return ab == nil
}
