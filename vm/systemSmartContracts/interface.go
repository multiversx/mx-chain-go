package systemSmartContracts

type TokenIdentifierCreatorHandler interface {
	CreateNewTokenIdentifier(caller []byte, ticker []byte) ([]byte, error)
	IsInterfaceNil() bool
}
