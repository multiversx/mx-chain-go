package p2p

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/process"
)

const interceptedShardValidatorInfoType = "intercepted shard validator info"

// ArgInterceptedShardValidatorInfo is the argument used in the intercepted shard validator info constructor
type ArgInterceptedShardValidatorInfo struct {
	Marshaller  marshal.Marshalizer
	DataBuff    []byte
	NumOfShards uint32
}

// interceptedShardValidatorInfo is a wrapper over ShardValidatorInfo
type interceptedShardValidatorInfo struct {
	shardValidatorInfo message.ShardValidatorInfo
	numOfShards        uint32
}

// NewInterceptedShardValidatorInfo creates a new intercepted shard validator info instance
func NewInterceptedShardValidatorInfo(args ArgInterceptedShardValidatorInfo) (*interceptedShardValidatorInfo, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	shardValidatorInfo, err := createShardValidatorInfo(args.Marshaller, args.DataBuff)
	if err != nil {
		return nil, err
	}

	return &interceptedShardValidatorInfo{
		shardValidatorInfo: *shardValidatorInfo,
		numOfShards:        args.NumOfShards,
	}, nil
}

func checkArgs(args ArgInterceptedShardValidatorInfo) error {
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

func createShardValidatorInfo(marshaller marshal.Marshalizer, buff []byte) (*message.ShardValidatorInfo, error) {
	shardValidatorInfo := &message.ShardValidatorInfo{}
	err := marshaller.Unmarshal(shardValidatorInfo, buff)
	if err != nil {
		return nil, err
	}

	return shardValidatorInfo, nil
}

// CheckValidity checks the validity of the received shard validator info
func (isvi *interceptedShardValidatorInfo) CheckValidity() error {
	if isvi.shardValidatorInfo.ShardId != common.MetachainShardId &&
		isvi.shardValidatorInfo.ShardId >= isvi.numOfShards {
		return process.ErrInvalidValue
	}

	return nil
}

// IsForCurrentShard always returns true
func (isvi *interceptedShardValidatorInfo) IsForCurrentShard() bool {
	return true
}

// Hash always returns an empty string
func (isvi *interceptedShardValidatorInfo) Hash() []byte {
	return []byte("")
}

// Type returns the type of this intercepted data
func (isvi *interceptedShardValidatorInfo) Type() string {
	return interceptedShardValidatorInfoType
}

// Identifiers always returns an array with an empty string
func (isvi *interceptedShardValidatorInfo) Identifiers() [][]byte {
	return [][]byte{make([]byte, 0)}
}

// String returns the most important fields as string
func (isvi *interceptedShardValidatorInfo) String() string {
	return fmt.Sprintf("shard=%d", isvi.shardValidatorInfo.ShardId)
}

// ShardID returns the shard id
func (isvi *interceptedShardValidatorInfo) ShardID() uint32 {
	return isvi.shardValidatorInfo.ShardId
}

// IsInterfaceNil returns true if there is no value under the interface
func (isvi *interceptedShardValidatorInfo) IsInterfaceNil() bool {
	return isvi == nil
}
