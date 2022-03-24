package p2p

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/process"
)

const interceptedValidatorInfoType = "intercepted validator info"

// ArgInterceptedValidatorInfo is the argument used in the intercepted validator info constructor
type ArgInterceptedValidatorInfo struct {
	Marshaller  marshal.Marshalizer
	DataBuff    []byte
	NumOfShards uint32
}

// interceptedValidatorInfo is a wrapper over ShardValidatorInfo
type interceptedValidatorInfo struct {
	shardValidatorInfo message.ShardValidatorInfo
	numOfShards        uint32
}

// NewInterceptedValidatorInfo creates a new intercepted validator info instance
func NewInterceptedValidatorInfo(args ArgInterceptedValidatorInfo) (*interceptedValidatorInfo, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	shardValidatorInfo, err := createShardValidatorInfo(args.Marshaller, args.DataBuff)
	if err != nil {
		return nil, err
	}

	return &interceptedValidatorInfo{
		shardValidatorInfo: *shardValidatorInfo,
		numOfShards:        args.NumOfShards,
	}, nil
}

func checkArgs(args ArgInterceptedValidatorInfo) error {
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
func (isvi *interceptedValidatorInfo) CheckValidity() error {
	if isvi.shardValidatorInfo.ShardId != common.MetachainShardId &&
		isvi.shardValidatorInfo.ShardId >= isvi.numOfShards {
		return process.ErrInvalidValue
	}

	return nil
}

// IsForCurrentShard always returns true
func (isvi *interceptedValidatorInfo) IsForCurrentShard() bool {
	return true
}

// Hash always returns an empty string
func (isvi *interceptedValidatorInfo) Hash() []byte {
	return []byte("")
}

// Type returns the type of this intercepted data
func (isvi *interceptedValidatorInfo) Type() string {
	return interceptedValidatorInfoType
}

// Identifiers always returns an array with an empty string
func (isvi *interceptedValidatorInfo) Identifiers() [][]byte {
	return [][]byte{make([]byte, 0)}
}

// String returns the most important fields as string
func (isvi *interceptedValidatorInfo) String() string {
	return fmt.Sprintf("shard=%d", isvi.shardValidatorInfo.ShardId)
}

// ShardID returns the shard id
func (isvi *interceptedValidatorInfo) ShardID() uint32 {
	return isvi.shardValidatorInfo.ShardId
}

// IsInterfaceNil returns true if there is no value under the interface
func (isvi *interceptedValidatorInfo) IsInterfaceNil() bool {
	return isvi == nil
}
