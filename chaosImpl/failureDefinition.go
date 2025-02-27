package chaosImpl

type failureDefinition struct {
	Name       string                 `json:"name"`
	Enabled    bool                   `json:"enabled"`
	Type       string                 `json:"type"`
	OnPoints   []string               `json:"onPoints"`
	Triggers   []string               `json:"triggers"`
	Parameters map[string]interface{} `json:"parameters"`
}

func (failure *failureDefinition) getParameterAsFloat64(parameterName string) float64 {
	value, ok := failure.Parameters[parameterName]
	if !ok {
		return 0
	}

	floatValue, ok := value.(float64)
	if !ok {
		return 0
	}

	return floatValue
}

func (failure *failureDefinition) getParameterAsBoolean(parameterName string) bool {
	value, ok := failure.Parameters[parameterName]
	if !ok {
		return false
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false
	}

	return boolValue
}

func (failure *failureDefinition) isOnPoint(pointName string) bool {
	for _, point := range failure.OnPoints {
		if point == pointName {
			return true
		}
	}

	return false
}
