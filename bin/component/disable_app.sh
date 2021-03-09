#!/bin/bash

# set global constants
source ${HOME}/env.sh

sudo /bin/systemctl --quiet stop ${SERVICE_FILE}
sudo /bin/systemctl --quiet disable ${SERVICE_FILE}
