#!/bin/bash

# set global constants
source ${HOME}/env.sh

CUR_DIR=$(cd $(dirname ${BASH_SOURCE:-${0}}); pwd)

${CUR_DIR}/disable_app.sh
${CUR_DIR}/disable_nginx.sh
${CUR_DIR}/disable_postgresql.sh
