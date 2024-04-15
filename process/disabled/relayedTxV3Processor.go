package disabled

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type relayedTxV3Processor struct {
}

// NewRelayedTxV3Processor returns a new instance of disabled relayedTxV3Processor
func NewRelayedTxV3Processor() *relayedTxV3Processor {
	return &relayedTxV3Processor{}
}

// CheckRelayedTx returns nil as it is disabled
func (proc *relayedTxV3Processor) CheckRelayedTx(_ *transaction.Transaction) error {
	return nil
}

// ComputeRelayedTxFees returns 0, 0 as it is disabled
func (proc *relayedTxV3Processor) ComputeRelayedTxFees(_ *transaction.Transaction) (*big.Int, *big.Int) {
	return big.NewInt(0), big.NewInt(0)
}

// GetUniqueSendersRequiredFeesMap returns an empty map as it is disabled
func (proc *relayedTxV3Processor) GetUniqueSendersRequiredFeesMap(_ []*transaction.Transaction) map[string]*big.Int {
	return make(map[string]*big.Int)
}

// IsInterfaceNil returns true if there is no value under the interface
func (proc *relayedTxV3Processor) IsInterfaceNil() bool {
	return proc == nil
}
