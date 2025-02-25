package chaos

type failType string

const (
	failTypePanic            failType = "panic"
	failTypeReturnError      failType = "returnError"
	failTypeEarlyReturn      failType = "earlyReturn"
	failTypeCorruptVariables failType = "corruptVariables"
	failTypeSleep            failType = "sleep"
)

var knownFailTypes = map[failType]struct{}{
	failTypePanic:            {},
	failTypeReturnError:      {},
	failTypeEarlyReturn:      {},
	failTypeCorruptVariables: {},
	failTypeSleep:            {},
}

var knownPoints = map[string]struct{}{
	"shardBlockCreateBlock":  {},
	"shardBlockProcessBlock": {},
	"metaBlockCreateBlock":   {},
	"metaBlockProcessBlock":  {},
	"epochConfirmed":         {},

	"consensusV1SubroundSignatureDoSignatureJobWhenSingleKey":                      {},
	"consensusV1SubroundSignatureDoSignatureJobWhenMultiKey":                       {},
	"consensusV1SubroundEndRoundCheckSignaturesValidity":                           {},
	"consensusV1SubroundEndRoundDoEndRoundJobByLeaderBeforeBroadcastingFinalBlock": {},

	"consensusV2SubroundBlockDoBlockJob":                      {},
	"consensusV2SubroundSignatureDoSignatureJobWhenSingleKey": {},
	"consensusV2SubroundSignatureDoSignatureJobWhenMultiKey":  {},

	"consensusV2SubroundEndRoundOnReceivedProof":                      {},
	"consensusV2SubroundEndRoundVerifyInvalidSigner":                  {},
	"consensusV2SubroundEndRoundCommitBlock":                          {},
	"consensusV2SubroundEndRoundDoEndRoundJobByNode":                  {},
	"consensusV2SubroundEndRoundPrepareBroadcastBlockData":            {},
	"consensusV2SubroundEndRoundWaitForProof":                         {},
	"consensusV2SubroundEndRoundFinalizeConfirmedBlock":               {},
	"consensusV2SubroundEndRoundSendProof":                            {},
	"consensusV2SubroundEndRoundShouldSendProof":                      {},
	"consensusV2SubroundEndRoundAggregateSigsAndHandleInvalidSigners": {},
	"consensusV2SubroundEndRoundVerifySignature":                      {},
	"consensusV2SubroundEndRoundVerifySignatureBeforeChangeScore":     {},
	"consensusV2SubroundEndRoundCreateAndBroadcastProof":              {},
	"consensusV2SubroundEndRoundCheckSignaturesValidity":              {},
	"consensusV2SubroundEndRoundIsOutOfTime":                          {},
	"consensusV2SubroundEndRoundRemainingTime":                        {},
	"consensusV2SubroundEndRoundOnReceivedSignature":                  {},
	"consensusV2SubroundEndRoundCheckReceivedSignatures":              {},
	"consensusV2SubroundEndRoundGetNumOfSignaturesCollected":          {},
}

type failureExpressionParameterName string

const (
	parameterPoint           failureExpressionParameterName = "point"
	parameterRandomNumber    failureExpressionParameterName = "randomNumber"
	parameterNow             failureExpressionParameterName = "now"
	parameterUptime          failureExpressionParameterName = "uptime"
	parameterNodeDisplayName failureExpressionParameterName = "nodeDisplayName"
	parameterShard           failureExpressionParameterName = "shard"
	parameterEpoch           failureExpressionParameterName = "epoch"
	parameterRound           failureExpressionParameterName = "round"

	parameterNodeIndex           failureExpressionParameterName = "nodeIndex"
	parameterNodePublicKey       failureExpressionParameterName = "nodePublicKey"
	parameterConsensusSize       failureExpressionParameterName = "consensusSize"
	parameterAmILeader           failureExpressionParameterName = "amILeader"
	parameterBlockNonce          failureExpressionParameterName = "blockNonce"
	parameterBlockIsStartOfEpoch failureExpressionParameterName = "blockIsStartOfEpoch"
)

const (
	defaultConfigFilePath = "./config/chaos.json"
)
