package p2p

import (
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

const interceptedDirectConnectionInfoType = "intercepted direct connection info"

// ArgInterceptedDirectConnectionInfo is the argument used in the intercepted direct connection info constructor
type ArgInterceptedDirectConnectionInfo struct {
	Marshaller  marshal.Marshalizer
	DataBuff    []byte
	NumOfShards uint32
}

// interceptedDirectConnectionInfo is a wrapper over DirectConnectionInfo
type interceptedDirectConnectionInfo struct {
	directConnectionInfo p2p.DirectConnectionInfo
	numOfShards          uint32
}

// NewInterceptedDirectConnectionInfo creates a new intercepted direct connection info instance
func NewInterceptedDirectConnectionInfo(args ArgInterceptedDirectConnectionInfo) (*interceptedDirectConnectionInfo, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	directConnectionInfo, err := createDirectConnectionInfo(args.Marshaller, args.DataBuff)
	if err != nil {
		return nil, err
	}

	return &interceptedDirectConnectionInfo{
		directConnectionInfo: *directConnectionInfo,
		numOfShards:          args.NumOfShards,
	}, nil
}

func checkArgs(args ArgInterceptedDirectConnectionInfo) error {
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

func createDirectConnectionInfo(marshaller marshal.Marshalizer, buff []byte) (*p2p.DirectConnectionInfo, error) {
	directConnectionInfo := &p2p.DirectConnectionInfo{}
	err := marshaller.Unmarshal(directConnectionInfo, buff)
	if err != nil {
		return nil, err
	}

	return directConnectionInfo, nil
}

// CheckValidity checks the validity of the received direct connection info
func (idci *interceptedDirectConnectionInfo) CheckValidity() error {
	shardId, err := strconv.ParseUint(idci.directConnectionInfo.ShardId, 10, 32)
	if err != nil {
		return err
	}
	if uint32(shardId) != common.MetachainShardId &&
		uint32(shardId) >= idci.numOfShards {
		return process.ErrInvalidValue
	}

	return nil
}

// IsForCurrentShard always returns true
func (idci *interceptedDirectConnectionInfo) IsForCurrentShard() bool {
	return true
}

// Hash always returns an empty string
func (idci *interceptedDirectConnectionInfo) Hash() []byte {
	return []byte("")
}

// Type returns the type of this intercepted data
func (idci *interceptedDirectConnectionInfo) Type() string {
	return interceptedDirectConnectionInfoType
}

// Identifiers always returns an array with an empty string
func (idci *interceptedDirectConnectionInfo) Identifiers() [][]byte {
	return [][]byte{make([]byte, 0)}
}

// String returns the most important fields as string
func (idci *interceptedDirectConnectionInfo) String() string {
	return fmt.Sprintf("shard=%s", idci.directConnectionInfo.ShardId)
}

// ShardID returns the shard id
func (idci *interceptedDirectConnectionInfo) ShardID() string {
	return idci.directConnectionInfo.ShardId
}

// IsInterfaceNil returns true if there is no value under the interface
func (idci *interceptedDirectConnectionInfo) IsInterfaceNil() bool {
	return idci == nil
}
