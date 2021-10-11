CREATE DATABASE stream ON CLUSTER cluster_1;

CREATE TABLE stream.http_status ON CLUSTER cluster_1
(
    `stream_id`   String,
    `code`        Int8,
    `insert_ts`   DateTime,
    `insert_date` Date
)
    ENGINE = Distributed('cluster_1', 'stream', 'http_status_sharded', rand());

CREATE TABLE stream.http_status_sharded ON CLUSTER cluster_1
(
    `stream_id`   String,
    `code`        Int8,
    `insert_ts`   DateTime,
    `insert_date` Date
)
    ENGINE = ReplicatedMergeTree('/clickhouse/tables/stream/{shard}/http_status_sharded', '{replica}')
        PARTITION BY toYYYYMM(insert_date)
        ORDER BY (stream_id, insert_date)
        TTL insert_ts + INTERVAL 12 MONTH;