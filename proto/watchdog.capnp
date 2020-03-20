@0xe6c88aaab6a1209e;
using Go = import "/go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct WatchDogMetricExt {
    set @0 :Bool;   # Is the extension present? Every extension must include this field.
    val @1 :UInt32;
}
