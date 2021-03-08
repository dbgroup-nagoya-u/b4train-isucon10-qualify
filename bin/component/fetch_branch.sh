#!/bin/bash

# prepare source codes
cd ${WORKSPACE}
git fetch origin
git checkout --quiet "${GIT_BRANCH}"
git merge --quiet "origin/${GIT_BRANCH}"
