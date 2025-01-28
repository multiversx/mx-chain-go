package common

import "github.com/multiversx/mx-chain-core-go/data"

// IsValidRelayedTxV3 returns true if the provided transaction is a valid transaction of type relayed v3
func IsValidRelayedTxV3(tx data.TransactionHandler) bool {
	relayedTx, isRelayedV3 := tx.(data.RelayedTransactionHandler)
	if !isRelayedV3 {
		return false
	}
	hasValidRelayer := len(relayedTx.GetRelayerAddr()) == len(tx.GetSndAddr()) && len(relayedTx.GetRelayerAddr()) > 0
	hasValidRelayerSignature := len(relayedTx.GetRelayerSignature()) == len(relayedTx.GetSignature()) && len(relayedTx.GetRelayerSignature()) > 0
	return hasValidRelayer && hasValidRelayerSignature
}

// IsRelayedTxV3 returns true if the provided transaction is a transaction of type relayed v3, without any further checks
func IsRelayedTxV3(tx data.TransactionHandler) bool {
	relayedTx, isRelayedV3 := tx.(data.RelayedTransactionHandler)
	if !isRelayedV3 {
		return false
	}

	hasRelayer := len(relayedTx.GetRelayerAddr()) > 0
	hasRelayerSignature := len(relayedTx.GetRelayerSignature()) > 0
	return hasRelayer || hasRelayerSignature
}
