package notifier

// SovereignNotifierBootstrapper defines a sovereign notifier bootstrapper
type SovereignNotifierBootstrapper interface {
	Start()
	Close() error
	IsInterfaceNil() bool
}
