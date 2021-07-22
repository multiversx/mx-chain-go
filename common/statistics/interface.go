package statistics

// SoftwareVersionChecker holds the actions needed to be handled by a components which will check the software version
type SoftwareVersionChecker interface {
	StartCheckSoftwareVersion()
	IsInterfaceNil() bool
	Close() error
}

// ResourceMonitorHandler defines the resource monitor supported actions
type ResourceMonitorHandler interface {
	GenerateStatistics() []interface{}
	SaveStatistics()
	StartMonitoring()
	Close() error
	IsInterfaceNil() bool
}
