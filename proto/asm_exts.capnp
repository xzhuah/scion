@0xe6c88f91b6a1209e;
using Go = import "go.capnp";
$Go.package("proto");
$Go.import("github.com/scionproto/scion/go/proto");

struct RoutingPolicyExt{
    set @0 :Bool;   # Is the extension present? Every extension must include this field.
    polType @1 :UInt8;  # The policy type
    ifID @2 :UInt64;
    isdases @3 :List(UInt64);
}

struct ISDAnnouncementExt{
    set @0 :Bool;   # TODO(Sezer): Implement announcement extension
}

struct BwCluster {
    clusterBW @0 :UInt32;
    interfaces @1 :List(UInt16);
}

struct BwInfo {
    bwClusters @0 :List(BwCluster);
    egressBW @1 :UInt32;
    inToOutBW @2 :UInt32;
}

struct GeoInfo {
    latitude @0 :Float64;
    longitude @1 :Float64;
}

struct WatchDogMetricExt{
    set @0 :Bool;
    bwInfo @1 :BwInfo;
    geoInfo @2 :GeoInfo;
}
