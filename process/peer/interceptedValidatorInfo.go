package peer

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ArgInterceptedValidatorInfo is the argument used to create a new intercepted validator info
type ArgInterceptedValidatorInfo struct {
	DataBuff    []byte
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher
}

// interceptedValidatorInfo is a wrapper over validatorInfo
type interceptedValidatorInfo struct {
	shardValidatorInfo *state.ShardValidatorInfo
	hash               []byte
}

// NewInterceptedValidatorInfo creates a new intercepted validator info instance
func NewInterceptedValidatorInfo(args ArgInterceptedValidatorInfo) (*interceptedValidatorInfo, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	shardValidatorInfo, err := createShardValidatorInfo(args.Marshalizer, args.DataBuff)
	if err != nil {
		return nil, err
	}

	return &interceptedValidatorInfo{
		shardValidatorInfo: shardValidatorInfo,
		hash:               args.Hasher.Compute(string(args.DataBuff)),
	}, nil
}

func checkArgs(args ArgInterceptedValidatorInfo) error {
	if len(args.DataBuff) == 0 {
		return process.ErrNilBuffer
	}
	if check.IfNil(args.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
	}

	return nil
}

func createShardValidatorInfo(marshalizer marshal.Marshalizer, buff []byte) (*state.ShardValidatorInfo, error) {
	shardValidatorInfo := &state.ShardValidatorInfo{}
	err := marshalizer.Unmarshal(shardValidatorInfo, buff)
	if err != nil {
		return nil, err
	}

	return shardValidatorInfo, nil
}

// CheckValidity checks the validity of the received validator info
func (ivi *interceptedValidatorInfo) CheckValidity() error {
	// Verify string properties len
	err := verifyPropertyLen(publicKeyProperty, ivi.shardValidatorInfo.PublicKey, publicKeyPropertyRequiredBytesLen, minSizeInBytes, maxSizeInBytes)
	if err != nil {
		return err
	}
	err = verifyPropertyLen(listProperty, []byte(ivi.shardValidatorInfo.List), 0, minSizeInBytes, maxSizeInBytes)
	if err != nil {
		return err
	}

	return nil
}

// IsForCurrentShard always returns true
func (ivi *interceptedValidatorInfo) IsForCurrentShard() bool {
	return true
}

// ValidatorInfo returns the current validator info structure
func (ivi *interceptedValidatorInfo) ValidatorInfo() *state.ShardValidatorInfo {
	return ivi.shardValidatorInfo
}

// Hash returns the hash of this validator info
func (ivi *interceptedValidatorInfo) Hash() []byte {
	return ivi.hash
}

// Type returns the type of this intercepted data
func (ivi *interceptedValidatorInfo) Type() string {
	return interceptedValidatorInfoType
}

// Identifiers returns the identifiers used in requests
func (ivi *interceptedValidatorInfo) Identifiers() [][]byte {
	return [][]byte{ivi.hash}
}

// String returns the validator's info most important fields as string
func (ivi *interceptedValidatorInfo) String() string {
	return fmt.Sprintf("pk=%s, shard=%d, list=%s, index=%d, tempRating=%d",
		logger.DisplayByteSlice(ivi.shardValidatorInfo.PublicKey),
		ivi.shardValidatorInfo.ShardId,
		ivi.shardValidatorInfo.List,
		ivi.shardValidatorInfo.Index,
		ivi.shardValidatorInfo.TempRating,
	)
}

// verifyPropertyLen returns an error if the provided value is longer than accepted by the network
func verifyPropertyLen(property string, value []byte, requiredLen, minSize, maxSize int) error {
	hasRequiredLen := requiredLen != 0
	isOverLimit := len(value) > maxSize
	isOverRequiredLen := len(value) > requiredLen
	isTooLong := isOverLimit || (hasRequiredLen && isOverRequiredLen)
	if isTooLong {
		return fmt.Errorf("%w for %s", process.ErrPropertyTooLong, property)
	}

	isUnderLimit := len(value) < minSize
	isUnderRequiredLen := len(value) < requiredLen
	isTooShort := isUnderLimit || (hasRequiredLen && isUnderRequiredLen)
	if isTooShort {
		return fmt.Errorf("%w for %s", process.ErrPropertyTooShort, property)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ivi *interceptedValidatorInfo) IsInterfaceNil() bool {
	return ivi == nil
}
