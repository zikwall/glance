CREATE DATABASE stream ON CLUSTER cluster_1;

CREATE TABLE stream.metrics ON CLUSTER cluster_1
(
    `stream_id`         String,
    `bitrate`           Float64,
    `frames`            UInt64,
    `height`            UInt64,
    `fps`               Float64,
    `bytes`             UInt64,
    `seconds`           Float64,
    `keyframe_interval` UInt64,
    `insert_ts`         DateTime,
    `date`              Date
)
    ENGINE = Distributed('cluster_1', 'stream', 'metrics_sharded', rand());

CREATE TABLE stream.metrics_sharded ON CLUSTER cluster_1
(
    `stream_id`         String,
    `bitrate`           Float64,
    `frames`            UInt64,
    `height`            UInt64,
    `fps`               Float64,
    `bytes`             UInt64,
    `seconds`           Float64,
    `keyframe_interval` UInt64,
    `insert_ts`         DateTime,
    `date`              Date
)
    ENGINE = ReplicatedMergeTree('/clickhouse/tables/stream/{shard}/metrics_sharded', '{replica}')
        PARTITION BY toYYYYMM(date)
        ORDER BY (stream_id, date)
        TTL insert_ts + INTERVAL 12 MONTH;