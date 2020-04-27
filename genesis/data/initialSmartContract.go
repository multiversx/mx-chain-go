package data

// InitialSmartContract provides the information regarding initial deployed SC
type InitialSmartContract struct {
	Owner      string `json:"owner"`
	Filename   string `json:"filename"`
	ownerBytes []byte
	hexCode    string
}

// OwnerBytes will return the owner's address as raw bytes
func (isc *InitialSmartContract) OwnerBytes() []byte {
	return isc.ownerBytes
}

// SetOwnerBytes will set the owner address as raw bytes
func (isc *InitialSmartContract) SetOwnerBytes(owner []byte) {
	isc.ownerBytes = owner
}

// SetCode will set the SC code as hex string
func (isc *InitialSmartContract) SetCode(hexCode string) {
	isc.hexCode = hexCode
}

// GetCode will return the SC code as hex string
func (isc *InitialSmartContract) GetCode() string {
	return isc.hexCode
}

// IsInterfaceNil returns if underlying object is true
func (isc *InitialSmartContract) IsInterfaceNil() bool {
	return isc == nil
}
