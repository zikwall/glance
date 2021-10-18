-- this migration use in external service
CREATE TABLE stream.heatmap ON CLUSTER cluster_1
(
    `uniqid`      String,
    `device_id`   String,
    `platform`    String,
    `app`         String,
    `version`     String,
    `os`          String,
    `browser`     String,
    `country`     String,
    `region`      String,
    `insert_ts`   DateTime,
    `insert_date` Date
)
    ENGINE = Distributed('cluster_1', 'stream', 'heatmap_sharded', rand());

CREATE TABLE stream.heatmap_sharded ON CLUSTER cluster_1
(
    `uniqid`      String,
    `device_id`   String,
    `platform`    String,
    `app`         String,
    `version`     String,
    `os`          String,
    `browser`     String,
    `country`     String,
    `region`      String,
    `insert_ts`   DateTime,
    `insert_date` Date
)
    ENGINE = ReplicatedMergeTree('/clickhouse/tables/stream/{shard}/heatmap_sharded', '{replica}')
    PARTITION BY toYYYYMM(insert_date)
    ORDER BY (platform, insert_date)
    TTL insert_ts + INTERVAL 13 MONTH;