#!/bin/bash

clickhouse-client -h$1 -u$2 --password --query="$(<(pwd)/migrations/all.sql)"
# docker run --rm --network="clickhouse-net" --link clickhouse-$1:clickhouse-server yandex/clickhouse-client -m --host clickhouse-server --query="$distributed"