namespace go rpc

// Metadata field of the request in rDSN's thrift protocol (version 1).
struct thrift_request_meta_v1
{
    // The replica's gpid.
    1:optional i32 app_id;
    2:optional i32 partition_index;

    // The timeout of this request that's set on client side.
    3:optional i32 client_timeout;

    // The hash value calculated from the hash key.
    4:optional i64 client_partition_hash;

    // Whether it is a backup request. If true, this request (only if it's a read) can be handled by
    // a secondary replica, which does not guarantee strong consistency.
    5:optional bool is_backup_request;
}
