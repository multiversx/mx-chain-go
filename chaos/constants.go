package chaos

type failureName string

const (
	failureProcessTransactionShouldReturnError        failureName = "processTransactionShouldReturnError"
	failureShouldCorruptSignature                     failureName = "shouldCorruptSignature"
	failureShouldSkipWaitingForSignatures             failureName = "shouldSkipWaitingForSignatures"
	failureShouldReturnErrorInCheckSignaturesValidity failureName = "shouldReturnErrorInCheckSignaturesValidity"
	failureShouldDelayBroadcastingFinalBlockAsLeader  failureName = "shouldDelayBroadcastingFinalBlockAsLeader"
	failureShouldCorruptLeaderSignature               failureName = "shouldCorruptLeaderSignature"
	failureShouldDelayLeaderSignature                 failureName = "shouldDelayLeaderSignature"
	failureShouldSkipSendingBlock                     failureName = "shouldSkipSendingBlock"
)

type failureExpressionParameterName string

const (
	parameterRandomNumber    failureExpressionParameterName = "randomNumber"
	parameterNow             failureExpressionParameterName = "now"
	parameterNodeDisplayName failureExpressionParameterName = "nodeDisplayName"
	parameterShard           failureExpressionParameterName = "shard"
	parameterEpoch           failureExpressionParameterName = "epoch"
	parameterRound           failureExpressionParameterName = "round"

	parameterNodeIndex     failureExpressionParameterName = "nodeIndex"
	parameterNodePublicKey failureExpressionParameterName = "nodePublicKey"
	parameterAmILeader     failureExpressionParameterName = "amILeader"
	parameterBlockNonce    failureExpressionParameterName = "blockNonce"
)
