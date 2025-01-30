package chaos

type failureName string

const (
	failureProcessTransactionShouldReturnError        failureName = "processTransactionShouldReturnError"
	failureMaybeCorruptSignature                      failureName = "maybeCorruptSignature"
	failureShouldSkipWaitingForSignatures             failureName = "shouldSkipWaitingForSignatures"
	failureShouldReturnErrorInCheckSignaturesValidity failureName = "shouldReturnErrorInCheckSignaturesValidity"
	failureMaybeCorruptLeaderSignature                failureName = "maybeCorruptLeaderSignature"
	failureShouldSkipSendingBlock                     failureName = "shouldSkipSendingBlock"
)

type failureExpressionParameterName string

const (
	parameterRandomNumber failureExpressionParameterName = "randomNumber"
	parameterNow          failureExpressionParameterName = "now"
	parameterShard        failureExpressionParameterName = "shard"
	parameterEpoch        failureExpressionParameterName = "epoch"
	parameterRound        failureExpressionParameterName = "round"

	parameterCounterProcessTransaction failureExpressionParameterName = "counterProcessTransaction"

	parameterBlockNonce              failureExpressionParameterName = "blockNonce"
	parameterNodePublicKeyLastByte   failureExpressionParameterName = "nodePublicKeyLastByte"
	parameterTransactionHashLastByte failureExpressionParameterName = "transactionHashLastByte"
)
