package heartbeat

import "fmt"

const maxSizeInBytes = 128

func verifyLengths(heartbeat *Heartbeat) error {
	err := VerifyHeartbeatProperyLen("Pubkey", heartbeat.Pubkey)
	if err != nil {
		return err
	}

	err = VerifyHeartbeatProperyLen("Payload", heartbeat.Payload)
	if err != nil {
		return err
	}

	err = VerifyHeartbeatProperyLen("NodeDisplayName", []byte(heartbeat.NodeDisplayName))
	if err != nil {
		return err
	}

	err = VerifyHeartbeatProperyLen("VersionNumber", []byte(heartbeat.VersionNumber))
	if err != nil {
		return err
	}

	err = VerifyHeartbeatProperyLen("Signature", heartbeat.Signature)
	if err != nil {
		return err
	}

	return nil
}

// VerifyHeartbeatProperyLen returns an error if the provided value is longer than accepted by the network
func VerifyHeartbeatProperyLen(property string, value []byte) error {
	if len(value) > maxSizeInBytes {
		return fmt.Errorf("%w for %s", ErrPropertyTooLong, property)
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
