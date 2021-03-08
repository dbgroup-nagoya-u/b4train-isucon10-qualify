#!/bin/bash

# copy new environment variables
cp -b ${WORKSPACE}/conf/env.sh ${HOME}/env.sh

# reload sysctl
sudo /bin/cp -b ${WORKSPACE}/conf/sysctl.conf /etc/sysctl.conf
sudo /sbin/sysctl -p --quiet
