//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. validatorInfo.proto

package state

import mathbig "math/big"

// IsInterfaceNil returns true if there is no value under the interface
func (vi *ValidatorInfo) IsInterfaceNil() bool {
	return vi == nil
}

// SetPublicKey sets validator's public key
func (vi *ValidatorInfo) SetPublicKey(publicKey []byte) {
	vi.PublicKey = publicKey
}

// SetList sets validator's list
func (vi *ValidatorInfo) SetList(list string) {
	vi.List = list
}

// SetPreviousList sets validator's previous list
func (vi *ValidatorInfo) SetPreviousList(list string) {
	vi.PreviousList = list
}

func (vi *ValidatorInfo) SetListAndIndex(list string, index uint32, updatePreviousValues bool) {
	if updatePreviousValues {
		vi.PreviousList = vi.List
		vi.PreviousIndex = vi.Index
	}

	vi.List = list
	vi.Index = index
}

// SetShardId sets validator's public shard id
func (vi *ValidatorInfo) SetShardId(shardID uint32) {
	vi.ShardId = shardID
}

// SetIndex sets validator's index
func (vi *ValidatorInfo) SetIndex(index uint32) {
	vi.Index = index
}

// SetTempRating sets validator's temp rating
func (vi *ValidatorInfo) SetTempRating(tempRating uint32) {
	vi.TempRating = tempRating
}

// SetRating sets validator's rating
func (vi *ValidatorInfo) SetRating(rating uint32) {
	vi.Rating = rating
}

// SetRatingModifier sets validator's rating modifier
func (vi *ValidatorInfo) SetRatingModifier(ratingModifier float32) {
	vi.RatingModifier = ratingModifier
}

// SetRewardAddress sets validator's reward address
func (vi *ValidatorInfo) SetRewardAddress(rewardAddress []byte) {
	vi.RewardAddress = rewardAddress
}

// SetLeaderSuccess sets leader success
func (vi *ValidatorInfo) SetLeaderSuccess(leaderSuccess uint32) {
	vi.LeaderSuccess = leaderSuccess
}

// SetLeaderFailure sets validator's leader failure
func (vi *ValidatorInfo) SetLeaderFailure(leaderFailure uint32) {
	vi.LeaderFailure = leaderFailure
}

// SetValidatorSuccess sets validator's success
func (vi *ValidatorInfo) SetValidatorSuccess(validatorSuccess uint32) {
	vi.ValidatorSuccess = validatorSuccess
}

// SetValidatorFailure sets validator's failure
func (vi *ValidatorInfo) SetValidatorFailure(validatorFailure uint32) {
	vi.ValidatorFailure = validatorFailure
}

// SetValidatorIgnoredSignatures sets validator's ignored signatures
func (vi *ValidatorInfo) SetValidatorIgnoredSignatures(validatorIgnoredSignatures uint32) {
	vi.ValidatorIgnoredSignatures = validatorIgnoredSignatures
}

// SetNumSelectedInSuccessBlocks sets validator's num of selected in success block
func (vi *ValidatorInfo) SetNumSelectedInSuccessBlocks(numSelectedInSuccessBlock uint32) {
	vi.NumSelectedInSuccessBlocks = numSelectedInSuccessBlock
}

// SetAccumulatedFees sets validator's accumulated fees
func (vi *ValidatorInfo) SetAccumulatedFees(accumulatedFees *mathbig.Int) {
	vi.AccumulatedFees = mathbig.NewInt(0).Set(accumulatedFees)
}

// SetTotalLeaderSuccess sets validator's total leader success
func (vi *ValidatorInfo) SetTotalLeaderSuccess(totalLeaderSuccess uint32) {
	vi.TotalLeaderSuccess = totalLeaderSuccess
}

// SetTotalLeaderFailure sets validator's total leader failure
func (vi *ValidatorInfo) SetTotalLeaderFailure(totalLeaderFailure uint32) {
	vi.TotalLeaderFailure = totalLeaderFailure
}

// SetTotalValidatorSuccess sets validator's total success
func (vi *ValidatorInfo) SetTotalValidatorSuccess(totalValidatorSuccess uint32) {
	vi.TotalValidatorSuccess = totalValidatorSuccess
}

// SetTotalValidatorFailure sets validator's total failure
func (vi *ValidatorInfo) SetTotalValidatorFailure(totalValidatorFailure uint32) {
	vi.TotalValidatorFailure = totalValidatorFailure
}

// SetTotalValidatorIgnoredSignatures sets validator's total ignored signatures
func (vi *ValidatorInfo) SetTotalValidatorIgnoredSignatures(totalValidatorIgnoredSignatures uint32) {
	vi.TotalValidatorIgnoredSignatures = totalValidatorIgnoredSignatures
}

// ShallowClone returns a clone of the object
func (vi *ValidatorInfo) ShallowClone() ValidatorInfoHandler {
	if vi == nil {
		return nil
	}

	validatorCopy := *vi
	return &validatorCopy
}

// IsInterfaceNil returns true if there is no value under the interface
func (svi *ShardValidatorInfo) IsInterfaceNil() bool {
	return svi == nil
}
