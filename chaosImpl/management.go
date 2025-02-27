package chaosImpl

type managementCommand struct {
	Action      string             `json:"action"`
	ProfileName string             `json:"profileName"`
	FailureName string             `json:"failureName"`
	ToggleValue bool               `json:"toggleValue"`
	Failure     *failureDefinition `json:"failure"`
}

type managementCommandAction string

const (
	managementToggleChaos   managementCommandAction = "toggleChaos"
	managementSelectProfile managementCommandAction = "selectProfile"
	managementToggleFailure managementCommandAction = "toggleFailure"
	managementAddFailure    managementCommandAction = "addFailure"
)
