CREATE TABLE logs
(
    service_id  LowCardinality(String),
    occurred_at DateTime64(3, 'UTC'),
    received_at DateTime64(3, 'UTC') DEFAULT now64(3),
    message     String,
    level       Enum8('DEBUG' = 1, 'INFO' = 2, 'WARN' = 3, 'ERROR' = 4),
    metadata Map(String, String)
)
ENGINE = MergeTree
PARTITION BY toDate(received_at)
ORDER BY (service_id, received_at);
