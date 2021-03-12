#!/bin/bash
set -ue -o pipefail

CURRENT_DIR=$(cd $(dirname $0);pwd)
source ${HOME}/env.sh
export LANG="C.UTF-8"

cd ${CURRENT_DIR}

if ! psql -q -h ${PG_HOST} -p ${PG_PORT} -U ${PG_USER} -d ${PG_DBNAME} -c "\q" > /dev/null 2>&1; then
  createdb ${PG_DBNAME}
fi
psql -h ${PG_HOST} -p ${PG_PORT} -U ${PG_USER} -d ${PG_DBNAME} \
  -f "0_Schema.sql" \
  -f "1_DummyEstateData.sql" \
  -f "2_DummyChairData.sql"
