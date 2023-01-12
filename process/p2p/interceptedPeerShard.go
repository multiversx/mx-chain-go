package p2p

import (
	"fmt"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
)

const interceptedPeerShardType = "intercepted peer shard"

// ArgInterceptedPeerShard is the argument used in the intercepted peer shard constructor
type ArgInterceptedPeerShard struct {
	Marshaller  marshal.Marshalizer
	DataBuff    []byte
	NumOfShards uint32
}

// interceptedPeerShard is a wrapper over PeerShard message
type interceptedPeerShard struct {
	peerShardMessage factory.PeerShard
	numOfShards      uint32
}

// NewInterceptedPeerShard creates a new intercepted peer shard instance
func NewInterceptedPeerShard(args ArgInterceptedPeerShard) (*interceptedPeerShard, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	peerShard, err := createPeerShardMessage(args.Marshaller, args.DataBuff)
	if err != nil {
		return nil, err
	}

	return &interceptedPeerShard{
		peerShardMessage: *peerShard,
		numOfShards:      args.NumOfShards,
	}, nil
}

func checkArgs(args ArgInterceptedPeerShard) error {
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if len(args.DataBuff) == 0 {
		return process.ErrNilBuffer
	}
	if args.NumOfShards == 0 {
		return process.ErrInvalidValue
	}

	return nil
}

func createPeerShardMessage(marshaller marshal.Marshalizer, buff []byte) (*factory.PeerShard, error) {
	peerShard := &factory.PeerShard{}
	err := marshaller.Unmarshal(peerShard, buff)
	if err != nil {
		return nil, err
	}

	return peerShard, nil
}

// CheckValidity checks the validity of the received peer shard
func (ips *interceptedPeerShard) CheckValidity() error {
	shardId, err := strconv.ParseUint(ips.peerShardMessage.ShardId, 10, 32)
	if err != nil {
		return err
	}
	if uint32(shardId) != common.MetachainShardId &&
		uint32(shardId) >= ips.numOfShards {
		return process.ErrInvalidValue
	}

	return nil
}

// IsForCurrentShard always returns true
func (ips *interceptedPeerShard) IsForCurrentShard() bool {
	return true
}

// Hash always returns an empty string
func (ips *interceptedPeerShard) Hash() []byte {
	return []byte("")
}

// Type returns the type of this intercepted data
func (ips *interceptedPeerShard) Type() string {
	return interceptedPeerShardType
}

// Identifiers always returns an array with an empty string
func (ips *interceptedPeerShard) Identifiers() [][]byte {
	return [][]byte{make([]byte, 0)}
}

// String returns the most important fields as string
func (ips *interceptedPeerShard) String() string {
	return fmt.Sprintf("shard=%s", ips.peerShardMessage.ShardId)
}

// ShardID returns the shard id
func (ips *interceptedPeerShard) ShardID() string {
	return ips.peerShardMessage.ShardId
}

// IsInterfaceNil returns true if there is no value under the interface
func (ips *interceptedPeerShard) IsInterfaceNil() bool {
	return ips == nil
}
