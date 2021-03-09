#!/bin/bash

# set global constants
source ${HOME}/env.sh

sudo /bin/systemctl --quiet stop postgresql.service
sudo /bin/systemctl --quiet disable postgresql.service
