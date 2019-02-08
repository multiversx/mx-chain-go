@0xddffc3d7f7f36183;
using Go = import "/go.capnp";
$Go.package("capnproto1");
$Go.import("_");

struct PeerDataCapn {
    publicKey @0: Data;
    action    @1: UInt8;
    timestamp @2: UInt64;
    value     @3: Data;
}

struct ShardDataCapn {
    shardId      @0: UInt32;
    headerHashes @1: List(Data);
}

struct ProofCapn {
    inclusionProof @0: Data;
    exclusionProof @1: Data;
}

struct MetaBlockCapn {
    nonce     @0: UInt64;
    epoch     @1: UInt32;
    round     @2: UInt32;
    shardInfo @3: List(ShardDataCapn);
    peerInfo  @4: List(PeerDataCapn);
    proof     @5: ProofCapn;
}

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnp/schema.metablock.capnp

