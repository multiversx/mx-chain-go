//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. validatorInfo.proto

package state

import math_big "math/big"

// IsInterfaceNil returns true if there is no value under the interface
func (vi *ValidatorInfo) IsInterfaceNil() bool {
	return vi == nil
}

func (vi *ValidatorInfo) SetPublicKey(publicKey []byte) {
	vi.PublicKey = publicKey
}

func (vi *ValidatorInfo) SetList(list string) {
	vi.List = list
}

func (vi *ValidatorInfo) SetShardId(shardID uint32) {
	vi.ShardId = shardID
}

func (vi *ValidatorInfo) SetIndex(index uint32) {
	vi.Index = index
}

func (vi *ValidatorInfo) SetTempRating(tempRating uint32) {
	vi.TempRating = tempRating
}

func (vi *ValidatorInfo) SetRating(rating uint32) {
	vi.Rating = rating
}

func (vi *ValidatorInfo) SetRatingModifier(ratingModifier float32) {
	vi.RatingModifier = ratingModifier
}

func (vi *ValidatorInfo) SetRewardAddress(rewardAddress []byte) {
	vi.RewardAddress = rewardAddress
}

func (vi *ValidatorInfo) SetLeaderSuccess(leaderSuccess uint32) {
	vi.LeaderSuccess = leaderSuccess
}

func (vi *ValidatorInfo) SetLeaderFailure(leaderFailure uint32) {
	vi.LeaderFailure = leaderFailure
}

func (vi *ValidatorInfo) SetValidatorSuccess(validatorSuccess uint32) {
	vi.ValidatorSuccess = validatorSuccess
}

func (vi *ValidatorInfo) SetValidatorFailure(validatorFailure uint32) {
	vi.ValidatorFailure = validatorFailure
}

func (vi *ValidatorInfo) SetValidatorIgnoredSignatures(validatorIgnoredSignatures uint32) {
	vi.ValidatorIgnoredSignatures = validatorIgnoredSignatures
}

func (vi *ValidatorInfo) SetNumSelectedInSuccessBlocks(numSelectedInSuccessBlock uint32) {
	vi.NumSelectedInSuccessBlocks = numSelectedInSuccessBlock
}

func (vi *ValidatorInfo) SetAccumulatedFees(accumulatedFees *math_big.Int) {
	vi.AccumulatedFees = math_big.NewInt(0).Set(accumulatedFees)
}

func (vi *ValidatorInfo) SetTotalLeaderSuccess(totalLeaderSuccess uint32) {
	vi.TotalLeaderSuccess = totalLeaderSuccess
}

func (vi *ValidatorInfo) SetTotalLeaderFailure(totalLeaderFailure uint32) {
	vi.TotalLeaderFailure = totalLeaderFailure
}

func (vi *ValidatorInfo) SetTotalValidatorSuccess(totalValidatorSuccess uint32) {
	vi.TotalValidatorSuccess = totalValidatorSuccess
}

func (vi *ValidatorInfo) SetTotalValidatorFailure(totalValidatorFailure uint32) {
	vi.TotalValidatorFailure = totalValidatorFailure
}

func (vi *ValidatorInfo) SetTotalValidatorIgnoredSignatures(totalValidatorIgnoredSignatures uint32) {
	vi.TotalValidatorIgnoredSignatures = totalValidatorIgnoredSignatures
}

// IsInterfaceNil returns true if there is no value under the interface
func (svi *ShardValidatorInfo) IsInterfaceNil() bool {
	return svi == nil
}
