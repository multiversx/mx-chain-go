package chaosImpl

import "fmt"

type chaosProfile struct {
	Name           string               `json:"name"`
	Failures       []*failureDefinition `json:"failures"`
	failuresByName map[string]*failureDefinition
}

func (profile *chaosProfile) verify() error {
	for _, failure := range profile.Failures {
		name := failure.Name
		failType := failType(failure.Type)

		if len(name) == 0 {
			return fmt.Errorf("all failures must have a name")
		}

		if len(failType) == 0 {
			return fmt.Errorf("failure '%s' has no fail type", name)
		}

		if _, ok := knownFailTypes[failType]; !ok {
			return fmt.Errorf("failure '%s' has unknown fail type: '%s'", name, failType)
		}

		if len(failure.OnPoints) == 0 {
			return fmt.Errorf("failure '%s' has no points configured", name)
		}

		for _, point := range failure.OnPoints {
			if _, ok := knownPoints[point]; !ok {
				return fmt.Errorf("failure '%s' has unknown activation point: '%s'", name, point)
			}
		}

		if len(failure.Triggers) == 0 {
			return fmt.Errorf("failure '%s' has no triggers configured", name)
		}

		if failType == failTypeSleep {
			if failure.getParameterAsFloat64("duration") == 0 {
				return fmt.Errorf("failure '%s', with fail type '%s', requires the parameter 'duration'", name, failType)
			}
		}
	}

	return nil
}

func (profile *chaosProfile) populateFailuresByName() {
	profile.failuresByName = make(map[string]*failureDefinition)

	for _, failure := range profile.Failures {
		profile.failuresByName[failure.Name] = failure
	}
}

func (profile *chaosProfile) toggleFailure(failureName string, enabled bool) error {
	failure, err := profile.getFailureByName(failureName)
	if err != nil {
		return err
	}

	failure.Enabled = enabled
	return nil
}

func (profile *chaosProfile) getFailureByName(name string) (*failureDefinition, error) {
	failure, ok := profile.failuresByName[name]
	if !ok {
		return nil, fmt.Errorf("failure not found: %s", name)
	}

	return failure, nil
}

func (profile *chaosProfile) addFailure(failure *failureDefinition) error {
	_, exists := profile.failuresByName[failure.Name]
	if exists {
		return fmt.Errorf("failure already exists: %s", failure.Name)
	}

	profile.Failures = append(profile.Failures, failure)
	profile.failuresByName[failure.Name] = failure

	return nil
}

func (profile *chaosProfile) getFailureParameterAsFloat64(failureName string, parameterName string) float64 {
	failure, err := profile.getFailureByName(failureName)
	if err != nil {
		return 0
	}

	return failure.getParameterAsFloat64(parameterName)
}

func (profile *chaosProfile) getFailureParameterAsBoolean(failureName string, parameterName string) bool {
	failure, err := profile.getFailureByName(failureName)
	if err != nil {
		return false
	}

	return failure.getParameterAsBoolean(parameterName)
}

func (profile *chaosProfile) getFailuresOnPoint(point string) []*failureDefinition {
	var failures []*failureDefinition

	for _, failure := range profile.Failures {
		if failure.Enabled && failure.isOnPoint(point) {
			failures = append(failures, failure)
		}
	}

	return failures
}
