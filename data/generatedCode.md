### Data structure definition ###

The usual message definition consists of:

```
message/
├── proto
│   └── message.proto
├── message.go
└── message.pb.go [generated]
```

#### message.proto ####

Contains the protobuf message definition (with gogo/protobuf extensions)

#### message.go ####

Should contain the go generate directive

```
//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. message.proto
```

And any necessary methods that are not generated.

#### Generating code ####

Check [gogo protobuf geting started](https://github.com/ElrondNetwork/protobuf#getting-started) for initial setup.

To regenerate all the data structures run:
```
go generate ./...
```
in the project root.

