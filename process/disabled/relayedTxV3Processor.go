package disabled

import (
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

// IsInterfaceNil returns true if there is no value under the interface
func (proc *relayedTxV3Processor) IsInterfaceNil() bool {
	return proc == nil
}
