package peer

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
)

// ArgInterceptedValidatorInfo is the argument used to create a new intercepted validator info
type ArgInterceptedValidatorInfo struct {
	DataBuff         []byte
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	NodesCoordinator process.NodesCoordinator
}

// interceptedValidatorInfo is a wrapper over validatorInfo
type interceptedValidatorInfo struct {
	validatorInfo    state.ValidatorInfo
	nodesCoordinator process.NodesCoordinator
	hash             []byte
}

// NewInterceptedValidatorInfo creates a new intercepted validator info instance
func NewInterceptedValidatorInfo(args ArgInterceptedValidatorInfo) (*interceptedValidatorInfo, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	validatorInfo, err := createValidatorInfo(args.Marshalizer, args.DataBuff)
	if err != nil {
		return nil, err
	}

	return &interceptedValidatorInfo{
		validatorInfo:    *validatorInfo,
		nodesCoordinator: args.NodesCoordinator,
		hash:             args.Hasher.Compute(string(args.DataBuff)),
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
	if check.IfNil(args.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}

	return nil
}

func createValidatorInfo(marshalizer marshal.Marshalizer, buff []byte) (*state.ValidatorInfo, error) {
	validatorInfo := &state.ValidatorInfo{}
	err := marshalizer.Unmarshal(validatorInfo, buff)
	if err != nil {
		return nil, err
	}

	return validatorInfo, nil
}

// CheckValidity checks the validity of the received validator info
func (ivi *interceptedValidatorInfo) CheckValidity() error {
	// Verify string properties len
	err := verifyPropertyLen(publicKeyProperty, ivi.validatorInfo.PublicKey, publicKeyPropertyRequiredBytesLen, minSizeInBytes, maxSizeInBytes)
	if err != nil {
		return err
	}
	err = verifyPropertyLen(listProperty, []byte(ivi.validatorInfo.List), 0, minSizeInBytes, maxSizeInBytes)
	if err != nil {
		return err
	}
	err = verifyPropertyLen(rewardAddressProperty, ivi.validatorInfo.RewardAddress, 0, minSizeInBytes, rewardAddressPropertyMaxPropertyBytesLen)
	if err != nil {
		return err
	}

	// Check if the public key is a validator
	_, _, err = ivi.nodesCoordinator.GetValidatorWithPublicKey(ivi.validatorInfo.PublicKey)
	return err
}

// IsForCurrentShard always returns true
func (ivi *interceptedValidatorInfo) IsForCurrentShard() bool {
	return true
}

// ValidatorInfo returns the current validator info structure
func (ivi *interceptedValidatorInfo) ValidatorInfo() state.ValidatorInfo {
	return ivi.validatorInfo
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
	return fmt.Sprintf("pk=%s, shard=%d, list=%s, index=%d, tempRating=%d, rating=%d",
		logger.DisplayByteSlice(ivi.validatorInfo.PublicKey),
		ivi.validatorInfo.ShardId,
		ivi.validatorInfo.List,
		ivi.validatorInfo.Index,
		ivi.validatorInfo.TempRating,
		ivi.validatorInfo.Rating,
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
