@0x9d2d27546a44231a;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");

struct BranchNodeCapn {
    encodedChildren @0: List(Data);
}

struct ExtensionNodeCapn {
    key          @0: Data;
    encodedChild @1: Data;
}

struct LeafNodeCapn {
    key     @0: Data;
    value   @1: Data;
}

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/capnp/schema.capnp
