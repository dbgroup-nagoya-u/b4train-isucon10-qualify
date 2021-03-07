#!/bin/bash
set -xe
set -o pipefail

CURRENT_DIR=$(cd $(dirname $0);pwd)
export PG_HOST=${PG_HOST:-127.0.0.1}
export PG_PORT=${PG_PORT:-5432}
export PG_USER=${PG_USER:-isucon}
export PG_DBNAME=${PG_DBNAME:-isuumo}
export LANG="C.UTF-8"
cd ${CURRENT_DIR}

dropdb --if-exists isuumo
createdb isuumo
psql -h ${PG_HOST} -p ${PG_PORT} -U ${PG_USER} -f "0_Schema.sql" -f "1_DummyEstateData.sql" -f "2_DummyChairData.sql" ${PG_DBNAME}
