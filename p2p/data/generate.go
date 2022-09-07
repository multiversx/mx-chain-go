//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. topicMessage.proto
package data
