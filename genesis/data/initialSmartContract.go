package data

// InitialSmartContract provides the information regarding initial deployed SC
type InitialSmartContract struct {
	Owner          string `json:"owner"`
	Filename       string `json:"filename"`
	VmType         string `json:"vm-type"`
	InitParameters string `json:"init-parameters"`
	Type           string `json:"type"`
	Version        string `json:"version"`
	ownerBytes     []byte
	vmTypeBytes    []byte
	addressesBytes [][]byte
	addresses      []string
}

// OwnerBytes will return the owner's address as raw bytes
func (isc *InitialSmartContract) OwnerBytes() []byte {
	return isc.ownerBytes
}

// SetOwnerBytes will set the owner address as raw bytes
func (isc *InitialSmartContract) SetOwnerBytes(owner []byte) {
	isc.ownerBytes = owner
}

// VmTypeBytes returns the vm type as raw bytes
func (isc *InitialSmartContract) VmTypeBytes() []byte {
	return isc.vmTypeBytes
}

// SetVmTypeBytes sets the vm type as raw bytes
func (isc *InitialSmartContract) SetVmTypeBytes(vmType []byte) {
	isc.vmTypeBytes = vmType
}

// GetOwner returns the smart contract owner address
func (isc *InitialSmartContract) GetOwner() string {
	return isc.Owner
}

// GetFilename returns the filename
func (isc *InitialSmartContract) GetFilename() string {
	return isc.Filename
}

// GetVmType returns the vm type string
func (isc *InitialSmartContract) GetVmType() string {
	return isc.VmType
}

// GetInitParameters returns the init parameters for the smart contract
func (isc *InitialSmartContract) GetInitParameters() string {
	return isc.InitParameters
}

// GetType returns the smart contract's type
func (isc *InitialSmartContract) GetType() string {
	return isc.Type
}

// AddAddressBytes adds a deployed address to the initial smart contract
func (isc *InitialSmartContract) AddAddressBytes(addressBytes []byte) {
	isc.addressesBytes = append(isc.addressesBytes, addressBytes)
}

// AddressesBytes returns the smart contract addresses bytes
func (isc *InitialSmartContract) AddressesBytes() [][]byte {
	return isc.addressesBytes
}

// AddAddress adds a deployed address to the initial smart contract addresses as string
func (isc *InitialSmartContract) AddAddress(address string) {
	isc.addresses = append(isc.addresses, address)
}

// Addresses returns the smart contract addresses string
func (isc *InitialSmartContract) Addresses() []string {
	return isc.addresses
}

// GetVersion returns the recorded version (if existing) of the SC
func (isc *InitialSmartContract) GetVersion() string {
	return isc.Version
}

// IsInterfaceNil returns if underlying object is true
func (isc *InitialSmartContract) IsInterfaceNil() bool {
	return isc == nil
}
