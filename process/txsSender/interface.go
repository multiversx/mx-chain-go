package txsSender

import (
	"io"

	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

// NetworkMessenger defines the basic functionality of a network messenger
// to broadcast buffer data on a channel, for a given topic
type NetworkMessenger interface {
	io.Closer
	// BroadcastOnChannelBlocking asynchronously waits until it can send a
	// message on the channel, but once it is able to, it synchronously sends the
	// message, blocking until sending is completed.
	BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error
	// IsInterfaceNil checks if the underlying pointer is nil
	IsInterfaceNil() bool
}

// TxsSender handles transactions sending
type TxsSender interface {
	io.Closer
	SendBulkTransactions(txs []*transaction.Transaction) (uint64, error)
	// IsInterfaceNil checks if the underlying pointer is nil
	IsInterfaceNil() bool
}
