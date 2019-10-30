package heartbeat

const maxSizeInBytes = 128

func verifyLengths(heartbeat *Heartbeat) error {
	if len(heartbeat.Pubkey) > maxSizeInBytes ||
		len(heartbeat.Payload) > maxSizeInBytes ||
		len(heartbeat.NodeDisplayName) > maxSizeInBytes ||
		len(heartbeat.VersionNumber) > maxSizeInBytes ||
		len(heartbeat.Signature) > maxSizeInBytes {
		return ErrPropertyTooLong
	}
	return nil
}

func trimLengths(heartbeat *Heartbeat) {
	if len(heartbeat.Payload) > maxSizeInBytes {
		heartbeat.Payload = heartbeat.Payload[:maxSizeInBytes]
	}

	if len(heartbeat.NodeDisplayName) > maxSizeInBytes {
		heartbeat.NodeDisplayName = heartbeat.NodeDisplayName[:maxSizeInBytes]
	}

	if len(heartbeat.VersionNumber) > maxSizeInBytes {
		heartbeat.VersionNumber = heartbeat.VersionNumber[:maxSizeInBytes]
	}
}
