package p2p

import (
	"github.com/multiversx/mx-chain-communication-go/p2p"
)

// NodeOperation defines the p2p node operation
type NodeOperation = p2p.NodeOperation

// NormalOperation defines the normal mode operation: either seeder, observer or validator
const NormalOperation = p2p.NormalOperation

// FullArchiveMode defines the node operation as a full archive mode
const FullArchiveMode = p2p.FullArchiveMode

// NetworkType defines the type of the network a messenger is running on
type NetworkType = p2p.NetworkType

// MainNetwork defines the main network
const MainNetwork NetworkType = "main"

// FullArchiveNetwork defines the full archive network
const FullArchiveNetwork NetworkType = "full archive"

// ListsSharder is the variant that uses lists
const ListsSharder = p2p.ListsSharder

// NilListSharder is the variant that will not do connection trimming
const NilListSharder = p2p.NilListSharder

// ConnectionWatcherTypePrint - new connection found will be printed in the log file
const ConnectionWatcherTypePrint = p2p.ConnectionWatcherTypePrint

// ListenAddrWithIp4AndTcp defines the listening address with ip v.4 and TCP
const ListenAddrWithIp4AndTcp = "/ip4/0.0.0.0/tcp/"

// ListenLocalhostAddrWithIp4AndTcp defines the local host listening ip v.4 address and TCP
const ListenLocalhostAddrWithIp4AndTcp = "/ip4/127.0.0.1/tcp/"

// BroadcastMethod defines the broadcast method of the message
type BroadcastMethod = p2p.BroadcastMethod
