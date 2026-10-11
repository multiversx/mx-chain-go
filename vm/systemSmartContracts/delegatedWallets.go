package systemSmartContracts

import "github.com/multiversx/mx-chain-go/vm"

const delegatedWalletUserKeyPrefix = "delegatedWalletUser"

var delegatedWalletUserKeyPrefixBytes = []byte(delegatedWalletUserKeyPrefix)

func getDelegatedWalletForUserFromManager(
	eei vm.SystemEI,
	managerAddress []byte,
	userAddress []byte,
) []byte {
	if len(userAddress) == 0 {
		return nil
	}

	key := buildDelegatedWalletKey(delegatedWalletUserKeyPrefixBytes, userAddress)
	data := eei.GetStorageFromAddress(managerAddress, key)

	return cloneBytesIfNotEmpty(data)
}

func setDelegatedWalletInManager(
	eei vm.SystemEI,
	managerAddress []byte,
	userAddress []byte,
	delegatedWallet []byte,
) {
	userKey := buildDelegatedWalletKey(delegatedWalletUserKeyPrefixBytes, userAddress)

	eei.SetStorageForAddress(managerAddress, userKey, delegatedWallet)
}

func clearDelegatedWalletFromManager(
	eei vm.SystemEI,
	managerAddress []byte,
	userAddress []byte,
) {
	userKey := buildDelegatedWalletKey(delegatedWalletUserKeyPrefixBytes, userAddress)
	delegatedWallet := eei.GetStorageFromAddress(managerAddress, userKey)
	eei.SetStorageForAddress(managerAddress, userKey, []byte{})

	if len(delegatedWallet) == 0 {
		return
	}
}

func buildDelegatedWalletKey(prefix []byte, parts ...[]byte) []byte {
	totalLen := len(prefix)
	for _, part := range parts {
		totalLen += len(part)
	}
	key := make([]byte, 0, totalLen)
	key = append(key, prefix...)
	for _, part := range parts {
		key = append(key, part...)
	}

	return key
}

func cloneBytesIfNotEmpty(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}

	cloned := make([]byte, len(value))
	copy(cloned, value)

	return cloned
}
