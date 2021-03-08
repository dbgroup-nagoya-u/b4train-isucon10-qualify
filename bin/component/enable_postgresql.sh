#!/bin/bash

# clear logs
if [ -f ${PG_LOG_PATH} ]; then
  sudo /bin/rm ${PG_LOG_PATH}
fi
sudo /usr/bin/touch ${PG_LOG_PATH}
sudo /bin/chown postgres:adm ${PG_LOG_PATH}
sudo /bin/chmod 640 ${PG_LOG_PATH}

# apply new authorization settings
cp -b ${WORKSPACE}/conf/postgresql/pgpass ${HOME}/.pgpass
chmod 600 ${HOME}/.pgpass

# apply new DB settings
sudo /bin/cp -b ${WORKSPACE}/conf/postgresql/override.conf ${PG_OVERRIDE_PATH}
sudo /bin/chown postgres:postgres ${PG_OVERRIDE_PATH}
sudo /bin/chmod 644 ${PG_OVERRIDE_PATH}
sudo /bin/cp -b ${WORKSPACE}/conf/postgresql/pg_hba.conf ${PG_HBA_PATH}
sudo /bin/chown postgres:postgres ${PG_HBA_PATH}
sudo /bin/chmod 640 ${PG_HBA_PATH}

# reload service
sudo /bin/systemctl start postgresql.service
sudo /bin/systemctl enable postgresql.service
