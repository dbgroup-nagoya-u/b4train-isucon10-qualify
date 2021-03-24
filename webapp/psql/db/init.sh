#!/bin/bash
set -ue -o pipefail

CURRENT_DIR=$(cd $(dirname $0);pwd)
source ${HOME}/env.sh
export LANG="C.UTF-8"

cd ${CURRENT_DIR}

if ! psql -q -h ${DB_HOST} -p ${PGPORT} -U ${PGUSER} -d ${DB_NAME} -c "\q" > /dev/null 2>&1; then
  createdb ${DB_NAME}
fi
psql -h ${DB_HOST} -p ${PGPORT} -U ${PGUSER} -d ${DB_NAME} \
  -f "0_Schema.sql" \
  -f "1_DummyEstateData.sql" \
  -f "2_DummyChairData.sql" \
  -f "3_PostInitialization.sql"
