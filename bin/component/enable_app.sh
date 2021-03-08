#!/bin/bash

# constants
readonly APP_OUT_LOG_PATH="${APP_DIR}/log/app.out"
readonly APP_ERR_LOG_PATH="${APP_DIR}/log/app.err"
readonly APP_SERVICE_PATH="/etc/systemd/system/${SERVICE_FILE}"

# compile app with new sources
cd ${APP_DIR}
make --quiet

# clear logs
if [ -f ${APP_OUT_LOG_PATH} ]; then
  rm ${APP_OUT_LOG_PATH}
fi
touch ${APP_OUT_LOG_PATH}
if [ -f ${APP_ERR_LOG_PATH} ]; then
  rm ${APP_ERR_LOG_PATH}
fi
touch ${APP_ERR_LOG_PATH}

# apply new settings
sudo /bin/cp -b ${WORKSPACE}/conf/app/${SERVICE_FILE} ${APP_SERVICE_PATH}
sudo /bin/systemctl daemon-reload

# reload service
sudo /bin/systemctl start ${SERVICE_FILE}
sudo /bin/systemctl enable ${SERVICE_FILE}
