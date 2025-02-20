package preprocess

import (
	"math/big"
	"sync"
)

type accountInfo struct {
	nonce   uint64
	balance *big.Int
}

type accountsTracker struct {
	accountInfo    map[string]accountInfo
	mutAccountInfo sync.RWMutex
}

func newAccountsTracker() *accountsTracker {
	return &accountsTracker{
		accountInfo:    make(map[string]accountInfo),
		mutAccountInfo: sync.RWMutex{},
	}
}

func (at *accountsTracker) init() {
	at.mutAccountInfo.Lock()
	defer at.mutAccountInfo.Unlock()

	at.accountInfo = make(map[string]accountInfo)
}

func (at *accountsTracker) getAccountInfo(address []byte) (accountInfo, bool) {
	at.mutAccountInfo.RLock()
	defer at.mutAccountInfo.RUnlock()

	accntInfo, found := at.accountInfo[string(address)]

	return accntInfo, found
}

func (at *accountsTracker) setAccountInfo(address []byte, accntInfo accountInfo) {
	at.mutAccountInfo.Lock()
	defer at.mutAccountInfo.Unlock()

	at.accountInfo[string(address)] = accntInfo
}
