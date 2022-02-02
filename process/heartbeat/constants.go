package heartbeat

const (
	minSizeInBytes                    = 1
	maxSizeInBytes                    = 128
	minDurationInSec                  = 10
	payloadExpiryThresholdInSec       = 10
	interceptedPeerAuthenticationType = "intercepted peer authentication"
	interceptedHeartbeatType          = "intercepted heartbeat"
	publicKeyProperty                 = "public key"
	signatureProperty                 = "signature"
	peerIdProperty                    = "peer id"
	payloadProperty                   = "payload"
	payloadSignatureProperty          = "payload signature"
	versionNumberProperty             = "version number"
	nodeDisplayNameProperty           = "node display name"
	identityProperty                  = "identity"
)
