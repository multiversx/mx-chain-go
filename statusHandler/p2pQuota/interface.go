package p2pQuota

// BlacklistHandler defines the behavior of a component that is able to add a blacklisted key (peer)
type BlacklistHandler interface {
	Add(key string) error
	IsInterfaceNil() bool
}
