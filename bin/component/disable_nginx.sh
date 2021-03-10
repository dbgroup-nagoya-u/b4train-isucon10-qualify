#!/bin/bash

# set global constants
source ${HOME}/env.sh

sudo /bin/systemctl --quiet stop nginx.service
sudo /bin/systemctl --quiet disable nginx.service
