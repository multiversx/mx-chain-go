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
	addressBytes   []byte
	address        string
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

// SetAddressBytes sets the initial smart contract address bytes
func (isc *InitialSmartContract) SetAddressBytes(addressBytes []byte) {
	isc.addressBytes = addressBytes
}

// AddressBytes returns the smart contract address bytes
func (isc *InitialSmartContract) AddressBytes() []byte {
	return isc.addressBytes
}

// SetAddress sets the initial smart contract address as string
func (isc *InitialSmartContract) SetAddress(address string) {
	isc.address = address
}

// Address returns the smart contract address string
func (isc *InitialSmartContract) Address() string {
	return isc.address
}

// GetVersion returns the recorded version (if existing) of the SC
func (isc *InitialSmartContract) GetVersion() string {
	return isc.Version
}

// IsInterfaceNil returns if underlying object is true
func (isc *InitialSmartContract) IsInterfaceNil() bool {
	return isc == nil
}
