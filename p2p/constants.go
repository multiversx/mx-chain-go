package p2p

import (
	"github.com/multiversx/mx-chain-communication-go/p2p"
)

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

// LocalHostListenAddrWithIp4AndTcp defines the local host listening ip v.4 address and TCP
const LocalHostListenAddrWithIp4AndTcp = p2p.LocalHostListenAddrWithIp4AndTcp

// BroadcastMethod defines the broadcast method of the message
type BroadcastMethod = p2p.BroadcastMethod
