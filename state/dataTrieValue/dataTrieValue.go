//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf --gogoslick_out=. dataTrieValue.proto
package dataTrieValue
