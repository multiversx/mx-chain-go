//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. heartbeat.proto
package heartbeat

//TODO clean after the transition phase:
//1. remove elrond-go/heartbeat/data, elrond-go/heartbeat/process, elrond-go/heartbeat/storage packages
//2. remove the old heartbeat instantiation mechanism
