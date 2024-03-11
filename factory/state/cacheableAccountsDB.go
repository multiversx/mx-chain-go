package state

import (
	"sync"

	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("CacheableAccountsDB")

type CacheableAccountsDB struct {
	state.AccountsAdapter
	Cache     map[string]map[string]vmcommon.AccountHandler
	mutCaches map[string]*sync.RWMutex
	mutCache  sync.RWMutex
}

func (cadb *CacheableAccountsDB) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	// log.Info("GetExistingAccount")
	cadb.mutCache.RLock()
	currentMap, ok := cadb.Cache[string(address)]
	currentMutex, _ := cadb.mutCaches[string(address)]
	cadb.mutCache.RUnlock()

	if !ok {
		cadb.mutCache.Lock()
		currentMap = make(map[string]vmcommon.AccountHandler)
		cadb.mutCaches[string(address)] = &sync.RWMutex{}
		currentMutex, _ = cadb.mutCaches[string(address)]
		cadb.Cache[string(address)] = currentMap
		cadb.mutCache.Unlock()
	}

	currentMutex.RLock()
	account, ok := currentMap[string(address)]
	currentMutex.RUnlock()
	if ok {
		return account, nil
	}

	account, err := cadb.AccountsAdapter.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	currentMutex.Lock()
	currentMap[string(address)] = account
	currentMutex.Unlock()

	return account, nil
}

func (cadb *CacheableAccountsDB) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	//log.Info("LoadAccount")
	cadb.mutCache.RLock()
	currentMap, ok := cadb.Cache[string(address)]
	currentMutex, _ := cadb.mutCaches[string(address)]
	cadb.mutCache.RUnlock()

	if !ok {
		cadb.mutCache.Lock()
		currentMap = make(map[string]vmcommon.AccountHandler)
		cadb.mutCaches[string(address)] = &sync.RWMutex{}
		cadb.Cache[string(address)] = currentMap
		currentMutex, _ = cadb.mutCaches[string(address)]
		cadb.mutCache.Unlock()
	}

	if ok {
		currentMutex.RLock()
		account, ok := currentMap[string(address)]
		currentMutex.RUnlock()
		//	log.Info("LoadAccount", "account", currentMap, "ok", ok)
		if ok {
			return account, nil
		}

	}

	account, err := cadb.AccountsAdapter.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	currentMutex.Lock()
	currentMap[string(address)] = account
	currentMutex.Unlock()

	return account, nil
}

func (cadb *CacheableAccountsDB) SaveAccount(account vmcommon.AccountHandler) error {
	//	log.Info("SaveAccount")
	userAccount, ok := account.(state.BaseAccountHandler)
	if ok {
		hasCode := len(userAccount.GetCode()) > 0
		if hasCode {
			err := cadb.AccountsAdapter.SaveAccount(account)
			if err != nil {
				return err
			}
		}
	}

	cadb.mutCache.RLock()
	currentMap, ok := cadb.Cache[string(account.AddressBytes())]
	currentMutex, _ := cadb.mutCaches[string(account.AddressBytes())]
	cadb.mutCache.RUnlock()

	if !ok {
		cadb.mutCache.Lock()
		currentMap = make(map[string]vmcommon.AccountHandler)
		cadb.mutCaches[string(account.AddressBytes())] = &sync.RWMutex{}
		cadb.Cache[string(account.AddressBytes())] = currentMap
		currentMutex, _ = cadb.mutCaches[string(account.AddressBytes())]
		cadb.mutCache.Unlock()
	}
	currentMutex.Lock()
	currentMap[string(account.AddressBytes())] = account
	currentMutex.Unlock()

	return nil
}

func (cadb *CacheableAccountsDB) Commit() ([]byte, error) {
	//	log.Info("Commit")
	cadb.mutCache.Lock()
	defer cadb.mutCache.Unlock()

	for _, accMaps := range cadb.Cache {
		for _, account := range accMaps {
			err := cadb.AccountsAdapter.SaveAccount(account)
			if err != nil {
				return nil, err
			}
		}
	}

	cadb.Cache = make(map[string]map[string]vmcommon.AccountHandler)
	cadb.mutCaches = make(map[string]*sync.RWMutex)
	return cadb.AccountsAdapter.Commit()
}
