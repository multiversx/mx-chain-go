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
//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. message.proto
```

And any necessary methods that are not generated.

#### Generating code ####

Check [gogo protobuf geting started](https://github.com/gogo/protobuf#getting-started) for initial setup.

**Note:** Currently we are using a not yet merged gogo/protobuf extension, [casttypewith](https://github.com/gogo/protobuf/pull/659), in order to use it before it gets merged you should:


```
cd $(go env GOPATH)/src/github.com/gogo/protobuf
git fetch origin pull/659/head:casttypewith && git checkout casttypewith
GOPATH=$(go env GOPATH) make install

```

To regenerate all the data structures run:
```
go generate ./...
```
in the project root.

