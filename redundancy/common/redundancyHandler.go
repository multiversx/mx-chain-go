package common

import (
	"errors"
	"fmt"
)

const minRoundsOfInactivity = 2 // the system does not work as expected with the value of 1
const roundsOfInactivityForMainMachine = 0

var errInvalidValue = errors.New("invalid value")

type redundancyHandler struct {
	roundsOfInactivity int
}

// NewRedundancyHandler creates an instance of type redundancyHandler that is able to manage the current counter of
// rounds without inactivity. Not a concurrent safe implementation.
func NewRedundancyHandler() *redundancyHandler {
	return &redundancyHandler{}
}

// CheckMaxRoundsOfInactivity will check the provided max rounds of inactivity value and return an error if it is not correct
func CheckMaxRoundsOfInactivity(maxRoundsOfInactivity int) error {
	if maxRoundsOfInactivity == roundsOfInactivityForMainMachine {
		return nil
	}
	if maxRoundsOfInactivity < minRoundsOfInactivity {
		return fmt.Errorf("%w for maxRoundsOfInactivity, minimum %d (or 0), got %d",
			errInvalidValue, minRoundsOfInactivity, maxRoundsOfInactivity)
	}

	return nil
}

// IsMainNode returns true if the provided maxRoundsOfInactivity value is equal with the
// roundsOfInactivityForMainMachine constant (0)
func IsMainNode(maxRoundsOfInactivity int) bool {
	return maxRoundsOfInactivity == roundsOfInactivityForMainMachine
}

// IncrementRoundsOfInactivity will increment the rounds of inactivity
func (handler *redundancyHandler) IncrementRoundsOfInactivity() {
	handler.roundsOfInactivity++
}

// ResetRoundsOfInactivity will reset the rounds of inactivity
func (handler *redundancyHandler) ResetRoundsOfInactivity() {
	handler.roundsOfInactivity = 0
}

// IsMainMachineActive returns true if the main machine is still active in case of a backup or the machine is the
// main machine
func (handler *redundancyHandler) IsMainMachineActive(maxRoundsOfInactivity int) bool {
	if IsMainNode(maxRoundsOfInactivity) {
		return true
	}

	return handler.mainMachineIsActive(maxRoundsOfInactivity)
}

func (handler *redundancyHandler) mainMachineIsActive(maxRoundsOfInactivity int) bool {
	return handler.roundsOfInactivity <= maxRoundsOfInactivity
}

// ShouldActAsValidator returns true if either the machine is a main machine or the machine is a backup machine but the
// main machine failed
func (handler *redundancyHandler) ShouldActAsValidator(maxRoundsOfInactivity int) bool {
	if IsMainNode(maxRoundsOfInactivity) {
		return true
	}

	return !handler.mainMachineIsActive(maxRoundsOfInactivity)
}

// RoundsOfInactivity returns the inner roundsOfInactivity value
func (handler *redundancyHandler) RoundsOfInactivity() int {
	return handler.roundsOfInactivity
}
