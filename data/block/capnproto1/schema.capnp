@0xb9f45775755d8a42;
using Go = import "/go.capnp";
$Go.package("capnproto1");
$Go.import("_");


struct HeaderCapn { 
   nonce          @0:   UInt64; 
   prevHash       @1:   Data; 
   pubKeysBitmap  @2:   List(Bool); 
   shardId        @3:   UInt32; 
   timeStamp      @4:   Data; 
   round          @5:   UInt32; 
   blockBodyHash  @6:   Data; 
   signature      @7:   Data; 
   commitment     @8:   Data; 
} 

struct MiniBlockCapn { 
   txHashes     @0:   List(Data); 
   destShardID  @1:   UInt32; 
} 

struct PeerBlockBodyCapn { 
   stateBlockBody  @0:   StateBlockBodyCapn; 
   changes         @1:   List(PeerChangeCapn); 
} 

struct PeerChangeCapn { 
   pubKey       @0:   Data; 
   shardIdDest  @1:   UInt32; 
} 

struct StateBlockBodyCapn { 
   rootHash  @0:   Data; 
} 

struct TxBlockBodyCapn { 
   stateBlockBody  @0:   StateBlockBodyCapn; 
   miniBlocks      @1:   List(MiniBlockCapn); 
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go-sandbox/data/block/capnproto1/schema.capnp

