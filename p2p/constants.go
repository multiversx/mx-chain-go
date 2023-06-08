package p2p

import (
	p2p "github.com/multiversx/mx-chain-p2p-go"
)

// NodeOperation defines the p2p node operation
type NodeOperation = p2p.NodeOperation

// NormalOperation defines the normal mode operation: either seeder, observer or validator
const NormalOperation = p2p.NormalOperation

// FullArchiveMode defines the node operation as a full archive mode
const FullArchiveMode = p2p.FullArchiveMode

// ListsSharder is the variant that uses lists
const ListsSharder = p2p.ListsSharder

// NilListSharder is the variant that will not do connection trimming
const NilListSharder = p2p.NilListSharder

// ConnectionWatcherTypePrint - new connection found will be printed in the log file
const ConnectionWatcherTypePrint = p2p.ConnectionWatcherTypePrint

// ListenLocalhostAddrWithIp4AndTcp defines the local host listening ip v.4 address and TCP
const ListenLocalhostAddrWithIp4AndTcp = "/ip4/127.0.0.1/tcp/%d"
