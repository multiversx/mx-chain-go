@0x9096faa5d587481d;
using Go = import "/go.capnp";
$Go.package("capnproto2");
$Go.import("_");

struct HeaderCapnp $Go.doc("header information, attachable to blocks or miniblocks"){
    nonce @0 :UInt64;
    prevHash @1:Data;
    # temporary keep list of public keys of signers in header
    # to be removed later
    pubKeys @2:List(Data);
    shardId @3:UInt32;
    timeStamp @4:Data;
    round @5:UInt32;
    blockHash @6: Data;
    signature @7: Data;
    commitment @8: Data;
}
