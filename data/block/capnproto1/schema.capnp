@0xc2691bfa765feae5;
using Go = import "/go.capnp";
$Go.package("capnproto1");
$Go.import("_");


struct BlockCapn { 
   miniBlocks  @0:   List(MiniBlockCapn); 
} 

struct HeaderCapn { 
   nonce       @0:   List(UInt8); 
   prevHash    @1:   List(UInt8); 
   pubKeys     @2:   List(List(UInt8)); 
   shardId     @3:   UInt32; 
   timeStamp   @4:   List(UInt8); 
   round       @5:   UInt32; 
   blockHash   @6:   List(UInt8); 
   signature   @7:   List(UInt8); 
   commitment  @8:   List(UInt8); 
} 

struct MiniBlockCapn { 
   txHashes     @0:   List(List(UInt8)); 
   destShardID  @1:   UInt32; 
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnproto1//schema.capnp

