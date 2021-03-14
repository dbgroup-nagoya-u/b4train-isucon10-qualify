#!/bin/bash

# set global constants
source ${HOME}/env.sh

# copy new environment variables
cp -b ${WORKSPACE}/conf/env.sh ${HOME}/env.sh

# copy new host information
sudo cp -b ${WORKSPACE}/conf/hosts /etc/hosts

# apply new DB authorization settings
cp -b ${WORKSPACE}/conf/postgresql/pgpass ${HOME}/.pgpass
chmod 600 ${HOME}/.pgpass

# reload sysctl
sudo /bin/cp -b ${WORKSPACE}/conf/sysctl.conf /etc/sysctl.conf
sudo /sbin/sysctl -p --quiet
