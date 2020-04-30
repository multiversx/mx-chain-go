package data

// InitialSmartContract provides the information regarding initial deployed SC
type InitialSmartContract struct {
	Owner      string `json:"owner"`
	Filename   string `json:"filename"`
	VmType     string `json:"vm-type"`
	ownerBytes []byte
}

// OwnerBytes will return the owner's address as raw bytes
func (isc *InitialSmartContract) OwnerBytes() []byte {
	return isc.ownerBytes
}

// SetOwnerBytes will set the owner address as raw bytes
func (isc *InitialSmartContract) SetOwnerBytes(owner []byte) {
	isc.ownerBytes = owner
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

// IsInterfaceNil returns if underlying object is true
func (isc *InitialSmartContract) IsInterfaceNil() bool {
	return isc == nil
}
