package storage

// StoredDataFactoryStub -
type StoredDataFactoryStub struct {
	CreateEmptyCalled func() interface{}
}

// CreateEmpty -
func (sdf *StoredDataFactoryStub) CreateEmpty() interface{} {
	if sdf.CreateEmptyCalled != nil {
		return sdf.CreateEmptyCalled()
	}

	return nil
}

// IsInterfaceNil -
func (sdf *StoredDataFactoryStub) IsInterfaceNil() bool {
	return sdf == nil
}
