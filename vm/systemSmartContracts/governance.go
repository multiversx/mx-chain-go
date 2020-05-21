//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. governance.proto
package systemSmartContracts

// ArgsNewGovernanceContract defines the arguments needed for the on-chain governance contract
type ArgsNewGovernanceContract struct {
}

type governanceContract struct {
}
