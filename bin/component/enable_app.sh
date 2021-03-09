#!/bin/bash

# set global constants
source ${HOME}/env.sh

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
sudo /bin/systemctl --quiet daemon-reload

# reload service
sudo /bin/systemctl --quiet start ${SERVICE_FILE}
sudo /bin/systemctl --quiet enable ${SERVICE_FILE}
