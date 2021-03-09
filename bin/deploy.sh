#!/bin/bash
set -ue -o pipefail

# set global constants
source ${HOME}/env.sh

# run scripts at workspace
cd ${WORKSPACE}

####################################################################################################
# Documents
####################################################################################################

usage() {
  cat 1>&2 << EOS
Usage:
  ${BASH_SOURCE:-${0}} <branch_name>
Description:
  Pull and deploy source codes to workers.
Arguments:
  <branch_name>: A branch name to deploy source codes.
EOS
  exit 1
}

####################################################################################################
# Check command-line arguments
####################################################################################################

# check whether there is input arguments
if [ ${#} -ne 1 ]; then
  usage
fi

# check whether there is a specified branch
readonly GIT_BRANCH=${1}
git fetch --quiet origin
if ! git branch --list "${GIT_BRANCH}" | grep "${GIT_BRANCH}" &> /dev/null; then
  echo "There is no branch: ${GIT_BRANCH}" 1>&2
  exit 1
fi

####################################################################################################
# Deploy app
####################################################################################################

# sync local sources with remote ones
git checkout --quiet "${GIT_BRANCH}"
git merge --quiet "origin/${GIT_BRANCH}"

# load environment variables
echo "Load new environment variables..."
source ${WORKSPACE}/conf/env.sh

# initialization
echo "Initialize a worker [${WORKERS}]..."
for WORKER in ${WORKERS}; do
  echo "${WORKER}:"
  echo "  fetch remote sources..."
  ssh ${WORKER} ${WORKSPACE}/bin/component/fetch_branch.sh "${GIT_BRANCH}"
  echo "  disable all services..."
  ssh ${WORKER} ${WORKSPACE}/bin/component/disable_all.sh
  echo "  update server environment..."
  ssh ${WORKER} ${WORKSPACE}/bin/component/environment.sh
  echo "${WORKER} done."
done

# start/enable DB daemon
for DB_HOST in ${DB_HOSTS}; do
  echo "start database service on ${DB_HOST}"
  ssh ${DB_HOST} ${WORKSPACE}/bin/component/enable_postgresql.sh
done

# start/enable app daemon
for APP_HOST in ${APP_HOSTS}; do
  echo "start app service on ${APP_HOST}"
  ssh ${APP_HOST} "export PATH=$PATH; ${WORKSPACE}/bin/component/enable_app.sh"
done

# start/enable Web daemon
echo "start web service on ${WEB_HOST}"
ssh ${WEB_HOST} ${WORKSPACE}/bin/component/enable_nginx.sh

echo "done."
