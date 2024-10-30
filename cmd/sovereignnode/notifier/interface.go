package notifier

type SovereignNotifierBootstrapper interface {
	Start()
	Close() error
}
