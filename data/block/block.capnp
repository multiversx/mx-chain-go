@0xec208700833b2f3f;
using Go = import "/go.capnp";
$Go.package("block");
$Go.import("_");

struct Block $Go.doc("Block of data, containing hashes of transaction"){
    miniBlocks @0:List(MiniBlock);
    struct MiniBlock {
        txHashes @0:List(UInt8);
        destShardID @1:UInt32;
    }
}