#!/bin/bash

# clear logs
if [ -f ${NGINX_LOG_PATH} ]; then
  sudo /bin/rm ${NGINX_LOG_PATH}
fi
sudo /usr/bin/touch ${NGINX_LOG_PATH}
sudo /bin/chown www-data:adm ${NGINX_LOG_PATH}
sudo /bin/chmod 640 ${NGINX_LOG_PATH}

# apply new settings
sudo /bin/cp -b ${WORKSPACE}/conf/nginx/nginx.conf ${NGINX_CONF_PATH}
sudo /bin/chmod 644 ${NGINX_CONF_PATH}
sudo /bin/cp -b ${WORKSPACE}/conf/nginx/${SERVICE_NAME}.conf ${NGINX_SITE_PATH}
sudo /bin/chmod 644 ${NGINX_SITE_PATH}

# reload service
sudo /bin/systemctl start nginx.service
sudo /bin/systemctl enable nginx.service
