@0xec208700833b2f3f;
using Go = import "/go.capnp";
$Go.package("capnproto2");
$Go.import("_");

struct BlockCapnp $Go.doc("Block of data, containing hashes of transaction"){
    miniBlocks @0:List(MiniBlock);
    struct MiniBlock {
        txHashes @0:List(Data);
        destShardID @1:UInt32;
    }
}