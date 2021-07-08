package core

import "github.com/denisbrodbeck/machineid"

const maxMachineIDLen = 10

// GetAnonymizedMachineID returns the machine ID anonymized with the provided app ID string
func GetAnonymizedMachineID(appID string) string {
	machineID, err := machineid.ProtectedID(appID)
	if err != nil {
		log.Warn("error fetching machine id", "error", err)
		machineID = "unknown machine ID"
	}
	if len(machineID) > maxMachineIDLen {
		machineID = machineID[:maxMachineIDLen]
	}

	return machineID
}
