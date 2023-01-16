//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. validatorInfo.proto

package state

// IsInterfaceNil returns true if there is no value under the interface
func (vi *ValidatorInfo) IsInterfaceNil() bool {
	return vi == nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (svi *ShardValidatorInfo) IsInterfaceNil() bool {
	return svi == nil
}
