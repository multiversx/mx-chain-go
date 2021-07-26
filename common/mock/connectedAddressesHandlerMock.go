package mock

// ConnectedAddressesMock represents a mock implementation of the ConnectedAddresses
type ConnectedAddressesMock struct {
}

// ConnectedAddresses returns an empty slice of string
func (cam *ConnectedAddressesMock) ConnectedAddresses() []string {
	return []string{}
}
