#!/bin/bash
set -e

clickhouse client -n <<-EOSQL
    CREATE DATABASE IF NOT EXISTS sayGames;
    CREATE TABLE IF NOT EXISTS sayGames.logs (
      clientTime String,
      deviceId String,
      deviceOs String,
      session String,
      sequence Int32,
      event String,
      paramInt Int32,
      paramStr String,
      ip String,
			serverTime String
    ) ENGINE = Log;
EOSQL