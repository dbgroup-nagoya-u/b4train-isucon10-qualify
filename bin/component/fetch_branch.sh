#!/bin/bash

# set global constants
source ${HOME}/env.sh

# set target branch
GIT_BRANCH=${1}

# prepare source codes
cd ${WORKSPACE}
git fetch --quiet origin
git checkout --quiet "${GIT_BRANCH}"
git merge --quiet "origin/${GIT_BRANCH}"
