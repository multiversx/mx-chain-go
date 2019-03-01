@0xb9f45775755d8a42;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");


struct HeaderCapn {
   nonce            @0:   UInt64;
   prevHash         @1:   Data;
   pubKeysBitmap    @2:   Data;
   shardId          @3:   UInt32;
   timeStamp        @4:   UInt64;
   round            @5:   UInt32;
   epoch            @6:   UInt32;
   blockBodyType    @7:   UInt8;
   signature        @8:   Data;
   commitment       @9:   Data;
   miniBlockHeaders @10:  List(MiniBlockHeaderCapn);
   peerChanges      @11:  List(Data);
   rootHash         @12:  Data;
}

struct MiniBlockHeaderCapn {
   hash            @0: Data;
   receiverShardID @1: UInt32;
   senderShardID   @2: UInt32;
}

struct MiniBlockCapn {
   txHashes     @0:   List(Data);
   shardID      @1:   UInt32;
}

struct PeerChangeCapn {
   pubKey       @0:   Data;
   shardIdDest  @1:   UInt32;
}

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnp/schema.capnp
