package chaos

type failureName string

const (
	failureProcessTransactionShouldReturnError        failureName = "processTransactionShouldReturnError"
	failureShouldCorruptSignature                     failureName = "shouldCorruptSignature"
	failureShouldSkipWaitingForSignatures             failureName = "shouldSkipWaitingForSignatures"
	failureShouldReturnErrorInCheckSignaturesValidity failureName = "shouldReturnErrorInCheckSignaturesValidity"
	failureShouldCorruptLeaderSignature               failureName = "shouldCorruptLeaderSignature"
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

	parameterCounterProcessTransaction failureExpressionParameterName = "counterProcessTransaction"

	parameterBlockNonce              failureExpressionParameterName = "blockNonce"
	parameterNodePublicKeyLastByte   failureExpressionParameterName = "nodePublicKeyLastByte"
	parameterTransactionHashLastByte failureExpressionParameterName = "transactionHashLastByte"
)
