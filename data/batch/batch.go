//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. batch.proto
package batch

func New(buffs ...[]byte) *Batch {
	return &Batch{
		Data: buffs,
	}
}
