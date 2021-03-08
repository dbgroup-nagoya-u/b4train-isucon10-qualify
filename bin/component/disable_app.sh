#!/bin/bash

sudo /bin/systemctl stop ${SERVICE_FILE}
sudo /bin/systemctl disable ${SERVICE_FILE}
