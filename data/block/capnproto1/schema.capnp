@0xc2691bfa765feae5;
using Go = import "/go.capnp";
$Go.package("capnproto1");
$Go.import("_");


struct BlockCapn { 
   miniBlocks  @0:   List(MiniBlockCapn); 
} 

struct HeaderCapn { 
   nonce       @0:   Data;
   prevHash    @1:   Data;
   pubKeys     @2:   List(Data);
   shardId     @3:   UInt32; 
   timeStamp   @4:   Data;
   round       @5:   UInt32; 
   blockHash   @6:   Data;
   signature   @7:   Data;
   commitment  @8:   Data;
} 

struct MiniBlockCapn { 
   txHashes     @0:   List(Data);
   destShardID  @1:   UInt32; 
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnproto1//schema.capnp

