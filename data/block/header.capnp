@0x9096faa5d587481d;
using Go = import "/go.capnp";
$Go.package("block");
$Go.import("_");

struct Header $Go.doc("header information, attachable to blocks or miniblocks"){
    nonce @0 :List(UInt8);
    prevHash @1:List(UInt8);
    # temporary keep list of public keys of signers in header
    # to be removed later
    pubKeys @2:List(List(UInt8));
    shardId @3:UInt32;
    timeStamp @4:List(UInt8);
    round @5:UInt32;
    blockHash @6: List(UInt8);
    signature @7: List(UInt8);
    commitment @8: List(UInt8);
}